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

	readLockStatusLock sync.Mutex
	readLockStatus  map[string][]string

	// key -> list of array clientIds
	writeLocksStatusLock sync.Mutex
	writeLocksStatus map[string]string

	// key -> []LockRequest
	requestQueueMapLock sync.Mutex
	requestsQueueMap map[string][]LockRequest
	UnimplementedLockServiceServer

	waitersLock sync.Mutex
	waiters map[string]chan bool
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

		ls.waitersLock.Lock()
		waiter, ok := ls.waiters[key]
		if !ok {
			waiter = make(chan bool)
			ls.waiters[key] = waiter
		}
		ls.waitersLock.Unlock()
		<- waiter
	}

	return &Success{Flag: true}, nil
}

func (ls *LockServer) processAcquireRequest(request LockRequest) {
	key := request.key
	rwflag := request.rwflag
	clientId := request.clientId

	ls.requestQueueMapLock.Lock()
	ls.requestsQueueMap[key] = append(ls.requestsQueueMap[key], request)
	ls.requestQueueMapLock.Unlock()

	ls.readLockStatusLock.Lock()
	_, readOwnedByOtherClient := ls.readLockStatus[key]
	ls.readLockStatusLock.Unlock()

	ls.writeLocksStatusLock.Lock()
	_, writeOwnedByOtherClient := ls.writeLocksStatus[key]
	ls.writeLocksStatusLock.Unlock()

	if readOwnedByOtherClient && writeOwnedByOtherClient {
		panicStr, _ := fmt.Printf("server [%d] has both read and write at one time for key %s\n", ls.addr, key)
		panic(panicStr)
	}

	if readOwnedByOtherClient {

	}

	if writeOwnedByOtherClient {

	}
}

func (ls *LockServer) Release(ctx context.Context, locksInfo *ReleaseLocksInfo) (*Success, error) {

	return &Success{Flag: true}, nil
}

//func (ls *LockServer) openValve() {
//
//	return &Success{Flag: true}, nil
//}

var _ LockServerInterface = new(LockServer)

func NewLockServer(addr string) *LockServer {
	return &LockServer{
		addr: addr,
		readLockStatus:  make(map[string][]string),
		writeLocksStatus: make(map[string]string),
	}
}
