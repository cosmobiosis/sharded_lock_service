package lockserver

import (
	context "context"
	"fmt"
	"sync"
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
		waiter, ok := ls.lm.waiters[key]
		if !ok {
			waiter = make(chan bool)
			ls.lm.waiters[key] = waiter
		}
		ls.lm.waitersLock.Unlock()
		<- waiter
	}

	return &Success{Flag: true}, nil
}

func (ls *LockServer) processAcquireRequest(request LockRequest) {
	key := request.key
	rwflag := request.rwflag
	clientId := request.clientId

	ls.lm.requestQueueMapLock.Lock()
	ls.lm.requestsQueueMap[key] = append(ls.lm.requestsQueueMap[key], request)
	ls.lm.requestQueueMapLock.Unlock()

	ls.lm.readLockStatusLock.Lock()
	_, readOwnedByOtherClient := ls.lm.readLockStatus[key]
	ls.lm.readLockStatusLock.Unlock()

	ls.lm.writeLocksStatusLock.Lock()
	_, writeOwnedByOtherClient := ls.lm.writeLocksStatus[key]
	ls.lm.writeLocksStatusLock.Unlock()

	if readOwnedByOtherClient && writeOwnedByOtherClient {
		panicStr, _ := fmt.Printf("server [%s] has both read and write at one time for key %s\n", ls.addr, key)
		panic(panicStr)
	}

	if readOwnedByOtherClient {

	} else if writeOwnedByOtherClient {
		// ls.openValve()
	}
}

func (ls *LockServer) Release(ctx context.Context, locksInfo *ReleaseLocksInfo) (*Success, error) {

	return &Success{Flag: true}, nil
}

func (ls *LockServer) openValve(key string) {

}

var _ LockServerInterface = new(LockServer)

func NewLockServer(addr string, lockManager *LockManager) *LockServer {
	return &LockServer{
		addr: addr,
		lm: lockManager,
	}
}
