package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/arcana-network/dkgnode/cmd/root"
)

func main() {
	log.SetLevel(log.DebugLevel)
	err := root.GetRootCmd().Execute()
	if err != nil {
		log.Fatalf("Could not start node %s", err.Error())
	}
}
