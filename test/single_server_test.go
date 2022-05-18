package test

import (
	"fmt"
	"sharded_lock_service/pkg/lockserver"
	"strconv"
	"testing"
	context "context"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

func TestBasicReadAcquire(t *testing.T) {
	shutChan := make(chan bool)
	err := lockserver.StartServer(":8080", shutChan)
	if err != nil {
		panic(err)
	}
	
	shutChan <- true
}