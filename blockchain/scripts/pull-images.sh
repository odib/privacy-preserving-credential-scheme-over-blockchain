#!/bin/bash
#
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

 export CA_TAG=1.2.1
 export CA_TOOLS_TAG=1.2.1
 export CA_PEER_TAG=1.2.1 
 export CA_ORDERER_TAG=1.2.1

# current version of thirdparty images (couchdb, kafka and zookeeper) released
export THIRDPARTY_IMAGE_VERSION=0.4.14

#export CA_TOOLS_TAG="x86_64-1.1.0"

# ensure we're in the fabric-samples directory
dir=`basename $PWD`
if [ "${dir}" == "scripts" ]; then
        cd ..
fi

dir=`basename $PWD`
if [ "${dir}" != "dib-fabric-samples" ]; then
	echo "You should run this script from the fabric-samples root directory."
	exit 1
fi

dockerCaPull() {
      echo "==> FABRIC CA IMAGE"
      echo
      docker pull hyperledger/fabric-ca:$CA_TAG
      docker tag hyperledger/fabric-ca:$CA_TAG hyperledger/fabric-ca
}

dockerCaToolsPull() {
      echo "==> FABRIC CA Tools IMAGE"
      echo
      docker pull hyperledger/fabric-ca-tools:$CA_TOOLS_TAG
      docker tag hyperledger/fabric-ca-tools:$CA_TOOLS_TAG hyperledger/fabric-ca-tools
}

dockerCaOrdererPull() {
      echo "==> FABRIC CA Orderer IMAGE"
      echo
      docker pull hyperledger/fabric-ca-orderer:$CA_ORDERER_TAG
      docker tag hyperledger/fabric-ca-orderer:$CA_ORDERER_TAG hyperledger/fabric-ca-orderer
}

dockerCaPeerPull() {
      echo "==> FABRIC CA Peer IMAGE"
      echo
      docker pull hyperledger/fabric-ca-peer:$CA_PEER_TAG
      docker tag hyperledger/fabric-ca-peer:$CA_PEER_TAG hyperledger/fabric-ca-peer
}


dockerThirdPartyImagesPull() {
  for IMAGES in couchdb kafka zookeeper; do
      echo "==> THIRDPARTY DOCKER IMAGE: $IMAGES"
      echo
      docker pull hyperledger/fabric-$IMAGES:$THIRDPARTY_IMAGE_VERSION
      docker tag hyperledger/fabric-$IMAGES:$THIRDPARTY_IMAGE_VERSION hyperledger/fabric-$IMAGES
  done
}


dockerInstall() {
  which docker >& /dev/null
  NODOCKER=$?
  if [ "${NODOCKER}" == 0 ]; then
	  echo "===> Pulling fabric ca Image"
	  dockerCaPull;
	  echo "===> Pulling fabric ca tools Image"
	  dockerCaToolsPull;
	  echo "===> Pulling fabric ca Orderer Image"
	  dockerCaOrdererPull;
	  echo "===> Pulling fabric ca Peer Image"
	  dockerCaPeerPull;
	  echo "===> Pulling thirdparty docker images"
	  dockerThirdPartyImagesPull;
	  echo
	  echo "===> List out hyperledger docker images"
	  docker images | grep hyperledger*
  else
    echo "========================================================="
    echo "Docker not installed, bypassing download of Fabric images"
    echo "========================================================="
  fi
}


echo
echo "Installing Hyperledger Fabric docker images"
echo
dockerInstall

