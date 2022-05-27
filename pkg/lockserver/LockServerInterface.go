package lockserver

import (
	context "context"
)

type LockServerInterface interface {
	Acquire(ctx context.Context, locksInfo *AcquireLocksInfo) (*Success, error)
	Release(ctx context.Context, locksInfo *ReleaseLocksInfo) (*Success, error)
	Ping(ctx context.Context, request *PingRequest) (*Success, error)
}
