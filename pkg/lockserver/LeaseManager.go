package lockserver

import (
	"math"
	"sharded_lock_service/pkg/types"
	"sharded_lock_service/pkg/utils"
	"sort"
	"strconv"
	"time"
)

type WrappedDeadlineQueue struct {
	wrappedSlice []*DeadlineInfo
}

type DeadlineInfo struct {
	deadline int64
	clientId string
}

type LeaseManager struct {
	lm *LockManager
	clientLocks       *utils.KeysLock // client id -> lock
	clientReadLeases *utils.ConcurrentStringSliceMap // client id -> [keys]
	clientWriteLeases *utils.ConcurrentStringSliceMap // client id -> [keys]
	expirationTimestampSet *utils.ConcurrentStringMap // client id -> key

	deadlineQueues []*WrappedDeadlineQueue
}

func (leaseM *LeaseManager) deadQIndex(cliendId string) int {
	return int(utils.Hash(cliendId)) % len(leaseM.deadlineQueues)
}
func (leaseM *LeaseManager) PushDeadlineInfo(deadline int64, clientId string) {
	if utils.LEASE_MANAGER_DISABLED {
		return
	}
	i := leaseM.deadQIndex(clientId)
	leaseM.deadlineQueues[i].wrappedSlice = append(leaseM.deadlineQueues[i].wrappedSlice, &DeadlineInfo{
		deadline: deadline,
		clientId: clientId,
	})
}

func (leaseM *LeaseManager) PopDeadlineInfo(workerIndex int) *DeadlineInfo {
	if len(leaseM.deadlineQueues[workerIndex].wrappedSlice) == 0 {
		return nil
	}
	head := leaseM.deadlineQueues[workerIndex].wrappedSlice[0]
	leaseM.deadlineQueues[workerIndex].wrappedSlice = leaseM.deadlineQueues[workerIndex].wrappedSlice[1:]
	return head
}

func (leaseM *LeaseManager) ShouldCleanUpDeadlineQueue(workerIndex int) bool {
	if len(leaseM.deadlineQueues[workerIndex].wrappedSlice) == 0 {
		return false
	}
	deadline := leaseM.deadlineQueues[workerIndex].wrappedSlice[0].deadline
	return deadline > time.Now().Unix()
}

func (leaseM *LeaseManager) getExpireTimestamp(clientId string) int64 {
	leaseM.clientLocks.Lock(clientId)
	valueString := leaseM.expirationTimestampSet.Get(clientId)
	leaseM.clientLocks.Unlock(clientId)
	value, _ := strconv.ParseInt(valueString, 10, 64)
	return value
}

func (leaseM *LeaseManager) setExpireTimestamp(clientId string, value int64) {
	leaseM.clientLocks.Lock(clientId)
	leaseM.expirationTimestampSet.Set(clientId, strconv.FormatInt(value, 10))
	leaseM.clientLocks.Unlock(clientId)
}

func (leaseM *LeaseManager) CleanExpireLocks(workerIndex int) {
	for leaseM.ShouldCleanUpDeadlineQueue(workerIndex) {
		// concurrent cleaning
		clientId := leaseM.PopDeadlineInfo(workerIndex).clientId
		curTime := time.Now().Unix()
		curExp := leaseM.getExpireTimestamp(clientId)
		if curExp > curTime  {
			// the deadline info is outdated, lease has been refreshed
			leaseM.PushDeadlineInfo(curExp, clientId)
			continue
		}
		leaseM.setExpireTimestamp(clientId, math.MaxInt64)
		readLease := make([]string, 0)
		if leaseM.clientReadLeases.Exists(clientId) {
			readLease = leaseM.clientReadLeases.Get(clientId)
		}
		readLeaseCopy := make([]string, len(readLease))
		copy(readLeaseCopy, readLease)

		writeLease := make([]string, 0)
		if leaseM.clientWriteLeases.Exists(clientId) {
			writeLease = leaseM.clientWriteLeases.Get(clientId)
		}
		writeLeaseCopy := make([]string, len(writeLease))
		copy(writeLeaseCopy, writeLease)

		leaseM.clientReadLeases.Delete(clientId)
		leaseM.clientWriteLeases.Delete(clientId)

		// concurrent release
		go func() {
			rwFlagSet := make(map[string]types.RWFlag)
			for _, key := range readLeaseCopy {
				rwFlagSet[key] = types.READ
			}
			for _, key := range writeLeaseCopy {
				rwFlagSet[key] = types.WRITE
			}

			keysToRelease := append(readLeaseCopy, writeLeaseCopy...)
			sort.Strings(keysToRelease)
			utils.SliceReverse(keysToRelease)

			for _, key := range keysToRelease {
				request := types.LockRequest{
					Key: key,
					Rwflag: rwFlagSet[key],
					ClientId: clientId,
				}
				leaseM.lm.processReleaseRequest(request)
			}
		}()
	}
}

func (leaseM *LeaseManager) ClientLeaseSet(clientId string, readKeys []string, writeKeys []string) {
	leaseM.clientLocks.Lock(clientId)
	for _, key := range readKeys {
		leaseM.clientReadLeases.Append(clientId, key)
	}
	for _, key := range writeKeys {
		leaseM.clientWriteLeases.Append(clientId, key)
	}
	leaseM.clientLocks.Unlock(clientId)
}

func (leaseM *LeaseManager) ClientLeaseRemove(clientId string, readKeys []string, writeKeys []string) {
	leaseM.clientLocks.Lock(clientId)
	for _, key := range readKeys {
		leaseM.clientReadLeases.Remove(clientId, key)
	}
	for _, key := range writeKeys {
		leaseM.clientWriteLeases.Remove(clientId, key)
	}
	leaseM.clientLocks.Unlock(clientId)
}

func (leaseM *LeaseManager) LeaseExtend(clientId string) int64 {
	if utils.LEASE_MANAGER_DISABLED {
		return 0
	}
	leaseM.clientLocks.Lock(clientId)
	newExp := time.Now().Unix() + utils.EXPIRATION_SECS
	leaseM.expirationTimestampSet.Set(clientId, strconv.FormatInt(newExp, 10))
	leaseM.clientLocks.Unlock(clientId)
	return newExp
}

func (leaseM *LeaseManager) Serve() {
	if utils.LEASE_MANAGER_DISABLED {
		return
	}
	for i := 0; i < utils.DEADLINE_QUEUE_NUM; i++ {
		go func(workerIndex int) {
			for {
				time.Sleep(time.Duration(utils.LEASE_CHECK_CRON_SECS) * time.Second)
				leaseM.CleanExpireLocks(workerIndex)
			}
		}(i)
	}
}

func NewLeaseManager(lm *LockManager) *LeaseManager {
	leaseM := &LeaseManager{
		lm: lm,
		clientLocks: utils.NewKeysLock(),
		clientReadLeases: utils.NewConcurrentStringSliceMap(),
		clientWriteLeases: utils.NewConcurrentStringSliceMap(),
		expirationTimestampSet: utils.NewConcurrentStringMap(),
		deadlineQueues: make([]*WrappedDeadlineQueue, utils.DEADLINE_QUEUE_NUM),
	}
	for i := 0; i < utils.DEADLINE_QUEUE_NUM ; i++ {
		leaseM.deadlineQueues[i] = &WrappedDeadlineQueue{
			wrappedSlice: make([]*DeadlineInfo, 0),
		}
	}
	return leaseM
}