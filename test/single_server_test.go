package test

import (
	context "context"
	"github.com/stretchr/testify/assert"
	grpc "google.golang.org/grpc"
	lockserver "sharded_lock_service/pkg/lockserver"
	"testing"
	"time"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func TestBasicReadAcquire(t *testing.T) {
	testingInfo := InitTest(1, 1000)
	serverAddr := testingInfo.serverAddrs[0]
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	check(err)
	c := lockserver.NewLockServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()
	resp, err := c.Acquire(ctx, &lockserver.AcquireLocksInfo{
		ClientId: "a",
		ReadKeys: []string{"rkey1", "rkey2"},
		WriteKeys: []string{"wkey3"},
	})
	check(err)
	testingInfo.ShutdownChannels[0] <- true
	assert.Equal(t, resp.Flag, true)
}

func TestDoubleReadAcquires(t *testing.T) {
	testingInfo := InitTest(1, 1000)
	serverAddr := testingInfo.serverAddrs[0]

	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	check(err)
	c1 := lockserver.NewLockServiceClient(conn)
	c2 := lockserver.NewLockServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()
	resp1, err := c1.Acquire(ctx, &lockserver.AcquireLocksInfo{
		ClientId: "a",
		ReadKeys: []string{"rkey1", "rkey2"},
		WriteKeys: []string{"wkey3"},
	})
	check(err)
	resp2, err := c2.Acquire(ctx, &lockserver.AcquireLocksInfo{
		ClientId: "b",
		ReadKeys: []string{"rkey1", "rkey2"},
	})
	check(err)
	testingInfo.ShutdownChannels[0] <- true
	assert.Equal(t, resp1.Flag, true)
	assert.Equal(t, resp2.Flag, true)
}

func TestSimpleReleaseAcquires(t *testing.T) {
	testingInfo := InitTest(1, 1000)
	serverAddr := testingInfo.serverAddrs[0]

	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	check(err)

	c1 := lockserver.NewLockServiceClient(conn)
	c2 := lockserver.NewLockServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()
	resps := make([]*lockserver.Success, 0);
	resp1, err := c1.Acquire(ctx, &lockserver.AcquireLocksInfo{
		ClientId: "a",
		ReadKeys: []string{},
		WriteKeys: []string{"wkey3"},
	})
	check(err)
	resps = append(resps, resp1)

	resp2, err := c1.Release(ctx, &lockserver.ReleaseLocksInfo{
		ClientId: "a",
		ReadKeys: []string{},
		WriteKeys: []string{"wkey3"},
	})
	check(err)
	resps = append(resps, resp2)

	resp3, err := c2.Acquire(ctx, &lockserver.AcquireLocksInfo{
		ClientId: "b",
		ReadKeys: []string{"wkey3"},
		WriteKeys: []string{},
	})
	check(err)
	resps = append(resps, resp3)

	testingInfo.ShutdownChannels[0] <- true

	for _, resp := range resps {
		assert.Equal(t, resp.Flag, true)
	}
}