[![Build docker images](https://github.com/allinbits/demeris-backend/actions/workflows/docker-build.yml/badge.svg)](https://github.com/allinbits/demeris-backend/actions/workflows/docker-build.yml)

# Emeris Backend

This is the entry-point project for the Emeris backend system.

## Intro to Emeris

The following blog posts give a good, high-level overview of the product's vision. 

* [What is Emeris](https://medium.com/emeris-blog/introducing-emeris-the-cross-chain-portal-to-all-crypto-apps-4e6eee5b53a8)
* [Why Emeris matters](https://blog.cosmos.network/why-emeris-matters-to-cosmos-f8f1dfc7664f)

## Architecture

The Emeris backend can be summarised as *a multi-blockchain indexer*. 

![Emeris backend architecture](./images/architecture.png)  
> Original diagram is [here](https://whimsical.com/backend-current-CP9C1GXs79j9CNs8XAnWJb)

## Components

* [api-server](https://github.com/allinbits/demeris-api-server)
* [cns-server](https://github.com/allinbits/emeris-cns-server)
* [price-oracle](https://github.com/allinbits/emeris-price-oracle)
* [rpc-watcher](https://github.com/allinbits/emeris-rpcwatcher)
* [trace-listener](https://github.com/allinbits/tracelistener/)
* [ticket-watcher](https://github.com/allinbits/emeris-ticket-watcher)
* [sdk-service](https://github.com/allinbits/sdk-service-meta)
* [models](https://github.com/allinbits/demeris-backend-models) (shared library)
* [utils](./utils) (shared library)

## CI/CD 

The Github actions to deploy to various envs are in the [.github](.github/workflows) subfolder.

The CI/CD workflows are described in the following diagram

![Emeris CI/CD](./images/CI_CD.png)
Original diagram is [here](https://whimsical.com/ci-cd-HTBa2HjDzroKsePps71hHE)

Each push/merge to a service's `main` branch (which passes testing), creates a Docker image.  
This in turn triggers an automatic deployment to the DEV env. 

## Local Kubernetes environment

### Requirements

* kubectl
* docker (docker desktop will probably install kubectl)
* helm
* kind

### Usage

Run the script to check how to use it.

```shell
$ ./local-env.sh
Manage demeris local environment

Usage:
  ./local-env.sh [command]

Available Commands:
  up 		 Setup the development environment
  down 		 Tear down the development environment
  connect-sql 	 Connect to database using cockroach built-in SQL Client

Flags:
  -p, --port 	 The local port at which the api will be served
  -n, --cluster-name 	 Kind cluster name
  -b, --build 		 Whether to (re)build docker images
  -h, --help 		 Show this menu
  -m, --monitoring   Setup monitoring infrastructure
```

For more instructions see [this page](https://www.notion.so/allinbits/Emeris-back-end-Dev-environment-setup-2b8a05f940274b45b0b3ba775f1fd6f8#ef44b157a985426d9d9743b5d017e86c).

### Grafana credentials

When monitoring is enabled, Grafana is installed with default credentials and will ask for a password change on first setup. Find below the default credentials

Username: admin

Password: admin
