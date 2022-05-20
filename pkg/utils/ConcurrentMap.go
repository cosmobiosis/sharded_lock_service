package utils

import "sync"

type SliceValue struct {
	wrappedSlice []string
}

type StringValue struct {
	wrappedStr string
}

type ConcurrentSliceMap struct {
	mapLock sync.RWMutex
	ValueMap map[string]*SliceValue
}

type ConcurrentStringMap struct {
	mapLock sync.RWMutex
	ValueMap map[string]*StringValue
}

func (cm *ConcurrentSliceMap) exists(key string) bool {
	cm.mapLock.RLock()
	_, exists := cm.ValueMap[key]
	cm.mapLock.RUnlock()
	return exists
}

func (cm *ConcurrentSliceMap) append(key string, value string) {
	// need keysLock
	cm.mapLock.Lock()
	_, exists := cm.ValueMap[key]
	if !exists {
		cm.ValueMap[key] = &SliceValue{
			wrappedSlice: make([]string, 0),
		}
	}
	cm.mapLock.Unlock()
	cm.ValueMap[key].wrappedSlice = append(cm.ValueMap[key].wrappedSlice, value)
}

func (cm *ConcurrentSliceMap) popHead(key string) string {
	// need keysLock
	head := cm.ValueMap[key].wrappedSlice[0]
	cm.ValueMap[key].wrappedSlice = cm.ValueMap[key].wrappedSlice[1:]
	return head
}

func (cm *ConcurrentSliceMap) empty(key string) bool {
	// need keysLock
	cm.mapLock.RLock()
	_, exists := cm.ValueMap[key]
	if !exists {
		cm.mapLock.RUnlock()
		return true
	}
	cm.mapLock.RUnlock()
	return len(cm.ValueMap[key].wrappedSlice) == 0
}

func (cm *ConcurrentSliceMap) head(key string) string {
	// need keysLock
	return cm.ValueMap[key].wrappedSlice[0]
}

func (cm *ConcurrentSliceMap) delete(key string) {
	cm.mapLock.Lock()
	delete(cm.ValueMap, key)
	cm.mapLock.Unlock()
}

func (cm *ConcurrentStringMap) exists(key string) bool {
	cm.mapLock.RLock()
	_, exists := cm.ValueMap[key]
	cm.mapLock.RUnlock()
	return exists
}

func (cm *ConcurrentStringMap) delete(key string) {
	cm.mapLock.Lock()
	delete(cm.ValueMap, key)
	cm.mapLock.Unlock()
}

func (cm *ConcurrentStringMap) set(key string, value string) {
	// need keysLock
	cm.ValueMap[key].wrappedStr = value
}

func (cm *ConcurrentStringMap) get(key string) string {
	// need keysLock
	return cm.ValueMap[key].wrappedStr
}

func NewConcurrentSliceMap() *ConcurrentSliceMap {
	return &ConcurrentSliceMap{
		ValueMap: make(map[string]*SliceValue),
	}
}

func NewConcurrentStringMap() *ConcurrentStringMap {
	return &ConcurrentStringMap{
		ValueMap: make(map[string]*StringValue),
	}
}