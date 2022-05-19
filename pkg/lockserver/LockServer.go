package lockserver

import (
	context "context"
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	"sharded_lock_service/pkg/utils"
	"sort"
	"time"
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
	if len(locksInfo.WriteKeys) == 0 && len(locksInfo.ReadKeys) == 0 {
		ls.lm.clientLeaseMu.Lock()
		ls.lm.clientLease[locksInfo.ClientId] = time.Now().Unix()
		ls.lm.clientLeaseMu.Unlock()
	}
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
		ls.lm.processAcquireRequest(request)

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
	utils.SliceReverse(keysToRelease)

	for _, key := range keysToRelease {
		request := LockRequest{
			key: key,
			rwflag: rwFlagSet[key],
			clientId: clientId,
		}
		ls.lm.processReleaseRequest(request)
	}
	// release lock by reversed sorting sequence
	return &Success{Flag: true}, nil
}

var _ LockServerInterface = new(LockServer)

func NewLockServer(addr string, lockManager *LockManager) *LockServer {
	return &LockServer{
		addr: addr,
		lm: lockManager,
	}
}

func StartServer(hostAddr string, shutChan chan bool) error {
	lis, err := net.Listen("tcp", hostAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	server := grpc.NewServer()
	lockManager := NewLockManager()
	lockServer := NewLockServer(hostAddr, lockManager)
	RegisterLockServiceServer(server, lockServer)

	go func() {
		<- shutChan
		server.Stop()
	}()

	fmt.Println("Setting up server", hostAddr)
	if errServe := server.Serve(lis); errServe != nil {
		return errors.New("fail to serve")
	}
	return nil
}