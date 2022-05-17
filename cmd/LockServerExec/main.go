package main

import (
	"sharded_lock_service/pkg/lockserver"
	"errors"
	"google.golang.org/grpc"
	"log"
	"net"
	"strconv"

	// "io/ioutil"
	// "flag"
	// "fmt"
	//"os"
	//"strconv"
	//"strings"
)

func main() {
	// Custom flag Usage message
	NUM_SERVERS := 10
	START_ADDR := 1000
	var shutdownChannels []chan bool
	setupErrDetected := make(chan bool)

	for i := 1; i < NUM_SERVERS; i++ {
		serverAddr := strconv.Itoa(START_ADDR + i)
		var shutChan chan bool
		shutdownChannels = append(shutdownChannels, shutChan)
		go func() {
			err := startServer(serverAddr, shutChan)
			if err != nil {
				setupErrDetected <- true
			}
		}()
	}
	<- setupErrDetected
	panic("Lock Servers Set Up Error Detected")
}

func startServer(hostAddr string, shutChan chan bool) error {
	lis, err := net.Listen("tcp", hostAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	server := grpc.NewServer()
	lockServer := lockserver.NewLockServer()
	lockserver.RegisterLockServiceServer(server, lockServer)

	go func() {
		<- shutChan
		server.Stop()
	}()

	if errServe := server.Serve(lis); errServe != nil {
		return errors.New("fail to serve")
	}
	return nil
}
