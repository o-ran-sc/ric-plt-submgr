# Subscription Manager

## Table of contents
* [Introduction](#introduction)
* [Release notes](#release-notes)
* [Prerequisites](#prerequisites)
* [Project folders structure](#project-folders-structure)
* [Installation guide](#installation-guide)
  * [Compiling code](#compiling-code)
  * [Building docker container](#building-docker-container)
  * [Installing Routing Manager](#installing-routing-manager)
  * [Testing and Troubleshoting](#testing-and-troubleshoting)
* [Upcoming changes](#upcoming-changes)
* [License](#license)

## Introduction
__Subscription Manager__ is a basic platform service of RIC. It is responsible to serve, coordinate and manage xApps' subscriptions.

Submgr acts as an anchor point for subscription related internal messaging, i.e. every xApp sends its subscription related messages to Submgr. Submgr invokes Routing Manager (Rtmgr) to create or tear down the subscription related routes, and the appropriate E2 Termination to signal the subscription related event also towards the RAN.

The solution base on the [xapp-frame](https://gerrit.o-ran-sc.org/r/admin/repos/ric-plt/xapp-frame) project which provides common HttpREST, RMR and SDL interfaces.

Current implementation provides the following functionalities:
* Handling RIC_SUB_REQ and RIC_SUB_RESP type RMR messages 
* Generating New subscription ID and forwarding subscription request to E2 termination
* Receiving Subscription response and sendig it back to the subscriber
  
## Release notes
Check the separated `RELNOTES` file.

## Prerequisites
* Healthy kubernetes cluster
* Access to the common docker registry

## Project folder structure
* /build: contains build tools (scripts, Dockerfiles, etc.)
* /manifest: contains deployment files (Kubernetes manifests, Helm chart)
* /cmd: contains go project's main file
* /pkg: contains go project's internal packages
* /config: contains default configuration file for the service
* /test: contains CI/CD testing files (scripts, mocks, manifests)

## Installation guide

### Compiling code
Enter the project root and execute `./build.sh` script.
The build script has two main phases. First is the code compilation, where it creates a temporary container for downloading all dependencies then compiles the code. In the second phase it builds the production ready container and taggs it to `submgr:builder`

**NOTE:** The script puts a copy of the binary into the `./bin` folder for further use cases

### Installing Subscription Manager
#### Preparing environment
Tag the `submgr` container according to the project release and push it to a registry accessible from all minions of the Kubernetes cluster.
Edit the container image section of `submgr-dep.yaml` file according to the `submgr` image tag.

#### Deploying Subscription Manager 
Issue the `kubectl create -f {manifest.yaml}` command in the following order
  1. `manifests/namespace.yaml`: creates the `example` namespace for routing-manager resources
  2. `manifests/submgr/submgr-dep.yaml`: instantiates the `submgr` deployment in the `example` namespace
  3. `manifests/submgr/submgr-svc.yaml`: creates the `submgr` service in `example` namespace

### Testing and Troubleshoting
Subscription Manager's behaviour can be tested using the stub xApp (called RCO) and the stub E2 Termination (called E2T) on the following way.

  1. [Compile](#compiling-code) and [Installing subscription manager](#installing-subscription-manager)
  2  Enter `./test/dbaas` folder and issue `kubectl apply -f ./manifests`
  3. Enter `./test/e2t/` folder and run `build.sh`. After docker image successfully built, issue `kubectl apply -f ./manifests`
  4. Enter `./test/rco/` folder and run `build.sh`. After docker image successfully built, issue `kubectl apply -f ./manifests`
  5. Configure RMR routes accordingly

Test scenario:
  1. RCO alternately sends Subscription Request (12010) and other (10000) type of messages towards SUBMGR. (non ASN1 code/decode)
  2. SUBMGR receives RCO's subscription request and generates a new ID for the given request and puts it in the header of RMR messages to be forwareded to E2T
  3. E2T receives the Subscription Request message and sends a Subscription Response to SUBMGR
  4. SUBMGR accepts the Subscirption Response and forwards it to RCO


## Configuration and Troubleshooting
Basic configuration file provided in `./config/` folder. Consult xapp-frame project documentation for custom configuration settings.

## Upcoming changes
[] ASN1 support

## License
This project is licensed under the Apache License, Version 2.0 - see the [LICENSE](LICENSE)

