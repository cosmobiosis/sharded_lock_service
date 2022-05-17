package lockserver

import "sync"

type LockManager struct {
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
