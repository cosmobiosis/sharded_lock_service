package lockserver

import (
	context "context"
	"fmt"
	"sort"
)

type RWFlag string
const (
	READ RWFlag = "read"
	WRITE     = "write"
)

type LockRequest struct {
	key string
	rwflag RWFlag
	clientId string
}

type LockServer struct {
	// key -> list of array clientIds
	addr string
	lm *LockManager
	UnimplementedLockServiceServer
}

func (ls *LockServer) Acquire(ctx context.Context, locksInfo *AcquireLocksInfo) (*Success, error) {
	clientId := locksInfo.ClientId

	rwFlagSet := make(map[string]RWFlag)
	for _, key := range locksInfo.ReadKeys {
		rwFlagSet[key] = READ
	}
	for _, key := range locksInfo.WriteKeys {
		rwFlagSet[key] = WRITE
	}

	keysToAcquire := append(locksInfo.ReadKeys, locksInfo.WriteKeys...)

	// IMPORTANT!!! acquire lock by sorting sequence
	sort.Strings(keysToAcquire)

	for _, key := range keysToAcquire {
		request := LockRequest{
			key: key,
			rwflag: rwFlagSet[key],
			clientId: clientId,
		}
		ls.processAcquireRequest(request)

		ls.lm.waitersLock.Lock()
		waiter, ok := ls.lm.waiters[clientId]
		if !ok {
			waiter = make(chan bool)
			ls.lm.waiters[clientId] = waiter
		}
		ls.lm.waitersLock.Unlock()
		<- waiter
	}

	return &Success{Flag: true}, nil
}

func (ls *LockServer) processAcquireRequest(request LockRequest) {
	key := request.key
	clientId := request.clientId
	var goodToGoPool []string

	ls.lm.metaMu.Lock()

	ls.lm.requestsQueueMap[key] = append(ls.lm.requestsQueueMap[key], request)
	curReaders, readOwnedByOtherClient := ls.lm.readLockStatus[key]
	_, writeOwnedByOtherClient := ls.lm.writeLocksStatus[key]

	if readOwnedByOtherClient && writeOwnedByOtherClient {
		panicStr, _ := fmt.Printf("server [%s] has both read and write at one time for key %s\n", ls.addr, key)
		panic(panicStr)
	}
	if readOwnedByOtherClient && !contains(curReaders, clientId) {
		goodToGoPool = ls.openRequestQueueValve(key)
		ls.notifyClientsToProceed(goodToGoPool)
	}
	ls.lm.metaMu.Unlock()
}

func (ls *LockServer) Release(ctx context.Context, locksInfo *ReleaseLocksInfo) (*Success, error) {
	clientId := locksInfo.ClientId

	rwFlagSet := make(map[string]RWFlag)
	for _, key := range locksInfo.ReadKeys {
		rwFlagSet[key] = READ
	}
	for _, key := range locksInfo.WriteKeys {
		rwFlagSet[key] = WRITE
	}

	keysToRelease := append(locksInfo.ReadKeys, locksInfo.WriteKeys...)
	sort.Strings(keysToRelease)
	reverse(keysToRelease)

	for _, key := range keysToRelease {
		request := LockRequest{
			key: key,
			rwflag: rwFlagSet[key],
			clientId: clientId,
		}
		ls.processReleaseRequest(request)
	}
	// release lock by reversed sorting sequence
	return &Success{Flag: true}, nil
}


func (ls *LockServer) processReleaseRequest(request LockRequest) {
	key := request.key
	clientId := request.clientId
	var goodToGoPool []string

	ls.lm.metaMu.Lock()

	curReaders, readOwnedByOtherClient := ls.lm.readLockStatus[key]
	curWriter, writeOwnedByOtherClient := ls.lm.writeLocksStatus[key]

	if readOwnedByOtherClient && writeOwnedByOtherClient {
		panicStr, _ := fmt.Printf("server [%s] has both read and write at one time for key %s\n", ls.addr, key)
		panic(panicStr)
	}
	if request.rwflag == READ {
		if !readOwnedByOtherClient || (readOwnedByOtherClient && !contains(curReaders, clientId)) {
			panicStr, _ := fmt.Printf("server [%s] tends to release a read key that client does not own %s by %s\n", ls.addr, key, clientId)
			panic(panicStr)
		}
		ls.lm.readLockStatus[key] = remove(ls.lm.readLockStatus[key], clientId)
		if len(ls.lm.readLockStatus[key]) == 0 {
			delete(ls.lm.readLockStatus, key)
		}
	} else if request.rwflag == WRITE {
		if !writeOwnedByOtherClient || (writeOwnedByOtherClient && curWriter != clientId) {
			panicStr, _ := fmt.Printf("server [%s] tends to release a write key that client does not own %s by %s\n", ls.addr, key, clientId)
			panic(panicStr)
		}
		delete(ls.lm.writeLocksStatus, key)
	}
	goodToGoPool = ls.openRequestQueueValve(key)
	ls.notifyClientsToProceed(goodToGoPool)

	ls.lm.metaMu.Unlock()
}


func (ls *LockServer) openRequestQueueValve(key string) []string {
	// key: the key we're allowing clients to acquire
	// retVal: goodToGoPool is for new clients that successfully acquire the lock, we need to signal them
	// nobody's owned the lock
	// readLockStatusLock locked
	// writeLocksStatusLock locked
	// requestQueueMapLock locked
	goodToGoPool := make([]string, 0)
	var headRequest LockRequest
	// first check whether nobody is using the key at all
	_, readOwnedByOtherClient := ls.lm.readLockStatus[key]
	_, writeOwnedByOtherClient := ls.lm.writeLocksStatus[key]

	if readOwnedByOtherClient && writeOwnedByOtherClient {
		panicStr, _ := fmt.Printf("server [%s] has both read and write at one time for key %s\n", ls.addr, key)
		panic(panicStr)
	}
	if writeOwnedByOtherClient {
		// we cannot let any one in because somebody is holding the exclusive lock
		return goodToGoPool
	}

	// rwValve: nobody's holding any rwlock of the key, so we're allowed to make write acquire coming in also
	rwValve := !readOwnedByOtherClient && !writeOwnedByOtherClient
	if rwValve && len(ls.lm.requestsQueueMap[key]) > 0 && ls.lm.requestsQueueMap[key][0].rwflag == WRITE {
		// only let the next write out
		headRequest, ls.lm.requestsQueueMap[key] = ls.lm.requestsQueueMap[key][0], ls.lm.requestsQueueMap[key][1:]
		// update the writeLocksStatus
		ls.lm.writeLocksStatus[key] = headRequest.clientId
		// update goodToGoPool
		goodToGoPool = append(goodToGoPool, headRequest.clientId)
		return goodToGoPool
	}
	// last possibility:
	// readOwnedByOtherClient && !writeOwnedByOtherClient
	// we're letting more readers coming in
	for len(ls.lm.requestsQueueMap[key]) > 0 && ls.lm.requestsQueueMap[key][0].rwflag != WRITE {
		// only let the next reads out
		headRequest, ls.lm.requestsQueueMap[key] = ls.lm.requestsQueueMap[key][0], ls.lm.requestsQueueMap[key][1:]
		// update the readLockStatus
		if !contains(ls.lm.readLockStatus[key], headRequest.clientId) {
			ls.lm.readLockStatus[key] = append(ls.lm.readLockStatus[key], headRequest.clientId)
		}
		// update goodToGoPool
		goodToGoPool = append(goodToGoPool, headRequest.clientId)
	}
	return goodToGoPool
}

func (ls *LockServer) notifyClientsToProceed(goodToGoPool []string) {
	// notify corresponding clients to proceed
	for _, clientToNotify := range goodToGoPool {
		go func(clientId string) {
			ls.lm.waitersLock.Lock()
			waiter, ok := ls.lm.waiters[clientId]
			if !ok {
				waiter = make(chan bool)
				ls.lm.waiters[clientId] = waiter
			}
			ls.lm.waitersLock.Unlock()
			waiter <- true
		}(clientToNotify)
	}
}

var _ LockServerInterface = new(LockServer)

func NewLockServer(addr string, lockManager *LockManager) *LockServer {
	return &LockServer{
		addr: addr,
		lm: lockManager,
	}
}
