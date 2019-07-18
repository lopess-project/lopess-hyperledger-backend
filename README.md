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
└── README.md
```

where:

    *  bin/             contains executable files provided by the hyperledger framework.
    *  chaincode/       contains the chaincode of the app in order to store and process data.
    *  crypto-config/   contains the output of the bin/cryptogen executable.
    *  deployment/      contains all .yaml files for deploying the required components as docker container.
    *  fabric-config/   contains the basic configuration necessary for the bin/cryptogen and bin/configtxgen executables.
    *  network-config/  contains the output of the bin/configtxgen executable.


# Setup and Installation Guide

This project contains configuration files and runnables in order to create a network consisting of two organizations. Each organisation has the following components which will be deployed when following the next steps:

*  2 OrdererNodes
*  2 PeerNodes
*  1 CA
*  1 CLI

Note: The installation will only work with images newer > v1.4.1 since raft is utilied as consensus algorithm. All required crypto material and channel config files have been generated beforehand and only need to be adapted when necessary.

Note2: Before deploying, make sure your gopath is set accordingly. Moreover, the library of the go implementation of the ed25519 signing algorithm needs to be included to your local gopath accordingly.

For deploying this network, prepare two hosts and create a docker swarm overlay network. How this is done, can be found out by following the link at the bottom of this page. 
After setting up the swarm, adjust network specific parameters within the docker-compose deployment files like the host names of the swarm machines as well as directory paths for mapping the volumes. If this is done, the images can be deployed:

Host 1: 
* clone this gitlab repo
* $ docker stack deploy --compose-file deployment/docker-compose-raft-org1.yaml net
* $ docker stack deploy --compose-file deployment/docker-compose-peer-org1.yaml net
* $ docker stack deploy --compose-file deployment/docker-compose-cli-org1.yaml net


Host 2: 
* clone this gitlab repo
* $ docker stack deploy --compose-file deployment/docker-compose-raft-org2.yaml net
* $ docker stack deploy --compose-file deployment/docker-compose-peer-org2.yaml net
* $ docker stack deploy --compose-file deployment/docker-compose-cli-org2.yaml net

Afterwards check on each machine if the containers are up and running. If so, the following steps need to be performed:

*  Create the channel (channel name is scka-channel btw) and let all the peers join
*  Update Anchor Peers
*  Install and instantiate the chaincode in order to interact
*  Invoke and query

# Create & Join Channel

* Host 1:

`$ docker exec -it <cli_org1_containername> bash`
`$ peer channel create -o orderer0.ordererOrg1.example.com:7050 -c scka-channel -f ./network-config/channel.tx --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/ordererOrg1.example.com/orderers/orderer0.ordererOrg1.example.com/msp/tlscacerts/tlsca.ordererOrg1.example.com-cert.pem`
`$ peer channel join -b scka-channel.block`

The second command will output a channel config block named scka-channel.block , which needs to be send to the other peers in order to join the networks. 

To send this block to the peer on the same host, simply change environment variables within the cli by adding them before the actual command:

`$ CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp CORE_PEER_ADDRESS=peer1.org1.example.com:7051 CORE_PEER_LOCALMSPID="Org1MSP" CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer1.org1.example.com/tls/ca.crt peer channel join -b scka-channel.block`

Now the channel block needs to be transfered to host 2. Therefore first copy it from your docker volume to your local machine:

`$ docker cp <peer0 org1 container name>:/opt/gopath/src/github.com/hyperledger/fabric/peer/scka-channel.block . `

From there transfer it to the second host via scp:

`$ scp scka-channel.block user@host2:<project-dir>`

* Host 2:

Join peer0 from org2:

`$ docker cp scka-channel.block <peer0 org2 container name>:/opt/gopath/src/github.com/hyperledger/fabric/peer/scka-channel.block`
`$ docker exec -it <cli_org2_containername>  bash`
`$ peer channel join -b scka-channel.block`

Join peer1 from org2:

`$ CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp CORE_PEER_ADDRESS=peer1.org2.example.com:7051 CORE_PEER_LOCALMSPID="Org2MSP" CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer1.org2.example.com/tls/ca.crt peer channel join -b scka-channel.block`

# Update Anchor Peers

* Host 1

`$ peer channel update -o orderer0.ordererOrg1.example.com:7050 -c scka-channel -f ./network-config/Org1MSPanchors.tx --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/ordererOrg1.example.com/orderers/orderer0.ordererOrg1.example.com/msp/tlscacerts/tlsca.ordererOrg1.example.com-cert.pem`


* Host 2

`$ peer channel update -o orderer0.ordererOrg2.example.com:7050 -c scka-channel -f ./network-config/Org2MSPanchors.tx --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/ordererOrg2.example.com/orderers/orderer0.ordererOrg2.example.com/msp/tlscacerts/tlsca.ordererOrg2.example.com-cert.pem`

# Install and instantiate chaincode

Install the chaincode on every peer of every org/host. Therefore proceed similar as previously:

* Host 1

When the cli was closed and reentered, proceed as follows. Otherwise do it the otherway round and change peer0 und peer1 accordingly.

`$ peer chaincode install -n mycc -v 1.0 -p github.com/chaincode`
`$ CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp CORE_PEER_ADDRESS=peer1.org1.example.com:7051 CORE_PEER_LOCALMSPID="Org1MSP" CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer1.org1.example.com/tls/ca.crt peer chaincode install -n mycc -v 1.0 -p github.com/chaincode`

* Host 2

`$ peer chaincode install -n mycc -v 1.0 -p github.com/chaincode`
`$ CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp CORE_PEER_ADDRESS=peer1.org2.example.com:7051 CORE_PEER_LOCALMSPID="Org2MSP" CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer1.org2.example.com/tls/ca.crt peer chaincode install -n mycc -v 1.0 -p github.com/chaincode`

Now that the chaincode is installed on every node, we need to instantiate it once. See below.

* Host 1

`$ peer chaincode instantiate -o orderer0.ordererOrg1.example.com:7050 --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/ordererOrg1.example.com/orderers/orderer0.ordererOrg1.example.com/msp/tlscacerts/tlsca.ordererOrg1.example.com-cert.pem -C scka-channel -n mycc -v 1.0 -c '{"Args":[]}' --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt`

There is also the option of defining a specific endorsement policy for the channel. Therefore, simply add '-P "AND ('Org1MSP.peer','Org2MSP.peer')"' as an argument. This policy defines that a transaction needs to be endorsed by at least one peer of org1 AND one peer of org2. Default is the OR operator. When this is done, we can invoke and query transactions.

# Invoke and query chaincode

* Host 2

`$ peer chaincode query -C scka-channel -n mycc -c '{"Args":["getMeasurementRecords"]}'`

* Host 1

`$ peer chaincode invoke -o orderer0.ordererOrg1.example.com:7050 --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/ordererOrg1.example.com/orderers/orderer0.ordererOrg1.example.com/msp/tlscacerts/tlsca.ordererOrg1.example.com-cert.pem -C scka-channel -n mycc --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses peer0.org2.example.com:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt -c '{"function":"initLedger","Args":[]}'`
`$ peer chaincode invoke -o orderer0.ordererOrg1.example.com:7050 --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/ordererOrg1.example.com/orderers/orderer0.ordererOrg1.example.com/msp/tlscacerts/tlsca.ordererOrg1.example.com-cert.pem -C scka-channel -n mycc --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses peer0.org2.example.com:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt -c '{"function":"registerMeasurement","Args":["qgABoAFwhAQ+AVjJqBFIAIQAAtcAOACmAq4AMDYwMDE1MDQ5MDA1OTU2ME4wMDA4MjU1MTY2MEU=","3xRcKIUbrdehGRbcZOWfFd01z6n6yge3Zsnl9J/L2plU2HiAhAw2RQdG91nmEYsvN005oyf1wy1PaRH09v4oAQ=="]}'`

* Host 2

`$ peer chaincode query -C scka-channel -n mycc -c '{"Args":["getMeasurementRecords"]}'`


# Helpful Tutorials

How the setup is done on a single node env can be read here:

[https://hyperledger-fabric.readthedocs.io/en/release-1.4/build_network.html#start-the-network](url)

How this can be done on multiple hosts can be read here:

[https://medium.com/coinmonks/hyperledger-fabric-cluster-on-multiple-hosts-af093f00436](url)

How to setup docker swarm:

[https://medium.com/@malliksarvepalli/hyperledger-fabric-on-multiple-hosts-using-docker-swarm-and-compose-f4b70c64fa7d](url)

