package lockserver

import (
	"fmt"
	"sharded_lock_service/pkg/utils"
	"sync"
)

type LockManager struct {
	keysLock *utils.KeysLock
	readLockStatus  *utils.ConcurrentStringSliceMap
	// key -> list of array clientIds
	writeLocksStatus *utils.ConcurrentStringMap
	// key -> []LockRequest
	requestsQueueMap *utils.ConcurrentLockRequestSliceMap
	UnimplementedLockServiceServer

	waitersLock sync.Mutex
	waiters map[string]chan bool

	clientsIdLock *utils.KeysLock
	clientReadLocks *utils.ConcurrentStringSliceMap
	clientWriteLocks *utils.ConcurrentStringSliceMap

	clientLeaseMu sync.Mutex
	clientLease map[string]int64
}

func NewLockManager() *LockManager {
	return &LockManager{
		readLockStatus: utils.NewConcurrentStringSliceMap(),
		writeLocksStatus: utils.NewConcurrentStringMap(),
		requestsQueueMap: utils.NewConcurrentLockRequestSliceMap(),
		keysLock: utils.NewKeysLock(),
		clientReadLocks: utils.NewConcurrentStringSliceMap(),
		clientWriteLocks: utils.NewConcurrentStringSliceMap(),
		clientLease: make(map[string]int64),
		waiters: make(map[string]chan bool),
	}
}

func (lm *LockManager) processAcquireRequest(request LockRequest) {
	key := request.key
	clientId := request.clientId
	var goodToGoPool []string

	lm.keysLock.Lock(key)

	lm.requestsQueueMap.Append(key, request)
	readOwned := lm.readLockStatus.Exists(key)
	writeOwned := lm.writeLocksStatus.Exists(key)

	if readOwned && writeOwned {
		panicStr, _ := fmt.Printf("server has both read and write at one time for key %s\n", key)
		panic(panicStr)
	}
	nobodyHasLock := !readOwned && !writeOwned
	somebodyNotMeHasReadLock := lm.readLockStatus.Contains(key, clientId)
	if nobodyHasLock || somebodyNotMeHasReadLock {
		goodToGoPool = lm.openRequestQueueValve(key)
		lm.notifyClientsToProceed(goodToGoPool)
	}

	lm.keysLock.Unlock(key)
}

func (lm *LockManager) processReleaseRequest(request LockRequest) {
	key := request.key
	clientId := request.clientId
	var goodToGoPool []string

	lm.keysLock.Lock(key)

	readOwned := lm.readLockStatus.Exists(key)
	writeOwned := lm.writeLocksStatus.Exists(key)

	if readOwned && writeOwned {
		panicStr, _ := fmt.Printf("server has both read and write at one time for key %s\n", key)
		panic(panicStr)
	}
	if request.rwflag == READ {
		if !readOwned || lm.readLockStatus.Contains(key, clientId) {
			panicStr, _ := fmt.Printf("server tends to release a read key that client does not own %s by %s\n", key, clientId)
			panic(panicStr)
		}
		lm.readLockStatus.Remove(key, clientId)
		// update the clientLocks set
		lm.clientsIdLock.Lock(clientId)
		lm.clientReadLocks.Remove(clientId, key)
		lm.clientsIdLock.Unlock(clientId)

		if lm.readLockStatus.Empty(key) {
			lm.readLockStatus.Delete(key)
		}

	} else if request.rwflag == WRITE {
		if !writeOwned || (writeOwned && lm.writeLocksStatus.Get(key) != clientId) {
			panicStr, _ := fmt.Printf("server tends to release a write key that client does not own %s by %s\n", key, clientId)
			panic(panicStr)
		}
		lm.writeLocksStatus.Delete(key)
		// update the clientLocks set
		lm.clientsIdLock.Lock(key)
		lm.clientWriteLocks.Remove(clientId, key)
		lm.clientsIdLock.Unlock(key)
	}
	goodToGoPool = lm.openRequestQueueValve(key)
	lm.notifyClientsToProceed(goodToGoPool)

	lm.keysLock.Unlock(key)
}

func (lm *LockManager) openRequestQueueValve(key string) []string {
	utils.Nlog("valve begins")
	// key: the key we're allowing clients to acquire
	// retVal: goodToGoPool is for new clients that successfully acquire the lock, we need to signal them
	// nobody's owned the lock
	// readLockStatusLock locked
	// writeLocksStatusLock locked
	// requestQueueMapLock locked
	goodToGoPool := make([]string, 0)
	// first check whether nobody is using the key at all
	readOwned := lm.readLockStatus.Exists(key)
	writeOwned := lm.writeLocksStatus.Exists(key)

	if readOwned && writeOwned {
		panicStr, _ := fmt.Printf("server has both read and write at one time for key %s\n", key)
		panic(panicStr)
	}
	if writeOwned {
		utils.Nlog("write owned by other client")
		// we cannot let any one in because somebody is holding the exclusive lock
		return goodToGoPool
	}

	// rwValve: nobody's holding any rwlock of the key, so we're allowed to make write acquire coming in also
	rwValve := !readOwned && !writeOwned
	if rwValve && !lm.requestsQueueMap.Empty(key) && lm.requestsQueueMap.Head(key).rwflag == WRITE {
		utils.Nlog("letting out write key: [%s]", key)
		// only let the next write out
		headRequest := lm.requestsQueueMap.PopHead(key)
		// update the writeLocksStatus
		lm.writeLocksStatus.Set(key, headRequest.clientId)
		// update the clientLocks set
		lm.clientsIdLock.Lock(key)
		lm.clientWriteLocks.Append(headRequest.clientId, key)
		lm.clientsIdLock.Unlock(key)
		// update goodToGoPool
		goodToGoPool = append(goodToGoPool, headRequest.clientId)
		return goodToGoPool
	}
	// last possibility:
	// readOwnedByOtherClient && !writeOwnedByOtherClient
	// we're letting more readers coming in
	for !lm.requestsQueueMap.Empty(key) && lm.requestsQueueMap.Head(key).rwflag != WRITE {
		utils.Nlog("letting out read key: [%s]", key)
		// only let the next reads out
		headRequest := lm.requestsQueueMap.PopHead(key)
		// update the readLockStatus
		if lm.readLockStatus.Contains(key, headRequest.clientId) {
			lm.readLockStatus.Append(key, headRequest.clientId)
			// update the clientLocks set
			lm.clientsIdLock.Lock(headRequest.clientId)
			lm.clientReadLocks.Append(headRequest.clientId, key)
			lm.clientsIdLock.Unlock(headRequest.clientId)
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
