package test

import (
	context "context"
	"fmt"
	"google.golang.org/grpc"
	"math"
	"math/rand"
	"sharded_lock_service/pkg/lockserver"
	"sharded_lock_service/pkg/utils"
	"strconv"
	"time"
)

type TestServerInfo struct {
	serverAddrs []string
	ShutdownChannels []chan bool
}

type TestTxnsConfig struct {
	numTxn int
	numKeysInPool int
	readPerTxn int
	writePerTxn int
	keyLength int
}

type RWFlag string
const (
	READ RWFlag = "read"
	WRITE     = "write"
)

type Operation struct {
	Rwflag   RWFlag
	Key    string
	Value    string
}

func InitTest(numServers int, startPort int) *TestServerInfo {
	shutdownChannels := make([]chan bool, 0)
	serverAddrs := make([]string, 0)
	for i := 0; i < numServers; i++ {
		serverAddr := ":" + strconv.Itoa(startPort + i)
		serverAddrs = append(serverAddrs, "127.0.0.1" + serverAddr)
		shutChan := make(chan bool)
		shutdownChannels = append(shutdownChannels, shutChan)
		go func() {
			err := lockserver.StartServer(serverAddr, shutChan)
			if err != nil {
				panic("Server Start failure")
			}
		}()
	}

	retVal := &TestServerInfo{
		serverAddrs: serverAddrs,
		ShutdownChannels: shutdownChannels,
	}
	fmt.Println(retVal.serverAddrs)
	return retVal
}

func InitChaoticTestingEnv(config TestTxnsConfig) []Transaction {
	txns := make([]Transaction, 0)
	keysPool := generateRandomKeyPools(config)
	for i := 0; i < config.numTxn; i++ {
		keySet := retrieveRandomSetOfKeys(keysPool, config)
		txn := generateRandomTxn(keySet, config)
		txns = append(txns, txn)
	}
	return txns
}

type Transaction struct {
	ops []Operation
}

type TransactionWorker struct {
	clientId string
	txn Transaction
	serverAddrs []string
	lockClients []*lockserver.LockServiceClient
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func (w *TransactionWorker) cacheGrpcClient(serverIndex int) {
	conn, err := grpc.Dial(w.serverAddrs[serverIndex], grpc.WithInsecure())
	check(err)
	c := lockserver.NewLockServiceClient(conn)
	w.lockClients[serverIndex] = &c
}

//func (w *TransactionWorker) closeGrpcClients() {
//	conn, err := grpc.Dial(w.serverAddrs[serverIndex], grpc.WithInsecure())
//	check(err)
//	c := lockserver.NewLockServiceClient(conn)
//	w.lockClients[serverIndex] = &c
//	for _, client := range w.lockClients {
//		(*client)
//	}
//}

func NewTransactionWorker(serverAddrs []string, txn Transaction) *TransactionWorker {
	txnWorker := TransactionWorker {
		txn: txn,
		serverAddrs: serverAddrs,
		clientId: generateRandomString(30),
		lockClients: make([]*lockserver.LockServiceClient, len(serverAddrs)),
	}
	for i := 0; i <  len(txnWorker.serverAddrs); i++ {
		txnWorker.cacheGrpcClient(i)
	}
	return &txnWorker
}

func (w *TransactionWorker) startTxn() {
	ctx, cancel := context.WithTimeout(context.Background(), 60 * time.Second)
	defer cancel()

	readBins := make([][]string, len(w.serverAddrs))
	writeBins := make([][]string, len(w.serverAddrs))

	for i := 0; i <  len(w.serverAddrs); i++ {
		readBins[i] = make([]string, 0)
		writeBins[i] = make([]string, 0)
	}

	for i := 0; i < len(w.txn.ops); i++  {
		op := w.txn.ops[i]
		binIndex := int(utils.Hash(op.Key) % uint32(len(w.serverAddrs)))
		if op.Rwflag == READ {
			readBins[binIndex] = append(readBins[binIndex], op.Key)
		} else {
			writeBins[binIndex] = append(writeBins[binIndex], op.Key)
		}
	}

	utils.Tlog("Begin Transaction %s", w.clientId)
	for i := 0; i < len(w.lockClients); i++ {
		readKeys := readBins[i]
		writeKeys := writeBins[i]
		c := w.lockClients[i]
		_, err := (*c).Acquire(ctx, &lockserver.AcquireLocksInfo {
			ClientId: w.clientId,
			ReadKeys: readKeys,
			WriteKeys: writeKeys,
		})
		check(err)
	}

	for i := 0; i < len(w.lockClients); i++ {
		readKeys := readBins[i]
		writeKeys := writeBins[i]
		c := w.lockClients[i]
		_, err := (*c).Release(ctx, &lockserver.ReleaseLocksInfo {
			ClientId: w.clientId,
			ReadKeys: readKeys,
			WriteKeys: writeKeys,
		})
		check(err)
	}
	utils.Tlog("End Transaction %s", w.clientId)
}

func inSlice(slice []int, target int) bool {
	for _, num := range slice {
		if num == target {
			return true
		}
	}
	return false
}

func generateRandomTxn(keySet []string, config TestTxnsConfig) Transaction {
	permutation := rand.Perm(len(keySet))
	ops := make([]Operation, 0)
	writeIndices := permutation[0 : config.writePerTxn]
	for i := 0; i < len(keySet); i++ {
		key := keySet[i]
		if inSlice(writeIndices, i) {
			ops = append(ops, Operation{Key: key, Rwflag: WRITE, Value: generateRandomString(config.keyLength)})
		} else {
			ops = append(ops, Operation{Key: key, Rwflag: READ, Value: ""})
		}
	}
	return Transaction{
		ops: ops,
	}
}

func retrieveRandomSetOfKeys(keysPool []string, config TestTxnsConfig) []string {
	keySet := make([]string, 0)
	permutation := rand.Perm(len(keysPool))
	indices := permutation[0 : (config.readPerTxn + config.writePerTxn)]
	for i := 0; i < len(indices); i++ {
		randInd := indices[i]
		keySet = append(keySet, keysPool[randInd])
	}
	return keySet
}

func generateRandomKeyPools(config TestTxnsConfig) []string {
	numKeys := int(math.Max(float64(config.numKeysInPool), float64(config.readPerTxn + config.writePerTxn)))
	keysPool := make([]string, 0)
	for i := 0; i < numKeys; i++ {
		key := generateRandomString(config.keyLength)
		keysPool = append(keysPool, key)
	}
	return keysPool
}

func generateRandomString(stringlen int) string {
	var letterRunes = []rune("0123456789abcdefghijklmnopqrstuvwxyz")
	b := make([]rune, stringlen)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}