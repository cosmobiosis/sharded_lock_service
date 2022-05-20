package test

import (
	context "context"
	"fmt"
	grpc "google.golang.org/grpc"
	lockserver "sharded_lock_service/pkg/lockserver"
	"testing"
	"time"
)

func TestBasicReadAcquire(t *testing.T) {
	testingInfo := InitTest(1, 1000)
	serverAddr := testingInfo.serverAddrs[0]
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	c := lockserver.NewLockServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()
	resp, err := c.Acquire(ctx, &lockserver.AcquireLocksInfo{
		ClientId: "a",
		ReadKeys: []string{"rkey1", "rkey2"},
		WriteKeys: []string{"wkey3"},
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(resp.Flag)
	testingInfo.ShutdownChannels[0] <- true
}