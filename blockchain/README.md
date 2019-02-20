# Blockchain: Hyperledger Fabric Deployer

This Blockchain module aims at automatically deploying a Hyperledger Fabric Blockchain. 
The number of organizations and peers can be automatically adjusted. 
By default, one single organization, one peer and a single orderer are considered. 

Along with the deployment, this module is used to implement an on-chain anonymous attribute verification. 
The code for this can be found in the method `verify` inside the `aav.go` chaincode



## Requirements 

To run and manipulate this code, the following elements have to be installed
- A PC with Ubuntu >= 16.04
- Docker (tested with version 19.03.2, build 6a30dfc )
- docker-compose (tested with version 1.24.1, build 4667896b)


## How to run this code 

To run this code, do the following
1) in a terminal, Enter the folder `blockchain` and Execute the file `main-start.sh`

## How to stop the blockchain 

To stop the blockchain, do the following
1) in a terminal, Enter the folder `blockchain` and Execute the file `main-stop.sh`

This will automatically stop all blockchain docker containers, and remove the blockchain data. 


## Contributors

- Omar DIB | IRT SystemX (https://www.irt-systemx.fr/en/)
- Clément HUYART | ERCOM (https://www.ercom.fr/)

