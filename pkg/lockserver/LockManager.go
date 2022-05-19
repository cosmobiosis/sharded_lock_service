package lockserver

import "sync"

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