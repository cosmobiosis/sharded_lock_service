package lockserver

import (
	"math"
	"sharded_lock_service/pkg/types"
	"sharded_lock_service/pkg/utils"
	"sort"
	"strconv"
	"time"
)

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

	deadlineQueue []*DeadlineInfo
}

func (leaseM *LeaseManager) PushDeadlineInfo(deadline int64, clientId string) {
	if utils.LEASE_MANAGER_DISABLED {
		return
	}
	leaseM.deadlineQueue = append(leaseM.deadlineQueue, &DeadlineInfo{
		deadline: deadline,
		clientId: clientId,
	})
}

func (leaseM *LeaseManager) PopDeadlineInfo() *DeadlineInfo {
	if len(leaseM.deadlineQueue) == 0 {
		return nil
	}
	head := leaseM.deadlineQueue[0]
	leaseM.deadlineQueue = leaseM.deadlineQueue[1:]
	return head
}

func (leaseM *LeaseManager) ShouldCleanUpDeadlineQueue() bool {
	if len(leaseM.deadlineQueue) == 0 {
		return false
	}
	deadline := leaseM.deadlineQueue[0].deadline
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

func (leaseM *LeaseManager) CleanExpireLocks() {
	for leaseM.ShouldCleanUpDeadlineQueue() {
		// concurrent cleaning
		clientId := leaseM.PopDeadlineInfo().clientId
		curTime := time.Now().Unix()
		if leaseM.getExpireTimestamp(clientId) > curTime  {
			// the deadline info is outdated, lease has been refreshed
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
	for {
		time.Sleep(time.Duration(utils.LEASE_CHECK_CRON_SECS) * time.Second)
		leaseM.CleanExpireLocks()
	}
}

func NewLeaseManager(lm *LockManager) *LeaseManager {
	return &LeaseManager{
		lm: lm,
		clientLocks: utils.NewKeysLock(),
		clientReadLeases: utils.NewConcurrentStringSliceMap(),
		clientWriteLeases: utils.NewConcurrentStringSliceMap(),
		expirationTimestampSet: utils.NewConcurrentStringMap(),
	}
}