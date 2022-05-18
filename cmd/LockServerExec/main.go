package main

import (
	"flag"
	"sharded_lock_service/pkg/lockserver"
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
	setupErrDetected := make(chan bool)

	numServers := flag.Int("numServers", 10, "number of total servers to set up")
	startPort := flag.Int("startPort", 1000, "starting address of servers")
	// Parse the flags.
	flag.Parse()

	for i := 1; i < *numServers; i++ {
		serverAddr := ":" + strconv.Itoa(*startPort + i)
		var shutChan chan bool
		go func() {
			err := lockserver.StartServer(serverAddr, shutChan)
			if err != nil {
				setupErrDetected <- true
			}
		}()
	}
	<- setupErrDetected
	panic("Lock Servers Set Up Error Detected")
}
