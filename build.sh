#!/bin/bash
eval $(cat $1 | sed 's/^/export /')

go build -o bin/dkg -ldflags "-X 'github.com/arcana-network/dkgnode/versioning.Version=$VERSION' \
-X 'github.com/arcana-network/dkgnode/config.DefaultGatewayURL=$GATEWAY_URL'" 
