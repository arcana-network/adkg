package keygen

import (
	"sync"
	"time"

	"github.com/arcana-network/dkgnode/common"
)

type KeygenTracker struct {
	keygens     sync.Map
	cleanupFunc func(id common.ADKGID)
}

const CUTOFF_PERIOD = -5 * time.Minute

func NewKeygenTracker(cleanupFunc func(id common.ADKGID)) *KeygenTracker {
	t := &KeygenTracker{
		cleanupFunc: cleanupFunc,
	}
	go t.StartJanitor()
	return t
}

func (t *KeygenTracker) StartJanitor() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		t.removeExpiredKeygen()
	}
}

func (t *KeygenTracker) removeExpiredKeygen() {
	cutoffTime := time.Now().Add(CUTOFF_PERIOD).Unix()
	t.keygens.Range(func(key, value interface{}) bool {
		id := key.(common.ADKGID)
		createdAt := value.(int64)
		if createdAt < cutoffTime {
			t.keygens.Delete(id)
			t.cleanupFunc(id)
		}
		return true
	})
}

func (t *KeygenTracker) Add(id common.ADKGID) {
	t.keygens.Store(id, time.Now().Unix())
}

func (t *KeygenTracker) Has(id common.ADKGID) bool {
	_, ok := t.keygens.Load(id)
	return ok
}
