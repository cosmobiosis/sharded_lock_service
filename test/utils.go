package test

import (
	"fmt"
	"sharded_lock_service/pkg/lockserver"
	"strconv"
	"testing"
	context "context"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

type TestInfo struct {
	ShutdownChannels []chan bool
}

func InitTest(numServers int, startPort int) *TestInfo {
	shutdownChannels := make([]chan bool, 0)
	for i := 1; i < numServers; i++ {
		serverAddr := ":" + strconv.Itoa(startPort + i)
		shutChan := make(chan bool)
		shutdownChannels = append(shutdownChannels, shutChan)
		go func() {
			err := lockserver.StartServer(serverAddr, shutChan)
			if err != nil {
				panic("Server Start failure")
			}
		}()
	}

	retVal := &TestInfo{
		ShutdownChannels: shutdownChannels,
	}
	return retVal
}
