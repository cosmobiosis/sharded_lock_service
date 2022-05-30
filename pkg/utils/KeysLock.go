package utils

import (
	"fmt"
	"sync"
)

type KeysLock struct {
	lockMapLock sync.RWMutex
	lockMap     map[string]*sync.RWMutex
}

func (kl *KeysLock) Lock(key string) {
	kl.lockMapLock.Lock()
	_, exists := kl.lockMap[key]
	if !exists {
		kl.lockMap[key] = &sync.RWMutex{}
	}
	keyGranularityLock, _ := kl.lockMap[key]
	kl.lockMapLock.Unlock()
	keyGranularityLock.Lock()
}

func (kl *KeysLock) Unlock(key string) {
	kl.lockMapLock.RLock()
	keyGranularityLock, exists := kl.lockMap[key]
	if !exists {
		panicStr, _ := fmt.Printf("KeysLock trying to unlock key [%s] that does not exist", key)
		panic(panicStr)
	}
	kl.lockMapLock.RUnlock()
	keyGranularityLock.Unlock()
}

func NewKeysLock() *KeysLock {
	return &KeysLock{
		lockMap: make(map[string]*sync.RWMutex),
	}
}
