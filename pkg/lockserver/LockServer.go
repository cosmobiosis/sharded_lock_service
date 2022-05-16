package lockserver

import (
	context "context"
	"sync"
)

type LockServer struct {
	// clientId -> list of array keys
	readHMapLock sync.Mutex
	readLockKeeper  map[string][]string

	// clientId -> list of array keys
	writeMapLock sync.Mutex
	writeLocksKeeper map[string][]string
	UnimplementedLockServiceServer
}

func (ls *LockServer) Acquire(ctx context.Context, locksInfo *AcquireLocksInfo) (*Success, error) {
	// clientId := locksInfo.ClientId
	ls.writeMapLock.Lock()
	ls.writeMapLock.Unlock()
	return &Success{Flag: true}, nil
}

func (ls *LockServer) Release(ctx context.Context, locksInfo *ReleaseLocksInfo) (*Success, error) {
	return &Success{Flag: true}, nil
}

// This line guarantees all method for BlockStore are implemented
var _ LockServerInterface = new(LockServer)

func NewLockServer() *LockServer {
	return &LockServer{
		readLockKeeper:  make(map[string][]string),
		writeLocksKeeper: make(map[string][]string),
	}
}
