#!/bin/bash
#
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

set -e

source $(dirname "$0")/env.sh

 CC_NAME=aav #stands for anonymous-attribute-verifier";
 CC_PATH=github.com/hyperledger/fabric-samples/chaincode/$CC_NAME/go
 CC_LANGUAGE=golang
 CC_VERSION=1.0


function main {

   done=false

   # Wait for setup to complete and then wait another 10 seconds for the orderer and peers to start
   #awaitSetup
   #sleep 10

   trap finish EXIT

   mkdir -p $LOGPATH
   logr "The docker 'run' container has started"

   # Set ORDERER_PORT_ARGS to the args needed to communicate with the 1st orderer
   IFS=', ' read -r -a OORGS <<< "$ORDERER_ORGS"
   echo "Start Debug .... "
   echo $OORGS
   echo "End Debug .... "
   initOrdererVars ${OORGS[0]} 1
   export ORDERER_PORT_ARGS="-o $ORDERER_HOST:7050 "

   # Convert PEER_ORGS to an array named PORGS
   IFS=', ' read -r -a PORGS <<< "$PEER_ORGS"

   # Create the channel
   logr "Creating the channel ... "
   createChannel

   # All peers join the channel
   #logr "Let all peers join the created channel ... "
   letPeersJoinChannel
  

   # Update the anchor peers
   #logr "Updating the anchor peers ... "
   updateAnchorPeers

   # Install chaincode on all peers in each org
   #logr "Installing the chaincode on all peers in each org ... "
   installChainCodeOnAllPeers

   #logr "Defining the policy related to the chaincode  ... "
   makePolicy
 
   #logr "Instantiating the chaincode on the channel ... "
   instantiateChainCode

   # Query chaincode from all peers on each Org
   #logr "Querying the chaincode from every peer in each organization  ... "
   queryChainCodeFromAllPeers

   
   # # Invoke chaincode on the 1st peer of the 1st org
   # initPeerVars ${PORGS[0]} 1
   # switchToUserIdentity
   # logr "Sending invoke transaction to $PEER_HOST ..."
   # peer chaincode invoke -C $CHANNEL_NAME -n $CC_NAME -c '{"function":"initLedger","Args":[""]}' $ORDERER_CONN_ARGS
   logr "Congratulations! The tests ran successfully."

   done=true
}

function letPeersJoinChannel {
    for ORG in $PEER_ORGS; do
      local COUNT=1
      while [[ "$COUNT" -le $NUM_PEERS ]]; do
         initPeerVars $ORG $COUNT
         joinChannel
         COUNT=$((COUNT+1))
      done
   done
}

function updateAnchorPeers {
   for ORG in $PEER_ORGS; do
      initPeerVars $ORG 1
      switchToAdminIdentity
      logr "Updating anchor peers for $PEER_HOST ..."
      peer channel update -c $CHANNEL_NAME -f $ANCHOR_TX_FILE $ORDERER_CONN_ARGS
   done
}

#The chaincode must be installed on each endorsing peer node of a channel that will run your chaincode.
function installChainCodeOnAllPeers {
      for ORG in $PEER_ORGS; do
      local COUNT=1
      while [[ "$COUNT" -le $NUM_PEERS ]]; do
         initPeerVars $ORG $COUNT
         installChaincode
         COUNT=$((COUNT+1))
      done
   done
}

function instantiateChainCode {
   local COUNT=1
   for ORG in $PEER_ORGS; do
      initPeerVars $ORG $COUNT
      switchToAdminIdentity
      logr "Instantiating chaincode on $PEER_HOST ..."
      peer chaincode instantiate -C $CHANNEL_NAME -n $CC_NAME -l "$CC_LANGUAGE" -v $CC_VERSION -c '{"Args":["testAavKey","testAavVal"]}' -P "$POLICY" $ORDERER_CONN_ARGS
   done
}

function queryChainCodeFromAllPeers {
    for ORG in $PEER_ORGS; do
       local COUNT=1
       while [[ "$COUNT" -le $NUM_PEERS ]]; do
          initPeerVars $ORG $COUNT
          switchToUserIdentity
          queryChainCode 10
          COUNT=$((COUNT+1))
       done
    done
}

# Enroll as a peer admin and create the channel
function createChannel {
   echo "Debug Create Channel "
   echo $PORGS
   initPeerVars ${PORGS[0]} 1
   switchToAdminIdentity
   logr "Creating channel '$CHANNEL_NAME' on $ORDERER_HOST ..."
   peer channel create --logging-level=DEBUG -c $CHANNEL_NAME -f $CHANNEL_TX_FILE $ORDERER_CONN_ARGS
}

# Enroll as a fabric admin and join the channel
function joinChannel {
   switchToAdminIdentity
   set +e
   local COUNT=1
   MAX_RETRY=10
   while true; do
      logr "Peer $PEER_HOST is attempting to join channel '$CHANNEL_NAME' (attempt #${COUNT}) ..."
      peer channel join -b $CHANNEL_NAME.block
      if [ $? -eq 0 ]; then
         set -e
         logr "Peer $PEER_HOST successfully joined channel '$CHANNEL_NAME'"
         return
      fi
      if [ $COUNT -gt $MAX_RETRY ]; then
         fatalr "Peer $PEER_HOST failed to join channel '$CHANNEL_NAME' in $MAX_RETRY retries"
      fi
      COUNT=$((COUNT+1))
      sleep 1
   done
}

function queryChainCode {
   if [ $# -ne 1 ]; then
      fatalr "Usage: queryChainCode  <expected-value>"
   fi
   set +e
   logr "Querying chaincode in the channel '$CHANNEL_NAME' on the peer '$PEER_HOST' ..."
   local rc=1
   local starttime=$(date +%s)
   # Continue to poll until we get a successful response or reach QUERY_TIMEOUT
   while test "$(($(date +%s)-starttime))" -lt "$QUERY_TIMEOUT"; do
      sleep 1
      peer chaincode query -C $CHANNEL_NAME -n $CC_NAME -c '{"Args":["initAav","testAavKey","testAavVal"]}' >& log.txt 
      logr "Query of channel '$CHANNEL_NAME' on peer '$PEER_HOST' was successful"
      set -e
      return 0
      echo -n "."
   done
   cat log.txt
   cat log.txt >> $RUN_SUMFILE
   logr "Failed to query channel '$CHANNEL_NAME' on peer '$PEER_HOST'; expected value was $1 and found $VALUE"
   fatalr "Failed to query channel '$CHANNEL_NAME' on peer '$PEER_HOST'; expected value was $1 and found $VALUE"
}

function makePolicy  {
   POLICY="OR("
   local COUNT=0
   for ORG in $PEER_ORGS; do
      if [ $COUNT -ne 0 ]; then
         POLICY="${POLICY},"
      fi
      initOrgVars $ORG
      POLICY="${POLICY}'${ORG_MSP_ID}.member'"
      COUNT=$((COUNT+1))
   done
   POLICY="${POLICY})"
   log "policy: $POLICY"
}

function installChaincode {
   switchToAdminIdentity
   logr "Installing chaincode on $PEER_HOST ..."
   peer chaincode install -n $CC_NAME -v $CC_VERSION -p "$CC_PATH" -l "$CC_LANGUAGE"
}

function fetchConfigBlock {
   logr "Fetching the configuration block of the channel '$CHANNEL_NAME'"
   peer channel fetch config $CONFIG_BLOCK_FILE -c $CHANNEL_NAME $ORDERER_CONN_ARGS
}

function updateConfigBlock {
   logr "Updating the configuration block of the channel '$CHANNEL_NAME'"
   peer channel update -f $CONFIG_UPDATE_ENVELOPE_FILE -c $CHANNEL_NAME $ORDERER_CONN_ARGS
}

function createConfigUpdatePayloadWithCRL {
   logr "Creating config update payload with the generated CRL for the organization '$ORG'"
   # Start the configtxlator
   configtxlator start &
   configtxlator_pid=$!
   log "configtxlator_pid:$configtxlator_pid"
   logr "Sleeping 5 seconds for configtxlator to start..."
   sleep 5

   pushd /tmp

   CTLURL=http://127.0.0.1:7059
   # Convert the config block protobuf to JSON
   curl -X POST --data-binary @$CONFIG_BLOCK_FILE $CTLURL/protolator/decode/common.Block > config_block.json
   # Extract the config from the config block
   jq .data.data[0].payload.data.config config_block.json > config.json

   # Update crl in the config json
   crl=$(cat $CORE_PEER_MSPCONFIGPATH/crls/crl*.pem | base64 | tr -d '\n')
   cat config.json | jq '.channel_group.groups.Application.groups.'"${ORG}"'.values.MSP.value.config.revocation_list = ["'"${crl}"'"]' > updated_config.json

   # Create the config diff protobuf
   curl -X POST --data-binary @config.json $CTLURL/protolator/encode/common.Config > config.pb
   curl -X POST --data-binary @updated_config.json $CTLURL/protolator/encode/common.Config > updated_config.pb
   curl -X POST -F original=@config.pb -F updated=@updated_config.pb $CTLURL/configtxlator/compute/update-from-configs -F channel=$CHANNEL_NAME > config_update.pb

   # Convert the config diff protobuf to JSON
   curl -X POST --data-binary @config_update.pb $CTLURL/protolator/decode/common.ConfigUpdate > config_update.json

   # Create envelope protobuf container config diff to be used in the "peer channel update" command to update the channel configuration block
   echo '{"payload":{"header":{"channel_header":{"channel_id":"'"${CHANNEL_NAME}"'", "type":2}},"data":{"config_update":'$(cat config_update.json)'}}}' > config_update_as_envelope.json
   curl -X POST --data-binary @config_update_as_envelope.json $CTLURL/protolator/encode/common.Envelope > $CONFIG_UPDATE_ENVELOPE_FILE

   # Stop configtxlator
   kill $configtxlator_pid

   popd
}

function finish {
   if [ "$done" = true ]; then
      logr "See $RUN_LOGFILE for more details"
      touch /$RUN_SUCCESS_FILE
   else
      logr "Tests did not complete successfully; see $RUN_LOGFILE for more details"
      touch /$RUN_FAIL_FILE
   fi
}

function logr {
   log $*
   log $* >> $RUN_SUMPATH
}

function fatalr {
   logr "FATAL: $*"
   exit 1
}

main

