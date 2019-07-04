# HyperledgerFabric

The project is structured as follows:

```bash
.hyperledger
├── bin
├── chaincode
├── crypto-config
├── deployment
├── fabric-config
├── network-config
├── .env
└── README.md
```

where:

    *  bin/             contains executable files provided by the hyperledger framework.
    *  chaincode/       contains the chaincode of the app in order to store and process data.
    *  crypto-config/   contains the output of the bin/cryptogen executable.
    *  deployment/      contains all .yaml files for deploying the required components as docker container.
    *  fabric-config/   contains the basic configuration necessary for the bin/cryptogen and bin/configtxgen executables.
    *  network-config/  contains the output of the bin/configtxgen executable.
    *  .env             defines project-wide config parameters.


# Setup and Installation Guide

This project contains configuration files and runnables in order to create a network consisting of two organizations. Each organisation has the following components which will be deployed when following the next steps:

*  2 OrdererNodes
*  2 PeerNodes
*  1 CA-Node
*  1 CLI

Note: The installation will only work with images newer > v1.4.1 since raft is utilied as consensus algorithm. All required crypto material and channel config files have been generated beforehand and only need to be adapted when necessary.

For deploying this network, prepare two hosts and edit the corresponding parameters within the .env file. If this is done, the images can be deployed:

Host 1: 
* clone this gitlab repo
* $ docker-compose -f deployment/docker-base-raft-org1.yaml up -d
* $ docker-compose -f deployment/docker-base-peer-org1.yaml up -d
* $ docker-compose -f deployment/docker-base-cli-org1.yaml up -d

Host 2: 
* clone this gitlab repo
* $ docker-compose -f deployment/docker-base-raft-org2.yaml up -d
* $ docker-compose -f deployment/docker-base-peer-org2.yaml up -d
* $ docker-compose -f deployment/docker-base-cli-org2.yaml up -d

Afterwards check on each machine if the containers are up and running. If so, the following steps need to be performed:
*  Start the network
*  Create the channel (channel name is scka-channel btw) and let all the peers join
*  Update Anchor Peers
*  Install and instantiate the chaincode in order to interact

How this is done on a single node env can be read here:

[https://hyperledger-fabric.readthedocs.io/en/release-1.4/build_network.html#start-the-network](url)

How this can be done on multiple hosts can be read here:

[https://medium.com/coinmonks/hyperledger-fabric-cluster-on-multiple-hosts-af093f00436](url)

