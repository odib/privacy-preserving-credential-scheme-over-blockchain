#!/bin/bash
#
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

set -e
SDIR=$(dirname "$0")
cd ${SDIR}
docker_network_name="net_fabric-ca"

# Delete docker containers
function deleteContainers {
   dockerContainers=$(docker ps -a | awk ' $2~/hyperledger/ || $2~/dev-peer/ {print $1}');
    if [ "$dockerContainers" != "" ]; then
        log "Deleting existing docker containers ...";
        docker rm -f $dockerContainers > /dev/null;
    fi
}

# Delete chaincode docker images
function deleteImages {
   chaincodeImages=`docker images | grep "^dev-peer" | awk '{print $3}'`
   if [ "$chaincodeImages" != "" ]; then
      log "Removing chaincode docker images ...";
      docker rmi -f $chaincodeImages > /dev/null;
   fi
}

#Clean blockchain Data direcoty
function cleanData {
   DDIR=${SDIR}/${DATA};
   if [ -d ${DDIR} ]; then
      log "Cleaning up the data directory from previous run at $DDIR";
      sudo rm -rf ${SDIR}/data;
   fi
   mkdir -p ${DDIR}/logs;
}

function existsNetwork {
	networkName=$1
	if [ -n "$(docker network ls -q -f name=$networkName)" ]; then
	    return 0 #true
	else
		return 1 #false
	fi
}

function deleteDockerNet {
    echo "Deleting  Docker network ...";
	if existsNetwork $docker_network_name; then
		docker network rm $docker_network_name;
	fi
}


# log a message
function log {
   if [ "$1" = "-n" ]; then
      shift
      echo -n "##### `date '+%Y-%m-%d %H:%M:%S'` $*"
   else
      echo "##### `date '+%Y-%m-%d %H:%M:%S'` $*"
   fi
}

function main {
   deleteContainers;
   deleteImages;
   deleteDockerNet;
   cleanData;
}

main