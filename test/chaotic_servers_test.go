package test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	//context "context"
	//grpc "google.golang.org/grpc"
	//"sharded_lock_service/pkg/lockserver"
	//"time"
)

func chaoticTest(config TestTxnsConfig) int64 {
	// return time
	testingInfo, txns := InitChaoticTestingEnv(config)
	txnWorkers := make([]*TransactionWorker, 0)
	for i := 0; i < len(txns); i++ {
		txnWorkers = append(txnWorkers, NewTransactionWorker(testingInfo.serverAddrs, txns[i], config.readLatency, config.writeLatency))
	}
	wg := sync.WaitGroup{}
	wg.Add(len(txnWorkers))
	start := time.Now().UTC().UnixMilli()
	for i := 0; i < len(txnWorkers); i++ {
		go func(iCopy int) {
			txnWorkers[iCopy].startTxn()
			wg.Done()
		}(i)
	}
	wg.Wait()
	end := time.Now().UTC().UnixMilli()
	ShutdownTest(testingInfo)
	return end - start
}

func TestSampleChaotic(t *testing.T) {
	config := TestTxnsConfig{
		numServers: 20,
		readLatency: 1,
		writeLatency: 1,
		keyLength: 10,
		numTxn: 100,
		numKeysInPool: 100,
		readPerTxn: 4,
		writePerTxn: 4,
	}
	chaoticTest(config)
}

func TestServerShardPerformance(t *testing.T) {
	config := TestTxnsConfig {
		numServers: 1,
		readLatency: 100,
		writeLatency: 100,
		keyLength: 10,
		numTxn: 10,
		numKeysInPool: 500,
		readPerTxn: 50,
		writePerTxn: 50,
	}
	gap := chaoticTest(config)
	fmt.Println(gap)
}

func TestTxnServerIncrease(t *testing.T) {
	for numServers := 1; numServers <= 30; numServers ++ {
		config := TestTxnsConfig {
			numServers:    numServers,
			readLatency:   0,
			writeLatency:  0,
			keyLength:     10,
			numTxn:        20,
			numKeysInPool: 1000,
			readPerTxn:    50,
			writePerTxn:   0,
		}
		timeGap := chaoticTest(config)
		fmt.Println(timeGap)
	}
}