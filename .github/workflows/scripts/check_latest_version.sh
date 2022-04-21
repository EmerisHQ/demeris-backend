#!/bin/bash

# Color Variables 

RED='\033[0;31m'          # Red
GREEN='\033[0;32m'        # Green
NC='\033[0m'              # No Color


################################
# Crescent 
################################

echo "Checking Crescent Version"

CRESCENT_DEPLOYED_VERSION=$(kubectl get nodeset -n emeris | grep "crescent" | awk '{print $4}')
CRESCENT_LATEST_VERSION=$(curl https://api.github.com/repos/crescent-network/crescent/releases/latest -s | jq .name -r)

if [[ "${CRESCENT_DEPLOYED_VERSION}" == "${CRESCENT_LATEST_VERSION}" ]]
then
  echo -e "${GREEN}Crescent is already running on latest version${NC}"
else
  echo -e "${RED}Crescent needs to be upgraded to ${CRESCENT_LATEST_VERSION} version${NC}"
fi

################################
# Osmosis 
################################

echo "Checking Osmosis Version"

OSMOSIS_DEPLOYED_VERSION=$(kubectl get nodeset -n emeris | grep "osmosis" | awk '{print $4}')
OSMOSIS_LATEST_VERSION=$(curl https://api.github.com/repos/osmosis-labs/osmosis/releases/latest -s | jq .name -r)

if [[ "${OSMOSIS_DEPLOYED_VERSION}" == "${OSMOSIS_LATEST_VERSION}" ]]
then
  echo -e "${GREEN}Osmosis is already running on latest version${NC}"
else
  echo -e "${RED}Osmosis needs to be upgraded to ${OSMOSIS_LATEST_VERSION} version${NC}"
fi
