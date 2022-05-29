package test

import (
	"sync"
	"testing"
	"time"

	//context "context"
	//grpc "google.golang.org/grpc"
	//"sharded_lock_service/pkg/lockserver"
	//"time"
)

func TestChaotic(t *testing.T) {
	NUM_SERVERS := 10
	testingInfo := InitTest(NUM_SERVERS, 1000)
	time.Sleep(2 * time.Second)
	txns := InitChaoticTestingEnv(TestTxnsConfig{
		numTxn: 300,
		numKeysInPool: 10,
		keyLength: 10,
		readPerTxn: 2,
		writePerTxn: 5,
	})
	txnWorkers := make([]*TransactionWorker, 0)
	for i := 0; i < len(txns); i++ {
		txnWorkers = append(txnWorkers, NewTransactionWorker(testingInfo.serverAddrs, txns[i]))
	}
	wg := sync.WaitGroup{}
	wg.Add(len(txnWorkers))
	for i := 0; i < len(txnWorkers); i++ {
		go func(iCopy int) {
			txnWorkers[iCopy].startTxn()
			wg.Done()
		}(i)
	}
	wg.Wait()
	for i := 0; i < len(testingInfo.ShutdownChannels); i++ {
		testingInfo.ShutdownChannels[i] <- true
	}
}
