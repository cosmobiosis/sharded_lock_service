package lockserver

import (
	"fmt"
	"sharded_lock_service/pkg/types"
	"sharded_lock_service/pkg/utils"
	"sync"
)

type LockManager struct {
	keysLock       *utils.KeysLock
	readLockStatus *utils.ConcurrentStringSliceMap
	// key -> list of array clientIds
	writeLocksStatus *utils.ConcurrentStringMap
	// key -> []LockRequest
	requestsQueueMap *utils.ConcurrentLockRequestSliceMap

	waitersLock sync.Mutex
	waiters     map[string]chan bool

	clientsIdLock    *utils.KeysLock
	clientReadLocks  *utils.ConcurrentStringSliceMap
	clientWriteLocks *utils.ConcurrentStringSliceMap

	clientLeaseMu sync.Mutex
	clientLease   map[string]int64

	UnimplementedLockServiceServer
}

func NewLockManager() *LockManager {
	return &LockManager{
		keysLock:         utils.NewKeysLock(),
		readLockStatus:   utils.NewConcurrentStringSliceMap(),
		writeLocksStatus: utils.NewConcurrentStringMap(),
		requestsQueueMap: utils.NewConcurrentLockRequestSliceMap(),

		waiters: make(map[string]chan bool),

		clientsIdLock:    utils.NewKeysLock(),
		clientReadLocks:  utils.NewConcurrentStringSliceMap(),
		clientWriteLocks: utils.NewConcurrentStringSliceMap(),

		clientLease: make(map[string]int64),
	}
}

func (lm *LockManager) processAcquireRequest(request types.LockRequest, isKeeper bool) {
	key := request.Key
	clientId := request.ClientId
	var goodToGoPool []string

	lm.keysLock.Lock(key)
	defer lm.keysLock.Unlock(key)

	utils.Nlog("Client [%s] tends to acquire [%s]", clientId, key)
	if isKeeper {
		lm.requestsQueueMap.PushHead(key, request)
	} else {
		lm.requestsQueueMap.Append(key, request)
	}
	readOwned := lm.readLockStatus.Exists(key)
	writeOwned := lm.writeLocksStatus.Exists(key)

	if readOwned && writeOwned {
		panicStr, _ := fmt.Printf("server has both read and write at one time for key %s\n", key)
		panic(panicStr)
	}
	nobodyHasLock := !readOwned && !writeOwned
	somebodyNotMeHasReadLock := !lm.readLockStatus.Contains(key, clientId) && readOwned
	if nobodyHasLock || somebodyNotMeHasReadLock {
		goodToGoPool = lm.openRequestQueueValve(key)
		lm.notifyClientsToProceed(goodToGoPool)
	}
}

func (lm *LockManager) processReleaseRequest(request types.LockRequest) {
	key := request.Key
	clientId := request.ClientId
	var goodToGoPool []string

	lm.keysLock.Lock(key)
	defer lm.keysLock.Unlock(key)

	readOwned := lm.readLockStatus.Exists(key)
	writeOwned := lm.writeLocksStatus.Exists(key)

	if readOwned && writeOwned {
		panicStr, _ := fmt.Printf("server has both read and write at one time for key %s\n", key)
		panic(panicStr)
	}
	utils.Nlog("Client [%s] tends to release [%s]", clientId, key)
	if request.Rwflag == types.READ {
		if !readOwned || !lm.readLockStatus.Contains(key, clientId) {
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

	} else if request.Rwflag == types.WRITE {
		if !writeOwned || (writeOwned && lm.writeLocksStatus.Get(key) != clientId) {
			panicStr, _ := fmt.Printf("client [%s] tends to release a write key [%s] he does not own\n", clientId, key)
			if writeOwned {
				fmt.Printf("The lock is currently owned by client [%s]\n", lm.writeLocksStatus.Get(key))
			}
			panic(panicStr)
		}
		lm.writeLocksStatus.Delete(key)
		// update the clientLocks setx
		lm.clientsIdLock.Lock(key)
		lm.clientWriteLocks.Remove(clientId, key)
		lm.clientsIdLock.Unlock(key)
	}
	goodToGoPool = lm.openRequestQueueValve(key)
	lm.notifyClientsToProceed(goodToGoPool)
}

func (lm *LockManager) openRequestQueueValve(key string) []string {
	utils.Nlog("%s: valve opens", key)
	// key: the key we're allowing clients to acquire
	// retVal: goodToGoPool is for new clients that successfully acquire the lock, we need to signal them
	// nobody's owned the lock
	goodToGoPool := make([]string, 0)
	requestQueueExists := lm.requestsQueueMap.Exists(key)
	if !requestQueueExists {
		return goodToGoPool
	}
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
	if rwValve && !lm.requestsQueueMap.Empty(key) && lm.requestsQueueMap.Head(key).Rwflag == types.WRITE {
		// only let the next write out
		headRequest := lm.requestsQueueMap.PopHead(key)
		// update the writeLocksStatus
		lm.writeLocksStatus.Set(key, headRequest.ClientId)
		// update the clientLocks set
		lm.clientsIdLock.Lock(key)
		lm.clientWriteLocks.Append(headRequest.ClientId, key)
		lm.clientsIdLock.Unlock(key)
		// update goodToGoPool
		utils.Nlog("letting out write key: [%s] for [%s]", key, headRequest.ClientId)
		goodToGoPool = append(goodToGoPool, headRequest.ClientId)
		return goodToGoPool
	}
	// last possibility:
	// readOwnedByOtherClient && !writeOwnedByOtherClient
	// we're letting more readers coming in
	for !lm.requestsQueueMap.Empty(key) && lm.requestsQueueMap.Head(key).Rwflag != types.WRITE {
		// only let the next reads out
		headRequest := lm.requestsQueueMap.PopHead(key)
		// update the readLockStatus
		if !lm.readLockStatus.Contains(key, headRequest.ClientId) {
			lm.readLockStatus.Append(key, headRequest.ClientId)
			// update the clientLocks set
			lm.clientsIdLock.Lock(headRequest.ClientId)
			lm.clientReadLocks.Append(headRequest.ClientId, key)
			lm.clientsIdLock.Unlock(headRequest.ClientId)
		}
		// update goodToGoPool
		utils.Nlog("letting out read key: [%s] for [%s]", key, headRequest.ClientId)
		goodToGoPool = append(goodToGoPool, headRequest.ClientId)
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
