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

