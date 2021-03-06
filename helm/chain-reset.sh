#!/bin/bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
RESET_DIR="chain-reset"

CHAIN=""
SDK_VERSION="0.42"
ENVIRONMENT="staging"
TRACELISTENER_VERSION="main"

usage()
{
    echo -e "Bulk-import Tracelistener data for a chain\n"
    echo -e "Usage: \n  $0 [flags]\n"
    echo -e "\nFlags:"
    echo -e "  -c, --chain \t\t The chain name (e.g. rizon, cosmos-hub)"
    echo -e "  -s, --sdk \t\t The SDK version of the chain (e.g. 0.42, 0.44), defaults to 0.42"
    echo -e "  -e, --env \t\t Environment name, defaults to staging"
    echo -e "  -t, --tracelistener \t\t Tracelistener docker image version (e.g. v1.0.0, v1.1.0), defaults to main"
    echo -e "  --erase-tables \t Database tables to be cleared, defaults to all"
    echo -e "  --import-modules \t Modules to be bulk imported, defaults to all"
    echo -e "  -h, --help \t\t Show this menu\n"
    exit 1
}

red=`tput setaf 1`
green=`tput setaf 2`
reset=`tput sgr0`

assert_executable_exists()
{
    if ! command -v $1 &> /dev/null
    then
        echo -e "${red}Error:${reset} $1 could not be found. Please install it and re-run this script."
        exit
    fi
}

POSITIONAL=()
while [[ $# -gt 0 ]]
do
key="$1"

case $key in
    -c|--chain)
    CHAIN="$2"
    shift
    shift
    ;;
    -s|--sdk)
    SDK_VERSION="$2"
    shift
    shift
    ;;
    -e|--env)
    ENVIRONMENT="$2"
    shift
    shift
    ;;
    -t|--tracelistener)
    TRACELISTENER_VERSION="$2"
    shift
    shift
    ;;
    --import-modules)
    IMPORT_MODULES="$2"
    shift
    shift
    ;;
    --erase-tables)
    ERASE_TABLES="$2"
    shift
    shift
    ;;
    -h|--help)
    usage
    shift
    ;;
    *)
    POSITIONAL+=("$1")
    shift
    ;;
esac
done
set -- "${POSITIONAL[@]}"

assert_executable_exists helm

if [[ ! "$CHAIN" ]]
then
    echo -e "${red}Error:${reset} chain name is required\n"
    usage
fi

YAML_FILE="${SCRIPT_DIR}/../ci/${ENVIRONMENT}/nodesets/${CHAIN}.yaml"

# replace tracelistener docker version in nodeset
sed -i '/gcr.io\/tendermint-dev\/emeris-tracelistener/ s/:/:'${TRACELISTENER_VERSION}'/2; s/'${TRACELISTENER_VERSION}'.*/'${TRACELISTENER_VERSION}'/' $YAML_FILE

escape_commas() { echo "${1//,/\\,}"; }
ERASE_TABLES="$(escape_commas $ERASE_TABLES)"
IMPORT_MODULES="$(escape_commas $IMPORT_MODULES)"

echo "-- Launcing bulk import job\n"
helm install "${CHAIN}" \
  --set sdkVersion="${SDK_VERSION}",traceListenerVersion="${TRACELISTENER_VERSION}" \
  --set-file nodesetFile="${YAML_FILE}" \
  --set eraseTables="${ERASE_TABLES}" \
  --set importModules="${IMPORT_MODULES}" \
  --namespace emeris \
  "${SCRIPT_DIR}/${RESET_DIR}"

echo -e "-- You can monitor the progress with 'kubectl get jobs'\n"

echo "-- Once chain nodes are fully synced (3/3), do not forget to helm uninstall ${CHAIN}"
