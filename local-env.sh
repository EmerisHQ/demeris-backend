#!/bin/bash

CLUSTER_NAME=demeris
BUILD=false
STARPORT_OPERATOR_REPO=git@github.com:allinbits/starport-operator.git
PORT=8000

usage()
{
    echo -e "Manage demeris local environment\n"
    echo -e "Usage: \n  $0 [command]\n"
    echo -e "Available Commands:"
    echo -e "  up \t\t Setup the development environment"
    echo -e "  down \t\t Tear down the development environment"
    echo -e "  connect-sql \t Connect to database using cockroach built-in SQL Client"
    echo -e "\nFlags:"
    echo -e "  -p, --port \t The local port at which the api will be served"
    echo -e "  -n, --cluster-name \t Kind cluster name"
    echo -e "  -b, --build \t\t Whether to (re)build docker images"
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
    -n|--cluster-name)
    CLUSTER_NAME="$2"
    shift
    shift
    ;;
    -p|--port)
    PORT="$2"
    shift
    shift
    ;;
    -b|--build)
    BUILD=true
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

COMMAND=${POSITIONAL[0]}

if [ "$COMMAND" = "" ]
then
    usage
fi

if [[ ! "$COMMAND" =~ ^(up|down|connect-sql)$ ]]
then
    echo -e "${red}Error:${reset} command does not exist\n"
    usage
fi

assert_executable_exists kind
assert_executable_exists helm
assert_executable_exists kubectl
assert_executable_exists docker

if [ "$COMMAND" = "up" ]
then
    ### Create the cluster

    if kind get clusters | grep $CLUSTER_NAME &> /dev/null
    then
        echo -e "${green}\xE2\x9C\x94${reset} Cluster $CLUSTER_NAME already exists"
    else
        echo -e "${green}\xE2\x9C\x94${reset} Creating cluster $CLUSTER_NAME"
        kind create cluster --name $CLUSTER_NAME
        kubectl label nodes $CLUSTER_NAME-control-plane ingress-ready=true --context kind-$CLUSTER_NAME
    fi

    ### Ensure nginx ingress controller is deployed

    echo -e "${green}\xE2\x9C\x94${reset} Ensure nginx ingress controller is installed"
    kubectl apply \
        --context kind-$CLUSTER_NAME \
        -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/static/provider/kind/deploy.yaml \
        &> /dev/null

    ### Setup container for proxying localhost:$PORT to nginx
    if [ ! "$(docker ps | grep $CLUSTER_NAME-local-proxy)" ]
    then
        echo -e "${green}\xE2\x9C\x94${reset} Ensure local container for proxying traffic to cluster"
        node_port=$(kubectl get service -n ingress-nginx ingress-nginx-controller -o=jsonpath="{.spec.ports[?(@.port == 80)].nodePort}")
        docker run -d --rm \
            --name $CLUSTER_NAME-local-proxy \
            -p 127.0.0.1:$PORT:80 \
            --network kind \
            --link $CLUSTER_NAME-control-plane:target \
            alpine/socat -dd tcp-listen:80,fork,reuseaddr tcp-connect:target:$node_port
    fi

    ### Ensure starport-operator is deployed

    if [ ! -d .starport-operator/.git ]
    then
        echo -e "${green}\xE2\x9C\x94${reset} Cloning starport-operator repo"
        git clone $STARPORT_OPERATOR_REPO .starport-operator &> /dev/null
    else
        echo -e "${green}\xE2\x9C\x94${reset} Fetching starport-operator latest changes"
        cd .starport-operator
        git pull $STARPORT_OPERATOR_REPO &> /dev/null
        cd ..
    fi

    echo -e "${green}\xE2\x9C\x94${reset} Ensure starport-operator is installed"
    helm upgrade starport-operator \
        --install \
        --create-namespace \
        --kube-context kind-$CLUSTER_NAME \
        --namespace starport-system \
        --set webHooksEnabled=false \
        --set enableAntiAffinity=false \
        .starport-operator/helm/starport-operator \
        &> /dev/null

    ### Ensure cockroach db is installed
    echo -e "${green}\xE2\x9C\x94${reset} Ensure cockroach db is installed and running"
    helm repo add cockroachdb https://charts.cockroachdb.com/ &> /dev/null
    helm upgrade cockroachdb \
        --install \
        --kube-context kind-$CLUSTER_NAME \
        --set tls.enabled=false \
        --set config.single-node=true \
        --set statefulset.replicas=1 \
        cockroachdb/cockroachdb \
        &> /dev/null

    ### Ensure tracelistener image
    if [[ "$(docker images -q demeris/tracelistener 2> /dev/null)" == "" ]]
    then
        echo -e "${green}\xE2\x9C\x94${reset} Building demeris/tracelistener image"
        docker build -t demeris/tracelistener -f Dockerfile.tracelistener .
    else
        if [ "$BUILD" = "true" ]
        then
            echo -e "${green}\xE2\x9C\x94${reset} Re-building demeris/tracelistener image"
            docker build -t demeris/tracelistener -f Dockerfile.tracelistener .
        else
            echo -e "${green}\xE2\x9C\x94${reset} Image demeris/tracelistener already exists"
        fi
    fi
    echo -e "${green}\xE2\x9C\x94${reset} Pushing demeris/tracelistener image to cluster"
    kind load docker-image demeris/tracelistener --name $CLUSTER_NAME &> /dev/null

    ### Setup chains
    echo -e "${green}\xE2\x9C\x94${reset} Create/update chains"
    kubectl apply \
        --context kind-$CLUSTER_NAME \
        -f local-env/nodes

    ### Ensure cns-server image
    if [[ "$(docker images -q demeris/cns-server 2> /dev/null)" == "" ]]
    then
        echo -e "${green}\xE2\x9C\x94${reset} Building demeris/cns-server image"
        docker build -t demeris/cns-server -f Dockerfile.cns-server .
    else
        if [ "$BUILD" = "true" ]
        then
            echo -e "${green}\xE2\x9C\x94${reset} Re-building demeris/cns-server image"
            docker build -t demeris/cns-server -f Dockerfile.cns-server .
        else
            echo -e "${green}\xE2\x9C\x94${reset} Image demeris/cns-server already exists"
        fi
    fi
    echo -e "${green}\xE2\x9C\x94${reset} Pushing demeris/cns-server image to cluster"
    kind load docker-image demeris/cns-server --name $CLUSTER_NAME &> /dev/null

    echo -e "${green}\xE2\x9C\x94${reset} Deploying demeris/cns-server"
    helm upgrade cns-server \
        --install \
        --kube-context kind-$CLUSTER_NAME \
        --set imagePullPolicy=Never \
        helm/demeris-cns-server \
        &> /dev/null

    ### Ensure api-server image
    if [[ "$(docker images -q demeris/api-server 2> /dev/null)" == "" ]]
    then
        echo -e "${green}\xE2\x9C\x94${reset} Building demeris/api-server image"
        docker build -t demeris/api-server -f Dockerfile.api-server .
    else
        if [ "$BUILD" = "true" ]
        then
            echo -e "${green}\xE2\x9C\x94${reset} Re-building demeris/api-server image"
            docker build -t demeris/api-server -f Dockerfile.api-server .
        else
            echo -e "${green}\xE2\x9C\x94${reset} Image demeris/api-server already exists"
        fi
    fi
    echo -e "${green}\xE2\x9C\x94${reset} Pushing demeris/api-server image to cluster"
    kind load docker-image demeris/api-server --name $CLUSTER_NAME &> /dev/null

    echo -e "${green}\xE2\x9C\x94${reset} Deploying demeris/api-server"
    helm upgrade api-server \
        --install \
        --kube-context kind-$CLUSTER_NAME \
        --set imagePullPolicy=Never \
        helm/demeris-api-server \
        &> /dev/null

    # ### Ensure price-oracle-server image
    # if [[ "$(docker images -q demeris/price-oracle-server 2> /dev/null)" == "" ]]
    # then
    #     echo -e "${green}\xE2\x9C\x94${reset} Building demeris/price-oracle-server image"
    #     docker build -t demeris/price-oracle-server -f Dockerfile.price-oracle .
    # else
    #     if [ "$BUILD" = "true" ]
    #     then
    #         echo -e "${green}\xE2\x9C\x94${reset} Re-building demeris/price-oracle-server image"
    #         docker build -t demeris/price-oracle-server -f Dockerfile.price-oracle .
    #     else
    #         echo -e "${green}\xE2\x9C\x94${reset} Image demeris/price-oracle-server already exists"
    #     fi
    # fi
    # echo -e "${green}\xE2\x9C\x94${reset} Pushing demeris/price-oracle-server image to cluster"
    # kind load docker-image demeris/price-oracle-server --name $CLUSTER_NAME &> /dev/null

    # ### Ensure tmwsproxy image
    # if [[ "$(docker images -q demeris/tmwsproxy 2> /dev/null)" == "" ]]
    # then
    #     echo -e "${green}\xE2\x9C\x94${reset} Building demeris/tmwsproxy image"
    #     docker build -t demeris/tmwsproxy -f Dockerfile.tmwsproxy .
    # else
    #     if [ "$BUILD" = "true" ]
    #     then
    #         echo -e "${green}\xE2\x9C\x94${reset} Re-building demeris/tmwsproxy image"
    #         docker build -t demeris/tmwsproxy -f Dockerfile.tmwsproxy .
    #     else
    #         echo -e "${green}\xE2\x9C\x94${reset} Image demeris/tmwsproxy already exists"
    #     fi
    # fi
    # echo -e "${green}\xE2\x9C\x94${reset} Pushing demeris/tmwsproxy image to cluster"
    # kind load docker-image demeris/tmwsproxy --name $CLUSTER_NAME &> /dev/null

    echo -e "${green}\xE2\x9C\x94${reset} Deploy demeris ingress"
    kubectl apply \
        --context kind-$CLUSTER_NAME \
        -f local-env/ingress.yaml
fi

if [ "$COMMAND" = "down" ]
then
    if [ "$(docker ps | grep $CLUSTER_NAME-local-proxy)" ]
    then
        echo -e "${green}\xE2\x9C\x94${reset} Deleting local proxy container"
        docker stop $CLUSTER_NAME-local-proxy &> /dev/null
    fi

    if kind get clusters | grep $CLUSTER_NAME &> /dev/null
    then
        echo -e "${green}\xE2\x9C\x94${reset} Deleting cluster $CLUSTER_NAME"
        kind delete cluster --name $CLUSTER_NAME &> /dev/null
    fi
fi

if [ "$COMMAND" = "connect-sql" ]
then
    kubectl run cockroachdb-client \
        --context kind-$CLUSTER_NAME \
        -it \
        --image=cockroachdb/cockroach:v20.2.8 \
        --rm \
        --restart=Never \
        -- \
        sql --insecure --host=cockroachdb-public
fi