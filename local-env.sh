#!/bin/bash

CLUSTER_NAME=emeris
BUILD=false
NO_CHAINS=false
STARPORT_OPERATOR_REPO=git@github.com:allinbits/starport-operator.git
KIND_CONFIG=""

usage()
{
    echo -e "Manage emeris local environment\n"
    echo -e "Usage: \n  $0 [command]\n"
    echo -e "Available Commands:"
    echo -e "  up \t\t Setup the development environment"
    echo -e "  down \t\t Tear down the development environment"
    echo -e "  connect-sql \t Connect to database using cockroach built-in SQL Client"
    echo -e "\nFlags:"
    echo -e "  -p, --port \t\t The local port at which the api will be served"
    echo -e "  -a, --address \t\t The address at which the api will be served, defaults to 127.0.0.1"
    echo -e "  -n, --cluster-name \t Kind cluster name"
    echo -e "  -b, --build \t\t Whether to (re)build docker images"
    echo -e "  -nc, --no-chains \t Do not deploy chains inside the cluster"
    echo -e "  -m, --monitoring \t Setup monitoring infrastructure"
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
    -a|--address)
    ADDRESS="$2"
    shift
    shift
    ;;
    -b|--build)
    BUILD=true
    shift
    ;;
    -m|--monitoring)
    MONITORING=true
    shift
    ;;
    -nc|--no-chains)
    NO_CHAINS=true
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
    if [ "$BUILD" = "true" ]; then
        if [ -z "$GITHUB_TOKEN" ]; then
          echo -e "${red}Error:${reset} you should export GITHUB_TOKEN with a valid GitHub token to build images.\n"
          usage
        fi
    fi
    ### Create the cluster

    if kind get clusters | grep $CLUSTER_NAME &> /dev/null
    then
        echo -e "${green}\xE2\x9C\x94${reset} Cluster $CLUSTER_NAME already exists"
    else
        echo -e "${green}\xE2\x9C\x94${reset} Creating cluster $CLUSTER_NAME"
        cat <<EOF | kind create cluster --name $CLUSTER_NAME --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  extraPortMappings:
  - containerPort: 30080
    hostPort: 8000
    protocol: TCP
  - containerPort: 30443
    hostPort: 443
    protocol: TCP
  - containerPort: 30090
    hostPort: 9090
    protocol: TCP
  - containerPort: 30880
    hostPort: 8080
    protocol: TCP
EOF
    fi

    ### Ensure emeris namespace
    kubectl create namespace emeris &> /dev/null
    kubectl config set-context kind-$CLUSTER_NAME --namespace=emeris &> /dev/null

    ### Ensure nginx ingress controller is deployed
    echo -e "${green}\xE2\x9C\x94${reset} Ensure nginx ingress controller is installed and running"
    kubectl apply \
        --context kind-$CLUSTER_NAME \
        -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/helm-chart-3.34.0/deploy/static/provider/kind/deploy.yaml \
        &> /dev/null
    kubectl patch \
      --context kind-$CLUSTER_NAME \
      --namespace ingress-nginx \
      svc ingress-nginx-controller \
      --patch "$(cat local-env/nginx-patch.yaml)" \
      &> /dev/null

    ### Wait for nginx to be up and running
    while : ; do
        kubectl get pod \
            --context kind-$CLUSTER_NAME \
            --namespace ingress-nginx \
            --selector=app.kubernetes.io/component=controller 2>&1 | grep -q controller && break
        sleep 2
    done
    kubectl wait pod \
        --context kind-$CLUSTER_NAME \
        --namespace ingress-nginx \
        --for=condition=ready \
        --selector=app.kubernetes.io/component=controller \
        --timeout=90s \
        &> /dev/null

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
    helm repo update &> /dev/null
    helm upgrade cockroachdb \
        --install \
        --kube-context kind-$CLUSTER_NAME \
        --namespace emeris \
        --set tls.enabled=false \
        --set config.single-node=true \
        --set statefulset.replicas=1 \
        cockroachdb/cockroachdb \
        &> /dev/null

    ### Ensure redis is installed
    echo -e "${green}\xE2\x9C\x94${reset} Ensure redis is installed and running"
    helm repo add bitnami https://charts.bitnami.com/bitnami &> /dev/null
    helm repo update &> /dev/null
    helm upgrade redis \
        --install \
        --kube-context kind-$CLUSTER_NAME \
        --namespace emeris \
        --set auth.enabled=false \
        --set auth.sentinel=false \
        --set architecture=standalone \
        bitnami/redis \
        &> /dev/null

    ### Ensure tracelistener image
    if [[ "$(docker images -q emeris/tracelistener 2> /dev/null)" == "" ]]
    then
        echo -e "${green}\xE2\x9C\x94${reset} Building emeris/tracelistener image"
        docker build -t emeris/tracelistener --build-arg GIT_TOKEN=$GITHUB_TOKEN -f Dockerfile.tracelistener .
    else
        if [ "$BUILD" = "true" ]
        then
            echo -e "${green}\xE2\x9C\x94${reset} Re-building emeris/tracelistener image"
            docker build -t emeris/tracelistener --build-arg GIT_TOKEN=$GITHUB_TOKEN -f Dockerfile.tracelistener .
        else
            echo -e "${green}\xE2\x9C\x94${reset} Image emeris/tracelistener already exists"
        fi
    fi
    echo -e "${green}\xE2\x9C\x94${reset} Pushing emeris/tracelistener image to cluster"
    kind load docker-image emeris/tracelistener --name $CLUSTER_NAME &> /dev/null

    ### Setup chains
    if [ "$NO_CHAINS" = "false" ]; then
      echo -e "${green}\xE2\x9C\x94${reset} Create/update chains"
      kubectl apply \
          --context kind-$CLUSTER_NAME \
          -f local-env/nodes
    fi

    ### Ensure cns-server image
    if [[ "$(docker images -q emeris/cns-server 2> /dev/null)" == "" ]]
    then
        echo -e "${green}\xE2\x9C\x94${reset} Building emeris/cns-server image"
        docker build -t emeris/cns-server --build-arg GIT_TOKEN=$GITHUB_TOKEN -f Dockerfile.cns-server .
    else
        if [ "$BUILD" = "true" ]
        then
            echo -e "${green}\xE2\x9C\x94${reset} Re-building emeris/cns-server image"
            docker build -t emeris/cns-server --build-arg GIT_TOKEN=$GITHUB_TOKEN -f Dockerfile.cns-server .
        else
            echo -e "${green}\xE2\x9C\x94${reset} Image emeris/cns-server already exists"
        fi
    fi
    echo -e "${green}\xE2\x9C\x94${reset} Pushing emeris/cns-server image to cluster"
    kind load docker-image emeris/cns-server --name $CLUSTER_NAME &> /dev/null

    echo -e "${green}\xE2\x9C\x94${reset} Deploying emeris/cns-server"
    helm upgrade cns-server \
        --install \
        --kube-context kind-$CLUSTER_NAME \
        --namespace emeris \
        --set imagePullPolicy=Never \
        helm/emeris-cns-server \
        &> /dev/null

    ### Ensure admin-ui image
    if [[ "$(docker images -q emeris/admin-ui 2> /dev/null)" == "" ]]
    then
        echo -e "${green}\xE2\x9C\x94${reset} Building emeris/admin-ui image"
        docker build -t emeris/admin-ui ./cns/admin/emeris-admin
    else
        if [ "$BUILD" = "true" ]
        then
            echo -e "${green}\xE2\x9C\x94${reset} Re-building emeris/admin-ui image"
            docker build -t emeris/admin-ui ./cns/admin/emeris-admin
        else
            echo -e "${green}\xE2\x9C\x94${reset} Image emeris/admin-ui already exists"
        fi
    fi
    echo -e "${green}\xE2\x9C\x94${reset} Pushing emeris/admin-ui image to cluster"
    kind load docker-image emeris/admin-ui --name $CLUSTER_NAME &> /dev/null

    echo -e "${green}\xE2\x9C\x94${reset} Deploying emeris/admin-ui"
    helm upgrade admin-ui \
        --install \
        --kube-context kind-$CLUSTER_NAME \
        --namespace emeris \
        --set imagePullPolicy=Never \
        helm/emeris-admin-ui \
        &> /dev/null

    ### Ensure api-server image
    if [[ "$(docker images -q emeris/api-server 2> /dev/null)" == "" ]]
    then
        echo -e "${green}\xE2\x9C\x94${reset} Building emeris/api-server image"
        docker build -t emeris/api-server --build-arg GIT_TOKEN=$GITHUB_TOKEN -f Dockerfile.api-server .
    else
        if [ "$BUILD" = "true" ]
        then
            echo -e "${green}\xE2\x9C\x94${reset} Re-building emeris/api-server image"
            docker build -t emeris/api-server --build-arg GIT_TOKEN=$GITHUB_TOKEN -f Dockerfile.api-server .
        else
            echo -e "${green}\xE2\x9C\x94${reset} Image emeris/api-server already exists"
        fi
    fi
    echo -e "${green}\xE2\x9C\x94${reset} Pushing emeris/api-server image to cluster"
    kind load docker-image emeris/api-server --name $CLUSTER_NAME &> /dev/null

    echo -e "${green}\xE2\x9C\x94${reset} Deploying emeris/api-server"
    helm upgrade api-server \
        --install \
        --kube-context kind-$CLUSTER_NAME \
        --namespace emeris \
        --set imagePullPolicy=Never \
        helm/emeris-api-server \
        &> /dev/null

    ### Ensure rpcwatcher image
    if [[ "$(docker images -q emeris/rpcwatcher 2> /dev/null)" == "" ]]
    then
        echo -e "${green}\xE2\x9C\x94${reset} Building emeris/rpcwatcher image"
        docker build -t emeris/rpcwatcher --build-arg GIT_TOKEN=$GITHUB_TOKEN -f Dockerfile.rpcwatcher .
    else
        if [ "$BUILD" = "true" ]
        then
            echo -e "${green}\xE2\x9C\x94${reset} Re-building emeris/rpcwatcher image"
            docker build -t emeris/rpcwatcher --build-arg GIT_TOKEN=$GITHUB_TOKEN -f Dockerfile.rpcwatcher .
        else
            echo -e "${green}\xE2\x9C\x94${reset} Image emeris/rpcwatcher already exists"
        fi
    fi
    echo -e "${green}\xE2\x9C\x94${reset} Pushing emeris/rpcwatcher image to cluster"
    kind load docker-image emeris/rpcwatcher --name $CLUSTER_NAME &> /dev/null

    echo -e "${green}\xE2\x9C\x94${reset} Deploying emeris/rpcwatcher"
    helm upgrade rpcwatcher \
        --install \
        --kube-context kind-$CLUSTER_NAME \
        --namespace emeris \
        --set imagePullPolicy=Never \
        helm/emeris-rpcwatcher \
        &> /dev/null

### Ensure ticket-watcher image
    if [[ "$(docker images -q emeris/ticket-watcher 2> /dev/null)" == "" ]]
    then
        echo -e "${green}\xE2\x9C\x94${reset} Building emeris/ticket-watcher image"
        docker build -t emeris/ticket-watcher --build-arg GIT_TOKEN=$GITHUB_TOKEN -f Dockerfile.ticket-watcher .
    else
        if [ "$BUILD" = "true" ]
        then
            echo -e "${green}\xE2\x9C\x94${reset} Re-building emeris/ticket-watcher image"
            docker build -t emeris/ticket-watcher --build-arg GIT_TOKEN=$GITHUB_TOKEN -f Dockerfile.ticket-watcher .
        else
            echo -e "${green}\xE2\x9C\x94${reset} Image emeris/ticket-watcher already exists"
        fi
    fi
    echo -e "${green}\xE2\x9C\x94${reset} Pushing emeris/ticket-watcher image to cluster"
    kind load docker-image emeris/ticket-watcher --name $CLUSTER_NAME &> /dev/null

    echo -e "${green}\xE2\x9C\x94${reset} Deploying emeris/ticket-watcher"
    helm upgrade ticket-watcher \
        --install \
        --kube-context kind-$CLUSTER_NAME \
        --namespace emeris \
        --set imagePullPolicy=Never \
        helm/emeris-ticket-watcher \
        &> /dev/null

    ### Ensure price-oracle-server image
     if [[ "$(docker images -q emeris/price-oracle-server 2> /dev/null)" == "" ]]
     then
         echo -e "${green}\xE2\x9C\x94${reset} Building emeris/price-oracle-server image"
         docker build -t emeris/price-oracle-server --build-arg GIT_TOKEN=$GITHUB_TOKEN -f Dockerfile.price-oracle .
     else
         if [ "$BUILD" = "true" ]
         then
             echo -e "${green}\xE2\x9C\x94${reset} Re-building emeris/price-oracle-server image"
             docker build -t emeris/price-oracle-server --build-arg GIT_TOKEN=$GITHUB_TOKEN -f Dockerfile.price-oracle .
         else
             echo -e "${green}\xE2\x9C\x94${reset} Image emeris/price-oracle-server already exists"
         fi
     fi
     echo -e "${green}\xE2\x9C\x94${reset} Pushing emeris/price-oracle-server image to cluster"
     kind load docker-image emeris/price-oracle-server --name $CLUSTER_NAME &> /dev/null

    helm upgrade price-oracle \
        --install \
        --kube-context kind-$CLUSTER_NAME \
        --namespace emeris \
        --set imagePullPolicy=Never \
        helm/emeris-price-oracle-server \
        &> /dev/null

    ## Ensure Emeris ingress
    echo -e "${green}\xE2\x9C\x94${reset} Deploy emeris ingress"
    kubectl apply \
        --context kind-$CLUSTER_NAME \
        --namespace emeris \
        -f local-env/ingress.yaml

    ## Setup monitoring infrastructure
    if [ "$MONITORING" = "true" ]; then
      echo -e "${green}\xE2\x9C\x94${reset} Deploying monitoring"
      helm upgrade monitoring-stack \
          --install \
          --create-namespace \
          --kube-context kind-$CLUSTER_NAME \
          --namespace monitoring \
          --set imagePullPolicy=Never \
          -f local-env/monitoring-values.yaml \
          prometheus-community/kube-prometheus-stack --version 15.4.6 \
          &> /dev/null

      kubectl apply \
        --context kind-$CLUSTER_NAME \
        --namespace emeris \
        -f local-env/service-monitors.yaml
    fi
fi

if [ "$COMMAND" = "down" ]
then
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
        --namespace emeris \
        -it \
        --image=cockroachdb/cockroach:v20.2.8 \
        --rm \
        --restart=Never \
        -- \
        sql --insecure --host=cockroachdb-public
fi
