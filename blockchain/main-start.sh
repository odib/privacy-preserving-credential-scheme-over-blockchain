#!/bin/bash
#
# Copyright SystemX. All Rights Reserved.
#
# Omar DIB
#

#Make Some configuration
function doConfig {
   set -e
   SDIR=$(dirname "$0")
   updateEnv "$@";
   source ${SDIR}/scripts/env.sh
   docker_network_name="net_fabric-ca"
   cd ${SDIR}   
}

function updateEnv {
    default_orgs="Org1"; 
    default_peers_by_org=1;

    if [ -z "$1" ]; then
		echo "No custom organisations supplied. Using default Orgs: $default_orgs";
		NEW_ORGS=$default_orgs;
	else
		NEW_ORGS=$1;
		echo "Using custom organisations. $NEW_ORGS";
	fi

    if [ -z "$2" ]; then
		echo "No custom peers by orgnisation supplied. Using default: $default_peers_by_org";
		PEERS_BY_ORG=$default_peers_by_org;
	else
		PEERS_BY_ORG=$2;
		echo "Using custom peers by organisation. $PEERS_BY_ORG";
	fi

    #NEW_ORGS="PSA";
    #PEERS_BY_ORG=2;
    sed_param=s/PEER_ORGS=.*/PEER_ORGS=\"${NEW_ORGS}\"/  
    sed -i "$sed_param" ${SDIR}/scripts/env.sh
    sed_param=s/NUM_PEERS=.*/NUM_PEERS=${PEERS_BY_ORG}/  
    sed -i "$sed_param" ${SDIR}/scripts/env.sh
}

# Delete docker containers
function deleteContainers {
   dockerContainers=$(docker ps -a | awk '$2~/hyperledger/ || $2~/dev-peer/ {print $1}')
   if [ "$dockerContainers" != "" ]; then
      log "Deleting existing docker containers ..."
      docker rm -f $dockerContainers > /dev/null
   fi
}

# Delete chaincode docker images
function deleteImages {
   chaincodeImages=`docker images | grep "^dev-peer" | awk '{print $3}'`
   if [ "$chaincodeImages" != "" ]; then
      log "Removing chaincode docker images ..."
      docker rmi -f $chaincodeImages > /dev/null
   fi
}

#Clean blockchain Data direcoty
function cleanData {
   DDIR=${SDIR}/${DATA}
   if [ -d ${DDIR} ]; then
      log "Cleaning up the data directory from previous run at $DDIR"
      rm -rf ${SDIR}/data
   fi
   mkdir -p ${DDIR}/logs
}

function existsNetwork {
	networkName=$1
	if [ -n "$(docker network ls -q -f name=$networkName)" ]; then
	    return 0 #true
	else
		return 1 #false
	fi
}

function createDockerNet {
	if existsNetwork $docker_network_name; then
		docker network rm $docker_network_name;
	fi
	echo "Creating default Docker network";
	docker network create $docker_network_name;
}



function deployCouchDbsCont {
      local prt=1
      for ORG in $PEER_ORGS; do
         for idp in `seq 1 $NUM_PEERS`; do
            deployCouchDb "$ORG" "$idp" "$prt"  &
         done
         prt=$((prt+1))
      done
     #for idp in `seq 1 $PEERS_BY_ORG`; do deployPeerCont "$idp"  & done
}

# deploy CouchDB 
function deployCouchDb {
   local org=$1;
   local idc=$2;
   local prt=$3;
   local peerPort="6"$prt"83"
   log "Deploying Couch DB starts...."; 
	docker run -d \
		--name "couchdb-"$org"-"$idc \
	   -p $((peerPort+idc)):5984 \
		-e COUCHDB_USER= \
      -e COUCHDB_PASSWORD= \
      --net $docker_network_name \
		hyperledger/fabric-couchdb;
}



#deploy Fabric CA docker container 
function deployFabricCaCont {
   log "Deploying Fabric Ca Docker starts...."; 
	docker run -d \
		--name "rca-org0" \
		-p 7054:7054 \
		-v $(pwd)/scripts:/"scripts" \
		-v $(pwd)/data:/"data" \
		-e FABRIC_CA_SERVER_HOME="/etc/hyperledger/fabric-ca" \
      -e FABRIC_CA_SERVER_CSR_CN="rca-org0" \
      -e FABRIC_CA_SERVER_CSR_HOSTS="rca-org0" \
      -e FABRIC_CA_SERVER_DEBUG="true" \
      -e BOOTSTRAP_USER_PASS="rca-org0-admin:rca-org0-adminpw" \
      -e TARGET_CERTFILE="/data/org0-ca-cert.pem" \
      -e FABRIC_ORGS="$ORGS" \
      --net $docker_network_name \
		hyperledger/fabric-ca \
		/bin/bash -c "/scripts/start-root-ca.sh 2>&1 | tee /data/logs/rca-org0.log";
}

#deploy the setup docker container 
function deploySetupCont {
   log "Deploy Setup docker container";   
   	docker run -d \
		--name "setup" \
		-v $(pwd)/scripts:/"scripts" \
		-v $(pwd)/data:/"data" \
      --net $docker_network_name \
		hyperledger/fabric-ca-tools \
		/bin/bash -c "/scripts/setup-fabric.sh 2>&1 | tee /data/logs/setup.log; sleep 99999";
}

#deploy the ordorer docker container 
function deployOrdorerCont {
   log "Deploy the ordorer docker container";   
   docker run -d \
      --name "orderer1-org0" \
      -p 7050:7050 \
      -v $(pwd)/scripts:/"scripts" \
      -v $(pwd)/data:/"data" \
      -e FABRIC_CA_CLIENT_HOME="/etc/hyperledger/orderer"  \
      -e ENROLLMENT_URL="http://orderer1-org0:orderer1-org0pw@rca-org0:7054"  \
      -e ORDERER_HOME="/etc/hyperledger/orderer" \
      -e ORDERER_HOST="orderer1-org0" \
      -e ORDERER_GENERAL_LISTENADDRESS="0.0.0.0" \
      -e ORDERER_GENERAL_GENESISMETHOD="file" \
      -e ORDERER_GENERAL_GENESISFILE="/data/genesis.block" \
      -e ORDERER_GENERAL_LOCALMSPID="org0MSP" \
      -e ORDERER_GENERAL_LOCALMSPDIR="/etc/hyperledger/orderer/msp" \
      -e ORDERER_GENERAL_LOGLEVEL="debug" \
      -e ORDERER_DEBUG_BROADCASTTRACEDIR="data/logs" \
      -e ORG="$ORGS" \
      -e ORG_ADMIN_CERT="/data/orgs/org0/msp/admincerts/cert.pem" \
      --net $docker_network_name \
      hyperledger/fabric-ca-orderer \
      /bin/bash -c "/scripts/start-orderer.sh 2>&1 | tee /data/logs/orderer1-org0.log";
}

   
function deployPeersCont {
      local prt=1
      for ORG in $PEER_ORGS; do
         for idp in `seq 1 $NUM_PEERS`; do
            deployPeerCont "$ORG" "$idp" "$prt"  &
         done
         prt=$((prt+1))
      done
     #for idp in `seq 1 $PEERS_BY_ORG`; do deployPeerCont "$idp"  & done
}

#deploy the peers docker containers 
function deployPeerCont {
   local org=$1;
   local idc=$2;
   local prt=$3;
   local peerPort="7"$prt"50"
   local eventPort="8"$prt"50"
   log "Deploy peer"${idc}"-"${org}" container"; 
   docker run -d \
      --name "peer"${idc}"-"${org} \
      -p $((peerPort+idc)):7051 \
      -p $((eventPort+idc)):7053 \
      -v $(pwd)/scripts:/"scripts" \
      -v $(pwd)/data:/"data" \
      -v /var/run:"/host/var/run"  \
      -e FABRIC_CA_CLIENT_HOME="/opt/gopath/src/github.com/hyperledger/fabric/peer" \
      -e ENROLLMENT_URL="http://peer"${idc}"-"${org}":peer"${idc}"-"${org}"pw@rca-org0:7054" \
      -e PEER_NAME="peer"${idc}"-"${org} \
      -e PEER_HOME="/opt/gopath/src/github.com/hyperledger/fabric/peer" \
      -e PEER_HOST="peer"${idc}"-"${org} \
      -e PEER_NAME_PASS="peer"${idc}"-"${org}":peer"${idc}"-"${org}"pw" \
      -e CORE_PEER_ID="peer"${idc}"-"${org} \
      -e CORE_PEER_ADDRESS="peer"${idc}"-"${org}":7051" \
      -e CORE_PEER_LOCALMSPID=${org}"MSP" \
      -e CORE_PEER_MSPCONFIGPATH="/opt/gopath/src/github.com/hyperledger/fabric/peer/msp" \
      -e CORE_VM_ENDPOINT="unix:///host/var/run/docker.sock" \
      -e CORE_VM_DOCKER_HOSTCONFIG_NETWORKMODE="net_fabric-ca" \
      -e CORE_LOGGING_LEVEL="DEBUG" \
      -e CORE_PEER_GOSSIP_USELEADERELECTION="true" \
      -e CORE_PEER_GOSSIP_ORGLEADER="false" \
      -e CORE_PEER_GOSSIP_EXTERNALENDPOINT="peer"${idc}"-"${org}":7051" \
      -e CORE_PEER_GOSSIP_SKIPHANDSHAKE="true" \
      -e ORG="$ORGS" \
      -e ORG_ADMIN_CERT="/data/orgs/"${org}"/msp/admincerts/cert.pem" \
      -e CORE_LEDGER_STATE_STATEDATABASE=CouchDB \
      -e CORE_LEDGER_STATE_COUCHDBCONFIG_COUCHDBADDRESS="couchdb-"$org"-"$idc:5984 \
      -e CORE_LEDGER_STATE_COUCHDBCONFIG_USERNAME= \
      -e CORE_LEDGER_STATE_COUCHDBCONFIG_PASSWORD= \
      --workdir "/opt/gopath/src/github.com/hyperledger/fabric/peer"  \
      --net $docker_network_name \
      hyperledger/fabric-ca-peer \
      /bin/bash -c "/scripts/start-peer.sh 2>&1 | tee /data/logs/peer"${idc}"-"${org}".log";
}

#deploy the run docker containers 
function deployRunCont {
   log "Deploy the run docker container";   
   
	docker run -d \
		--name "run" \
		-v $(pwd)/scripts:/"scripts" \
		-v $(pwd)/data:/"data" \
		-v $(pwd)/chaincode:/"opt/gopath/src/github.com/hyperledger/fabric-samples/chaincode" \
		-v "/src/github.com/hyperledger/fabric":/"opt/gopath/src/github.com/hyperledger/fabric" \
    -v $(pwd)/utilities:/"opt/gopath/src" \
		-e GOPATH="/opt/gopath" \
		--net $docker_network_name \
		hyperledger/fabric-ca-tools \
		/bin/bash -c "sleep 3;/scripts/start-fabric.sh 2>&1 | tee /data/logs/run.log; sleep 99999";
}

# Create the docker containers
function deployComponents {
   log "Creating docker containers ..."
   deployCouchDbsCont;
   sudo chmod -R 777 ${SDIR}/${DATA}
   sleep 5
   deployFabricCaCont;
   sleep 5
   deploySetupCont;
   sleep 5
   deployOrdorerCont;
   sleep 5
   deployPeersCont;
   sleep 5
   deployRunCont;
}


#Wait for every thing to be ok
function waitTillComplete {
   # Wait for the setup container to complete
   dowait "the 'setup' container to finish registering identities, creating the genesis block and other artifacts" 90 $SDIR/$SETUP_LOGFILE $SDIR/$SETUP_SUCCESS_FILE
   
   # Wait for the run container to start and then tails it's summary log
   dowait "the docker 'run' container to start" 60 ${SDIR}/${SETUP_LOGFILE} ${SDIR}/${RUN_SUMFILE}
   tail -f ${SDIR}/${RUN_SUMFILE}&
   TAIL_PID=$!
   
   # Wait for the run container to complete
   while true; do 
      if [ -f ${SDIR}/${RUN_SUCCESS_FILE} ]; then
         kill -9 $TAIL_PID
         break
         #exit 0
      elif [ -f ${SDIR}/${RUN_FAIL_FILE} ]; then
         kill -9 $TAIL_PID
         break
         #exit 1
      else
         sleep 1
      fi
   done
}

function removeOld {
   deleteContainers;
   deleteImages;
   cleanData;
}

# Create the docker-compose file
#${SDIR}/makeDocker.sh

function main {
   doConfig "$@";
   removeOld;
   createDockerNet;
   deployComponents;
   sleep 10
   waitTillComplete;
}

#Pass arguments to function exactly as-is
main "$@"