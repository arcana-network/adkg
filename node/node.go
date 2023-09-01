package node

import (
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/arcana-network/dkgnode/cache"
	"github.com/arcana-network/dkgnode/chain"
	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/config"
	"github.com/arcana-network/dkgnode/db"
	"github.com/arcana-network/dkgnode/eventbus"
	"github.com/arcana-network/dkgnode/keygen"
	"github.com/arcana-network/dkgnode/keystore"
	"github.com/arcana-network/dkgnode/p2p"
	"github.com/arcana-network/dkgnode/server"
	"github.com/arcana-network/dkgnode/tendermint"
	"github.com/arcana-network/dkgnode/verifier"
)

func Start(conf *config.Config) {

	config.GlobalConfig = conf

	log.SetLevel(log.InfoLevel)
	bus := eventbus.New()

	serviceRegistry := common.NewServiceRegistry(bus)
	serviceRegistry.SetupMethodRouting()

	services := []common.IService{
		chain.New(bus),
		p2p.New(bus),
		keygen.New(bus),
		db.New(),
		tendermint.NewCore(bus),
		tendermint.NewABCI(bus),
		cache.New(),
		server.New(bus),
		verifier.New(bus),
		keystore.New(bus),
	}

	for _, s := range services {
		err := serviceRegistry.RegisterService(s)
		if err != nil {
			log.Fatalf("Error while registering service=%s, err=%s", s.ID(), err)
		}
	}

	err := serviceRegistry.StartAll()
	if err != nil {
		log.Fatalf("Error while starting all services: err=%s", err)
	}

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			debug.FreeOSMemory()
		}
	}()

	stopOnInterrupt(serviceRegistry)
}

func stopOnInterrupt(serviceRegistry *common.ServiceRegistry) {
	osSignal := make(chan os.Signal, 1)
	signal.Notify(osSignal, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	osSig := <-osSignal
	log.Println("Termination started, signal: " + osSig.String())
	err := serviceRegistry.StopAll()
	if err != nil {
		log.Fatalf("Error while stopping all services: err=%s", err)
	}
}
