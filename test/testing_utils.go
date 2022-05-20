package test

import (
	"fmt"
	"sharded_lock_service/pkg/lockserver"
	"strconv"
)

type TestInfo struct {
	serverAddrs []string
	ShutdownChannels []chan bool
}

func InitTest(numServers int, startPort int) *TestInfo {
	shutdownChannels := make([]chan bool, 0)
	serverAddrs := make([]string, 0)
	for i := 0; i < numServers; i++ {
		serverAddr := ":" + strconv.Itoa(startPort + i)
		serverAddrs = append(serverAddrs, "localhost" + serverAddr)
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
		serverAddrs: serverAddrs,
		ShutdownChannels: shutdownChannels,
	}
	fmt.Println(retVal.serverAddrs)
	return retVal
}
