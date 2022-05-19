package lockserver

import (
	"fmt"
	"sync"
)

type LockManager struct {
	metaMu sync.Mutex
	readLockStatus  map[string][]string
	// key -> list of array clientIds
	writeLocksStatus map[string]string
	// key -> []LockRequest
	requestsQueueMap map[string][]LockRequest
	UnimplementedLockServiceServer

	waitersLock sync.Mutex
	waiters map[string]chan bool

	clientLocksMu sync.Mutex
	clientReadLocks map[string][]string
	clientWriteLocks map[string][]string

	clientLeaseMu sync.Mutex
	clientLease map[string]int64
}

func NewLockManager() *LockManager {
	return &LockManager{
		readLockStatus: make(map[string][]string),
		writeLocksStatus: make(map[string]string),
		requestsQueueMap: make(map[string][]LockRequest),
		waiters: make(map[string]chan bool),
	}
}

func (lm *LockManager) processAcquireRequest(request LockRequest) {
	key := request.key
	clientId := request.clientId
	var goodToGoPool []string

	lm.metaMu.Lock()

	lm.requestsQueueMap[key] = append(lm.requestsQueueMap[key], request)
	curReaders, readOwnedByOtherClient := lm.readLockStatus[key]
	_, writeOwnedByOtherClient := lm.writeLocksStatus[key]

	if readOwnedByOtherClient && writeOwnedByOtherClient {
		panicStr, _ := fmt.Printf("server has both read and write at one time for key %s\n", key)
		panic(panicStr)
	}
	if readOwnedByOtherClient && !contains(curReaders, clientId) {
		goodToGoPool = lm.openRequestQueueValve(key)
		lm.notifyClientsToProceed(goodToGoPool)
	}
	lm.metaMu.Unlock()
}

func (lm *LockManager) processReleaseRequest(request LockRequest) {
	key := request.key
	clientId := request.clientId
	var goodToGoPool []string

	lm.metaMu.Lock()

	curReaders, readOwnedByOtherClient := lm.readLockStatus[key]
	curWriter, writeOwnedByOtherClient := lm.writeLocksStatus[key]

	if readOwnedByOtherClient && writeOwnedByOtherClient {
		panicStr, _ := fmt.Printf("server has both read and write at one time for key %s\n", key)
		panic(panicStr)
	}
	if request.rwflag == READ {
		if !readOwnedByOtherClient || (readOwnedByOtherClient && !contains(curReaders, clientId)) {
			panicStr, _ := fmt.Printf("server tends to release a read key that client does not own %s by %s\n", key, clientId)
			panic(panicStr)
		}
		lm.readLockStatus[key] = remove(lm.readLockStatus[key], clientId)
		// update the clientLocks set
		lm.clientLocksMu.Lock()
		lm.clientReadLocks[clientId] = remove(lm.clientReadLocks[clientId], key)
		lm.clientLocksMu.Unlock()
		if len(lm.readLockStatus[key]) == 0 {
			delete(lm.readLockStatus, key)
		}
	} else if request.rwflag == WRITE {
		if !writeOwnedByOtherClient || (writeOwnedByOtherClient && curWriter != clientId) {
			panicStr, _ := fmt.Printf("server tends to release a write key that client does not own %s by %s\n", key, clientId)
			panic(panicStr)
		}
		delete(lm.writeLocksStatus, key)
		// update the clientLocks set
		lm.clientLocksMu.Lock()
		lm.clientWriteLocks[clientId] = remove(lm.clientWriteLocks[clientId], key)
		lm.clientLocksMu.Unlock()
	}
	goodToGoPool = lm.openRequestQueueValve(key)
	lm.notifyClientsToProceed(goodToGoPool)

	lm.metaMu.Unlock()
}

func (lm *LockManager) openRequestQueueValve(key string) []string {
	// key: the key we're allowing clients to acquire
	// retVal: goodToGoPool is for new clients that successfully acquire the lock, we need to signal them
	// nobody's owned the lock
	// readLockStatusLock locked
	// writeLocksStatusLock locked
	// requestQueueMapLock locked
	goodToGoPool := make([]string, 0)
	var headRequest LockRequest
	// first check whether nobody is using the key at all
	_, readOwnedByOtherClient := lm.readLockStatus[key]
	_, writeOwnedByOtherClient := lm.writeLocksStatus[key]

	if readOwnedByOtherClient && writeOwnedByOtherClient {
		panicStr, _ := fmt.Printf("server has both read and write at one time for key %s\n", key)
		panic(panicStr)
	}
	if writeOwnedByOtherClient {
		// we cannot let any one in because somebody is holding the exclusive lock
		return goodToGoPool
	}

	// rwValve: nobody's holding any rwlock of the key, so we're allowed to make write acquire coming in also
	rwValve := !readOwnedByOtherClient && !writeOwnedByOtherClient
	if rwValve && len(lm.requestsQueueMap[key]) > 0 && lm.requestsQueueMap[key][0].rwflag == WRITE {
		// only let the next write out
		headRequest, lm.requestsQueueMap[key] = lm.requestsQueueMap[key][0], lm.requestsQueueMap[key][1:]
		// update the writeLocksStatus
		lm.writeLocksStatus[key] = headRequest.clientId
		// update the clientLocks set
		lm.clientLocksMu.Lock()
		lm.clientWriteLocks[headRequest.clientId] = append(lm.clientWriteLocks[headRequest.clientId], key)
		lm.clientLocksMu.Unlock()
		// update goodToGoPool
		goodToGoPool = append(goodToGoPool, headRequest.clientId)
		return goodToGoPool
	}
	// last possibility:
	// readOwnedByOtherClient && !writeOwnedByOtherClient
	// we're letting more readers coming in
	for len(lm.requestsQueueMap[key]) > 0 && lm.requestsQueueMap[key][0].rwflag != WRITE {
		// only let the next reads out
		headRequest, lm.requestsQueueMap[key] = lm.requestsQueueMap[key][0], lm.requestsQueueMap[key][1:]
		// update the readLockStatus
		if !contains(lm.readLockStatus[key], headRequest.clientId) {
			lm.readLockStatus[key] = append(lm.readLockStatus[key], headRequest.clientId)
			// update the clientLocks set
			lm.clientLocksMu.Lock()
			lm.clientReadLocks[headRequest.clientId] = append(lm.clientReadLocks[headRequest.clientId], key)
			lm.clientLocksMu.Unlock()
		}
		// update goodToGoPool
		goodToGoPool = append(goodToGoPool, headRequest.clientId)
	}
	return goodToGoPool
}

func (lm *LockManager) notifyClientsToProceed(goodToGoPool []string) {
	// notify corresponding clients to proceed
	for _, clientToNotify := range goodToGoPool {
		go func(clientId string) {
			lm.waitersLock.Lock()
			waiter, ok := lm.waiters[clientId]
			if !ok {
				waiter = make(chan bool)
				lm.waiters[clientId] = waiter
			}
			lm.waitersLock.Unlock()
			waiter <- true
		}(clientToNotify)
	}
}
