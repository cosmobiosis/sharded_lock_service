package main

import (
	"flag"
	"fmt"
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

	numServers := flag.Int("numServers", 1, "number of total servers to set up")
	startPort := flag.Int("startPort", 2000, "starting address of servers")
	// Parse the flags.
	flag.Parse()

	for i := 0; i < *numServers; i++ {
		serverAddr := ":" + strconv.Itoa(*startPort + i)
		var shutChan chan bool
		go func() {
			err := lockserver.StartServer(serverAddr, shutChan)
			if err != nil {
				setupErrDetected <- true
			}
		}()
	}
	fmt.Println("Servers set up completed")
	<- setupErrDetected
	panic("Lock Servers Set Up Error Detected")
}
