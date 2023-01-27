# Keygen Node

## Pre-Requisites

- [Go 1.6+](https://go.dev/doc/install)

Before running the make file, save the following environment variables to `.envrc`:

```sh
touch .envrc
```

```sh
File: .envrc
--------------------------------------------------------------------
export PRIVATE_KEY=
export IP_ADDRESS=
export ETH_CONN=
export CONTRACT_ADDR=
export TLS=
export SERVER_CERT=
export SERVER_KEY=
export SERVER_PORT=
export BASE_PATH=
```

- `SERVER_CERT` and `SERVER_KEY` are optional  

## Commands

### Help

Show list of available make commands

```sh
make help
```

### Upgrade

Upgrades all go dependencies

```sh
make upgrade
```

### Run

Runs the dkg node

```sh
make run
```

### Build

Builds the dkg node

```sh
make build
```

### Lint

Runs linter across the project using golangci-lint

```sh
make lint
```

## Local cluster deployment

The config folders should have `config.local.${1..6}.json` in `config/` folder for cluster to start

The following `keys` are required in the json config:

```json
{
  "privatekey": "",
  "ipAddress": "192.167.10.1_",
  "serverPort": "8000",
  "TLS": false,
  "contractAddress": "",
  "ethConnection": "https://rinkeby.infura.io/v3/"
}
```

### Start cluster

```sh
docker-compose -f docker-compose.yml up -d
```

### Stop

```sh
docker-compose down
```

# Setup DKG nodes on local

## Quick start

#### Prerequisits

- [Docker](https://docs.docker.com/engine/install/)

1.  Clone the repository
```
git clone git@github.com:arcana-network/dkgnode.git
```
2. Create config files for all the six nodes and move to config directory. Refer confluence page - [local-setup](https://team-1624093970686.atlassian.net/wiki/spaces/AN/pages/196608024/Local+environment+setup+for+developers+WIP)

3. Run local environment with DKG nodes
```
make run-local
```