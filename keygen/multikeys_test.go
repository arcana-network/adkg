package keygen

import (
	"math/big"
	"os"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"sync"
	"testing"
	"time"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/keygen/message_handlers/acss"
	log "github.com/sirupsen/logrus"
)

var normalNodesNum = 7
var crushedNodesNum = 0
var keyNums = 50

func TestMultiKey(t *testing.T) {
	timeout := time.After(300 * time.Second)
	done := make(chan bool)

	log.SetLevel(log.WarnLevel)
	runtime.GOMAXPROCS(20)

	nodes, transport := setupNodes(normalNodesNum, crushedNodesNum)

	for i := 1; i <= keyNums; i++ {
		// go func(index int) {
		id := common.GenerateADKGID(*big.NewInt(int64(i)))
		for _, n := range nodes {
			t.Logf("key id: %d", i)
			go func(node *Node) {
				round := common.RoundDetails{
					ADKGID: id,
					Dealer: node.ID(),
					Kind:   "acss",
				}
				msg, err := acss.NewShareMessage(
					round.ID(),
					common.SECP256K1,
				)
				if err != nil {
					log.WithError(err).Error("EndBlock:Acss.NewShareMessage")
				}
				node.ReceiveMessage(node.Details(), *msg)
			}(n)
		}
		// }(i)
	}

	go func() {
		countMap := make(map[string]int)
		keys := make(map[string]struct{})
		lock := &sync.Mutex{}
		for {
			output := <-transport.output

			key := output
			lock.Lock()
			countMap[key]++
			if countMap[key] >= f-1 {
				_, ok := keys[key]
				if !ok {
					keys[key] = struct{}{}
					log.Infof("Key: %s; length: %d", key, len(keys))
					if len(keys) == keyNums {
						t.Logf("OutputArray: %s", keys)
						debug.FreeOSMemory()
						runtime.GC()
						f, err := os.Create("heap.out")
						if err != nil {
							log.Errorf("Could not create heap.out: %s", err)
							return
						}
						defer f.Close()
						gf, err := os.Create("goroutine.out")
						if err != nil {
							log.Errorf("Could not create goroutine.out: %s", err)
							return
						}
						defer gf.Close()
						err = pprof.Lookup("goroutine").WriteTo(gf, 0)
						if err != nil {
							log.Errorf("Could not write heap profile: %s", err)
						}
						err = pprof.WriteHeapProfile(f)
						if err != nil {
							log.Errorf("Could not write heap profile: %s", err)
						}
						done <- true
					}
				}
			}
			lock.Unlock()
		}
	}()

	select {
	case <-timeout:
		t.Fatal("Test didn't finish in time")
	case <-done:
	}
}
