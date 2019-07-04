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
*  Create the channel (channel name is scka-channel btw) and let all the peers join
*  Update Anchor Peers
*  Install and instantiate the chaincode in order to interact

# Create & Join Channel

* Host 1:

`$ docker exec -it cli bash`
`$ peer channel create -o orderer0.org1.example.com:7050 -c scka-channel -f ./network-config/channel.tx --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/org1.example.com/orderers/orderer0.org1.example.com/msp/tlscacerts/tlsca.org1.example.com-cert.pem`
`$ peer channel join -b scka-channel.block`

The second command will output a channel config block named scka-channel.block , which needs to be send to the other peers in order to join the networks. 

To send this block to the peer on the same host, simply change environment variables within the cli by adding them before the actual command:

`CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp CORE_PEER_ADDRESS=peer1.org1.example.com:8051 CORE_PEER_LOCALMSPID="Org1MSP" CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer1.org1.example.com/tls/ca.crt peer channel join -b scka-channel.block`

Now the channel block needs to be transfered to host 2. Therefore first copy it from your docker volume to your local machine:

`docker cp peer0.org1.example.com:/opt/gopath/src/github.com/hyperledger/fabric/peer/scka-channel.block . `

From there transfer it to the second host via scp:

`scp scka-channel.block user@host2:<project-dir>`

* Host 2:

Join peer0 from org2:

`$ docker cp scka-channel.block peer0.org2.example.com:/opt/gopath/src/github.com/hyperledger/fabric/peer/scka-channel.block`
`$ docker exec -it cli bash`
`$ peer channel join -b scka-channel.block`

Join peer1 from org2:

`CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp CORE_PEER_ADDRESS=peer1.org2.example.com:8051 CORE_PEER_LOCALMSPID="Org2MSP" CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer1.org2.example.com/tls/ca.crt peer channel join -b scka-channel.block`

# Update Anchor Peers

(still within the cli of peer1)

* Host 1

`peer channel update -o orderer0.org1.example.com:7050 -c scka-channel -f ./network-config/Org1MSPanchors.tx --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/org1.example.com/orderers/orderer0.org1.example.com/msp/tlscacerts/tlsca.org1.example.com-cert.pem`


* Host 2

`peer channel update -o orderer0.org2.example.com:7050 -c scka-channel -f ./network-config/Org2MSPanchors.tx --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/org2.example.com/orderers/orderer0.org2.example.com/msp/tlscacerts/tlsca.org2.example.com-cert.pem`

# Install and instantiate chaincode



# Helpful Tutorials

How the setup is done on a single node env can be read here:

[https://hyperledger-fabric.readthedocs.io/en/release-1.4/build_network.html#start-the-network](url)

How this can be done on multiple hosts can be read here:

[https://medium.com/coinmonks/hyperledger-fabric-cluster-on-multiple-hosts-af093f00436](url)

