package lockserver

import (
	"sort"
	"time"
	"github.com/robfig/cron/v3"
)

type LeaseManager struct {
	lm *LockManager
}

func (leaseM *LeaseManager) Release(clientId string) error {
	rwFlagSet := make(map[string]RWFlag)
	keysToRelease := make([]string, 0)
	leaseM.lm.clientLocksMu.Lock()
	for _, key := range leaseM.lm.clientReadLocks[clientId] {
		rwFlagSet[key] = READ
		keysToRelease = append(keysToRelease, key)
	}
	for _, key := range leaseM.lm.clientWriteLocks[clientId] {
		rwFlagSet[key] = WRITE
		keysToRelease = append(keysToRelease, key)
	}
	leaseM.lm.clientLocksMu.Unlock()

	sort.Strings(keysToRelease)
	reverse(keysToRelease)

	for _, key := range keysToRelease {
		request := LockRequest{
			key: key,
			rwflag: rwFlagSet[key],
			clientId: clientId,
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
		if lease + EXPIRATION_SECS < curTime {
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