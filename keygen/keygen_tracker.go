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
	expiryTime := time.Now().Add(-10 * time.Minute)
	t.keygens.Range(func(key, value interface{}) bool {
		id := key.(common.ADKGID)
		createdAt := value.(time.Time)
		if createdAt.Before(expiryTime) {
			t.Remove(id)
		}
		return true
	})
}

func (t *KeygenTracker) Add(id common.ADKGID) {
	t.keygens.Store(id, time.Now())
}

func (t *KeygenTracker) Has(id common.ADKGID) bool {
	_, ok := t.keygens.Load(id)
	return ok
}

func (t *KeygenTracker) Remove(id common.ADKGID) {
	if _, ok := t.keygens.LoadAndDelete(id); !ok {
		return
	}
	t.cleanupFunc(id)
}
