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
	testingInfo := InitTest(1, 1000)

	testingInfo.ShutdownChannels[0] <- true
}