package lockserver

import (
	context "context"
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	"sharded_lock_service/pkg/types"
	"sharded_lock_service/pkg/utils"
	"sort"
)

type LockServer struct {
	// key -> list of array clientIds
	addr string
	lm *LockManager
	leaseM *LeaseManager
	UnimplementedLockServiceServer
}

func (ls *LockServer) Acquire(ctx context.Context, locksInfo *AcquireLocksInfo) (*Success, error) {
	clientId := locksInfo.ClientId
	exp := ls.leaseM.LeaseExtend(clientId)
	ls.leaseM.PushDeadlineInfo(exp, clientId)

	rwFlagSet := make(map[string]types.RWFlag)
	for _, key := range locksInfo.ReadKeys {
		ls.leaseM.clientReadLeases.Append(clientId, key)
		rwFlagSet[key] = types.READ
	}
	for _, key := range locksInfo.WriteKeys {
		ls.leaseM.clientWriteLeases.Append(clientId, key)
		rwFlagSet[key] = types.WRITE
	}

	keysToAcquire := append(locksInfo.ReadKeys, locksInfo.WriteKeys...)

	// IMPORTANT!!! acquire lock by sorting sequence
	sort.Strings(keysToAcquire)

	for _, key := range keysToAcquire {
		request := types.LockRequest{
			Key:      key,
			Rwflag:   rwFlagSet[key],
			ClientId: clientId,
		}
		ls.lm.processAcquireRequest(request, locksInfo.IsKeeper)

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

	rwFlagSet := make(map[string]types.RWFlag)
	for _, key := range locksInfo.ReadKeys {
		rwFlagSet[key] = types.READ
	}
	for _, key := range locksInfo.WriteKeys {
		rwFlagSet[key] = types.WRITE
	}

	keysToRelease := append(locksInfo.ReadKeys, locksInfo.WriteKeys...)
	sort.Strings(keysToRelease)
	utils.SliceReverse(keysToRelease)

	for _, key := range keysToRelease {
		request := types.LockRequest{
			Key: key,
			Rwflag: rwFlagSet[key],
			ClientId: clientId,
		}
		ls.lm.processReleaseRequest(request)
	}
	// release lock by reversed sorting sequence
	return &Success{Flag: true}, nil
}

func (ls *LockServer) Heartbeat(ctx context.Context, info *HeartbeatInfo) (*Success, error) {
	ls.leaseM.LeaseExtend(info.ClientId)
	return &Success{Flag: true}, nil
}

func (ls *LockServer) Ping(ctx context.Context, request *PingRequest) (*Success, error) {
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