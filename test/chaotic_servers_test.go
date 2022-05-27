package test

import (
	context "context"
	"fmt"
	grpc "google.golang.org/grpc"
	"sharded_lock_service/pkg/lockserver"
	"testing"
	"time"
)

func TestChaotic(t *testing.T) {
	testingInfo := InitTest(1, 1000)
	serverAddr := testingInfo.serverAddrs[0]
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	c := lockserver.NewLockServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := c.Acquire(ctx, &lockserver.AcquireLocksInfo{
		ClientId: "a",
		ReadKeys: []string{"rkey1", "rkey2"},
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(resp.Flag)
	time.Sleep(time.Second)

	endTestChan := make(chan struct{})

	go func() {
		resp, err = c.Acquire(ctx, &lockserver.AcquireLocksInfo{
			ClientId:  "b",
			WriteKeys: []string{"rkey1", "rkey2"},
		})
		if err != nil {
			panic(err)
		}
		fmt.Println(resp.Flag)
		time.Sleep(time.Second)

		resp, err = c.Release(ctx, &lockserver.ReleaseLocksInfo{
			ClientId:  "b",
			WriteKeys: []string{"rkey1", "rkey2"},
		})
		if err != nil {
			panic(err)
		}
		fmt.Println(resp.Flag)
		endTestChan <- struct{}{}
	}()

	resp, err = c.Release(ctx, &lockserver.ReleaseLocksInfo{
		ClientId: "a",
		ReadKeys: []string{"rkey2", "rkey1"},
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(resp.Flag)

	<-endTestChan
	testingInfo.ShutdownChannels[0] <- true
}
