package main

import (
	context "context"
	"fmt"
	"google.golang.org/grpc"
	"sharded_lock_service/pkg/lockserver"
	"time"
)

func main() {
	conn, err := grpc.Dial("localhost:2000", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	c := lockserver.NewLockServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 2 * time.Second)
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
}