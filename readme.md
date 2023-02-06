# Keygen Node

> **Warning** *Limited Release*
>
> ADKG protocol is currently available only for trusted partners as a binary file.  We do not recommend building or editing this protocol and deploying a local build copy yet as listed in the developer guide below. 
>
> Trusted partners that are running validator nodes using [the latest binary](https://github.com/arcana-network/adkg/releases) can refer to [Validator Onboarding Guide](https://docs.arcana.network/validator_intro.html) and stay tuned on the special Slack channel for validators. 
>
> *Please report any issues immediately on the channel right away!*

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

The config folders should have `config.local.${1..6}.json` in the `config/` folder for the cluster to start.

The following `keys` are required in the JSON config:

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

#### Prerequisites

- [Docker](https://docs.docker.com/engine/install/)

1.  Clone the repository
```
git clone git@github.com:arcana-network/dkgnode.git
```
2. Create config files for all the nodes and move them to the config directory. Refer to instructions on the confluence page - [local-setup](https://team-1624093970686.atlassian.net/wiki/spaces/AN/pages/196608024/Local+environment+setup+for+developers+WIP). Note: We will be updating this link in the future as it may not be accessible externally, yet.

3. Run the local environment with DKG nodes
```
make run-local
```