package utils

import (
	"fmt"
	"sync"
)

type KeysLock struct {
	lockMapLock sync.RWMutex
	lockMap map[string]*sync.RWMutex
}

func (kl *KeysLock) lock(key string) {
	kl.lockMapLock.Lock()
	_, exists := kl.lockMap[key]
	if !exists {
		kl.lockMap[key] = &sync.RWMutex{}
	}
	keyGranularityLock, _ := kl.lockMap[key]
	kl.lockMapLock.Unlock()
	keyGranularityLock.Lock()
}

func (kl *KeysLock) unlock(key string) {
	kl.lockMapLock.Lock()
	keyGranularityLock, exists := kl.lockMap[key]
	if !exists {
		panicStr, _ := fmt.Printf("KeysLock trying to unlock key [%s] that does not exist", key)
		panic(panicStr)
	}
	kl.lockMapLock.Unlock()
	keyGranularityLock.Unlock()
}