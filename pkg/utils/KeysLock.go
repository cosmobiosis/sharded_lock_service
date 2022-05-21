package utils

import (
	"fmt"
	"sync"
)

type KeysLock struct {
	lockMapLock sync.RWMutex
	lockMap map[string]*sync.RWMutex
}

func (kl *KeysLock) Lock(key string) {
	kl.lockMapLock.Lock()
	_, exists := kl.lockMap[key]
	if !exists {
		kl.lockMap[key] = &sync.RWMutex{}
	}
	kl.lockMapLock.Unlock()
	keyGranularityLock, _ := kl.lockMap[key]
	keyGranularityLock.Lock()
}

func (kl *KeysLock) Unlock(key string) {
	kl.lockMapLock.Lock()
	keyGranularityLock, exists := kl.lockMap[key]
	if !exists {
		panicStr, _ := fmt.Printf("KeysLock trying to unlock key [%s] that does not exist", key)
		panic(panicStr)
	}
	kl.lockMapLock.Unlock()
	keyGranularityLock.Unlock()
}

func NewKeysLock() *KeysLock {
	return &KeysLock{
		lockMap: make(map[string]*sync.RWMutex),
	}
}