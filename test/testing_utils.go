package test

import (
	"fmt"
	"math/rand"
	"sharded_lock_service/pkg/lockserver"
	"strconv"
	"math"
)

type TestInfo struct {
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

type Transaction struct {
	ops []Operation
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
	keyPool := make([]string, numKeys)
	for i := 0; i < numKeys; i++ {
		key := generateRandomString(config.keyLength)
		keyPool = append(keyPool, key)
	}
	return keyPool
}

func generateRandomString(stringlen int) string {
	var letterRunes = []rune("0123456789abcdefghijklmnopqrstuvwxyz")
	b := make([]rune, stringlen)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}