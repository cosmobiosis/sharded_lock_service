package lockserver

import (
	"github.com/robfig/cron/v3"
	"sharded_lock_service/pkg/types"
	"sharded_lock_service/pkg/utils"
	"sort"
	"time"
)

type LeaseManager struct {
	lm *LockManager
}

func (leaseM *LeaseManager) Release(clientId string) error {
	rwFlagSet := make(map[string]types.RWFlag)
	keysToRelease := make([]string, 0)
	leaseM.lm.clientsIdLock.Lock(clientId)
	for _, key := range leaseM.lm.clientReadLocks.Get(clientId) {
		rwFlagSet[key] = types.READ
		keysToRelease = append(keysToRelease, key)
	}
	for _, key := range leaseM.lm.clientWriteLocks.Get(clientId) {
		rwFlagSet[key] = types.WRITE
		keysToRelease = append(keysToRelease, key)
	}
	leaseM.lm.clientsIdLock.Unlock(clientId)

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
	// release lock by reversed sorting sequence
	return nil
}

func (leaseM *LeaseManager) CleanExpireLocks() {
	curTime := time.Now().Unix()
	leaseM.lm.clientLeaseMu.Lock()
	ExpireSet := make(map[string]int64)
	for clientId, lease := range leaseM.lm.clientLease {
		if lease +utils.EXPIRATION_SECS < curTime {
			ExpireSet[clientId] = lease
		}
	}
	leaseM.lm.clientLeaseMu.Unlock()
	for client, _ := range ExpireSet {
		_ = leaseM.Release(client)
	}
}

func (leaseM *LeaseManager) Serve() {
	c := cron.New()
	spec := "*/5 * * * * ?"
	_, _ = c.AddFunc(spec, leaseM.CleanExpireLocks)
	c.Start()

	select{}
}