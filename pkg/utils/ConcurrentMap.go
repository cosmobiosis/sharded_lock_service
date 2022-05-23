package utils

import (
	"sharded_lock_service/pkg/types"
	"sync"
)

type SliceLockRequest struct {
	wrappedSlice []types.LockRequest
}

type ConcurrentLockRequestSliceMap struct {
	mapLock sync.RWMutex
	ValueMap map[string]*SliceLockRequest
}

func (cm *ConcurrentLockRequestSliceMap) Exists(key string) bool {
	cm.mapLock.RLock()
	_, exists := cm.ValueMap[key]
	cm.mapLock.RUnlock()
	return exists
}

func (cm *ConcurrentLockRequestSliceMap) Append(key string, value types.LockRequest) {
	cm.mapLock.Lock()
	_, exists := cm.ValueMap[key]
	if !exists {
		cm.ValueMap[key] = &SliceLockRequest{
			wrappedSlice: make([]types.LockRequest, 0),
		}
	}
	cm.mapLock.Unlock()
	cm.ValueMap[key].wrappedSlice = append(cm.ValueMap[key].wrappedSlice, value)
}

func (cm *ConcurrentLockRequestSliceMap) PushHead(key string, value types.LockRequest) {
	cm.ValueMap[key].wrappedSlice = append([]types.LockRequest{value}, cm.ValueMap[key].wrappedSlice...)
}

func (cm *ConcurrentLockRequestSliceMap) PopHead(key string) types.LockRequest {
	head := cm.ValueMap[key].wrappedSlice[0]
	cm.ValueMap[key].wrappedSlice = cm.ValueMap[key].wrappedSlice[1:]
	return head
}

func (cm *ConcurrentLockRequestSliceMap) Get(key string) []types.LockRequest {
	return cm.ValueMap[key].wrappedSlice
}

func (cm *ConcurrentLockRequestSliceMap) Empty(key string) bool {
	return len(cm.ValueMap[key].wrappedSlice) == 0
}

func (cm *ConcurrentLockRequestSliceMap) Head(key string) types.LockRequest {
	return cm.ValueMap[key].wrappedSlice[0]
}

func (cm *ConcurrentLockRequestSliceMap) Delete(key string) {
	cm.mapLock.Lock()
	delete(cm.ValueMap, key)
	cm.mapLock.Unlock()
}

func NewConcurrentLockRequestSliceMap() *ConcurrentLockRequestSliceMap {
	return &ConcurrentLockRequestSliceMap {
		ValueMap: make(map[string]*SliceLockRequest),
	}
}

type SliceStringValue struct {
	wrappedSlice []string
}

type ConcurrentStringSliceMap struct {
	mapLock sync.RWMutex
	ValueMap map[string]*SliceStringValue
}

func (cm *ConcurrentStringSliceMap) Exists(key string) bool {
	cm.mapLock.RLock()
	_, exists := cm.ValueMap[key]
	cm.mapLock.RUnlock()
	return exists
}

func (cm *ConcurrentStringSliceMap) Contains(key string, value string) bool {
	if !cm.Exists(key) {
		return false
	}
	return SliceContains(cm.Get(key), value)
}

func (cm *ConcurrentStringSliceMap) Get(key string) []string {
	return cm.ValueMap[key].wrappedSlice
}

func (cm *ConcurrentStringSliceMap) Remove(key string, value string) {
	cm.ValueMap[key].wrappedSlice = SliceRemove(cm.ValueMap[key].wrappedSlice, value)
}

func (cm *ConcurrentStringSliceMap) Append(key string, value string) {
	cm.mapLock.Lock()
	_, exists := cm.ValueMap[key]
	if !exists {
		cm.ValueMap[key] = &SliceStringValue{
			wrappedSlice: make([]string, 0),
		}
	}
	cm.mapLock.Unlock()
	cm.ValueMap[key].wrappedSlice = append(cm.ValueMap[key].wrappedSlice, value)
}

func (cm *ConcurrentStringSliceMap) PopHead(key string) string {
	head := cm.ValueMap[key].wrappedSlice[0]
	cm.ValueMap[key].wrappedSlice = cm.ValueMap[key].wrappedSlice[1:]
	return head
}

func (cm *ConcurrentStringSliceMap) Empty(key string) bool {
	return len(cm.ValueMap[key].wrappedSlice) == 0
}

func (cm *ConcurrentStringSliceMap) Head(key string) string {
	return cm.ValueMap[key].wrappedSlice[0]
}

func (cm *ConcurrentStringSliceMap) Delete(key string) {
	cm.mapLock.Lock()
	delete(cm.ValueMap, key)
	cm.mapLock.Unlock()
}

func NewConcurrentStringSliceMap() *ConcurrentStringSliceMap {
	return &ConcurrentStringSliceMap{
		ValueMap: make(map[string]*SliceStringValue),
	}
}

type StringValue struct {
	wrappedStr string
}

type ConcurrentStringMap struct {
	mapLock sync.RWMutex
	ValueMap map[string]*StringValue
}

func (cm *ConcurrentStringMap) Exists(key string) bool {
	cm.mapLock.RLock()
	_, exists := cm.ValueMap[key]
	cm.mapLock.RUnlock()
	return exists
}

func (cm *ConcurrentStringMap) Delete(key string) {
	cm.mapLock.Lock()
	delete(cm.ValueMap, key)
	cm.mapLock.Unlock()
}

func (cm *ConcurrentStringMap) Set(key string, value string) {
	cm.ValueMap[key].wrappedStr = value
}

func (cm *ConcurrentStringMap) Get(key string) string {
	return cm.ValueMap[key].wrappedStr
}

func NewConcurrentStringMap() *ConcurrentStringMap {
	return &ConcurrentStringMap{
		ValueMap: make(map[string]*StringValue),
	}
}