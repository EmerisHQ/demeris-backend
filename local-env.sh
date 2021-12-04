#!/bin/bash

CLUSTER_NAME=emeris
REBUILD=false
NO_CHAINS=false
STARPORT_OPERATOR_REPO=git@github.com:allinbits/starport-operator.git
TRACELISTENER_REPO=git@github.com:allinbits/tracelistener.git
PRICE_ORACLE_REPO=git@github.com:allinbits/emeris-price-oracle.git
CNS_SERVER_REPO=git@github.com:allinbits/emeris-cns-server.git
TICKET_WATCHER_REPO=git@github.com:allinbits/emeris-ticket-watcher.git
RPC_WATCHER_REPO=git@github.com:allinbits/emeris-rpcwatcher.git
API_SERVER_REPO=git@github.com:allinbits/demeris-api-server.git
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
    echo -e "  -b, --rebuild \t\t Whether to (re)build docker images"
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
    -b|--rebuild)
    REBUILD=true
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
    if [ "$REBUILD" = "true" ]; then
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

    ### Apply CRDs
    kubectl apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/release-0.43/example/prometheus-operator-crd/monitoring.coreos.com_alertmanagerconfigs.yaml
    kubectl apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/release-0.43/example/prometheus-operator-crd/monitoring.coreos.com_alertmanagers.yaml
    kubectl apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/release-0.43/example/prometheus-operator-crd/monitoring.coreos.com_podmonitors.yaml
    kubectl apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/release-0.43/example/prometheus-operator-crd/monitoring.coreos.com_probes.yaml
    kubectl apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/release-0.43/example/prometheus-operator-crd/monitoring.coreos.com_prometheuses.yaml
    kubectl apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/release-0.43/example/prometheus-operator-crd/monitoring.coreos.com_prometheusrules.yaml
    kubectl apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/release-0.43/example/prometheus-operator-crd/monitoring.coreos.com_servicemonitors.yaml
    kubectl apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/release-0.43/example/prometheus-operator-crd/monitoring.coreos.com_thanosrulers.yaml

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
        --version 6.0.8 \
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
    if [ "$(docker images -q emeris/tracelistener 2> /dev/null)" != "" ] && [ "$REBUILD" = "false" ]
    then
        echo -e "${green}\xE2\x9C\x94${reset} Image emeris/tracelistener already exists"
    else
        if [ ! -d .tracelistener/.git ]
        then
            echo -e "${green}\xE2\x9C\x94${reset} Cloning tracelistener repo"
            git clone $TRACELISTENER_REPO .tracelistener &> /dev/null
        else
            echo -e "${green}\xE2\x9C\x94${reset} Fetching tracelistener latest changes"
            cd .tracelistener
            git pull $TRACELISTENER_REPO &> /dev/null
            cd ..
        fi
        echo -e "${green}\xE2\x9C\x94${reset} Building emeris/tracelistener image"
        cd .tracelistener
        docker build -t emeris/tracelistener --build-arg GIT_TOKEN=$GITHUB_TOKEN -f Dockerfile .
        cd ..
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

    ### Ensure cns-server and admin-ui images
    if [ "$(docker images -q emeris/cns-server 2> /dev/null)" != "" ] && [ "$(docker images -q emeris/admin-ui 2> /dev/null)" != "" ] && [ "$REBUILD" = "false" ]
    then
        echo -e "${green}\xE2\x9C\x94${reset} Image emeris/cns-server already exists"
        echo -e "${green}\xE2\x9C\x94${reset} Image emeris/admin-ui already exists"
    else
        if [ ! -d .cns-server/.git ]
        then
            echo -e "${green}\xE2\x9C\x94${reset} Cloning cns-server repo"
            git clone $CNS_SERVER_REPO .cns-server &> /dev/null
        else
            echo -e "${green}\xE2\x9C\x94${reset} Fetching cns-server latest changes"
            cd .cns-server
            git pull $CNS_SERVER_REPO &> /dev/null
            cd ..
        fi
        cd .cns-server
        echo -e "${green}\xE2\x9C\x94${reset} Building emeris/cns-server image"
        docker build -t emeris/cns-server --build-arg GIT_TOKEN=$GITHUB_TOKEN -f Dockerfile .
        echo -e "${green}\xE2\x9C\x94${reset} Building emeris/admin-ui image"
        docker build -t emeris/admin-ui ./cns/admin/emeris-admin
        cd ..
    fi
    echo -e "${green}\xE2\x9C\x94${reset} Pushing emeris/cns-server image to cluster"
    kind load docker-image emeris/cns-server --name $CLUSTER_NAME &> /dev/null

    echo -e "${green}\xE2\x9C\x94${reset} Pushing emeris/admin-ui image to cluster"
    kind load docker-image emeris/admin-ui --name $CLUSTER_NAME &> /dev/null

    echo -e "${green}\xE2\x9C\x94${reset} Deploying emeris/cns-server"
    helm upgrade cns-server \
        --install \
        --kube-context kind-$CLUSTER_NAME \
        --namespace emeris \
        --set imagePullPolicy=Never \
        helm/emeris-cns-server \
        &> /dev/null

    echo -e "${green}\xE2\x9C\x94${reset} Deploying emeris/admin-ui"
    helm upgrade admin-ui \
        --install \
        --kube-context kind-$CLUSTER_NAME \
        --namespace emeris \
        --set imagePullPolicy=Never \
        helm/emeris-admin-ui \
        &> /dev/null

    ### Ensure api-server image
    if [ "$(docker images -q emeris/api-server 2> /dev/null)" != "" ] && [ "$BUILD" = "false" ]
    then
        echo -e "${green}\xE2\x9C\x94${reset} Image emeris/api-server already exists"
    else
        if [ ! -d .api-server/.git ]
        then
            echo -e "${green}\xE2\x9C\x94${reset} Cloning api-server repo"
            git clone $API_SERVER_REPO .api-server &> /dev/null
        else
            echo -e "${green}\xE2\x9C\x94${reset} Fetching api-server latest changes"
            cd .api-server
            git pull $API_SERVER_REPO &> /dev/null
            cd ..
        fi
        cd .api-server
        echo -e "${green}\xE2\x9C\x94${reset} Re-building emeris/api-server image"
        docker build -t emeris/api-server --build-arg GIT_TOKEN=$GITHUB_TOKEN -f Dockerfile .
        cd ..
    fi
    echo -e "${green}\xE2\x9C\x94${reset} Pushing emeris/api-server image to cluster"
    kind load docker-image emeris/api-server --name $CLUSTER_NAME &> /dev/null

    helm upgrade api-server \
        --install \
        --kube-context kind-$CLUSTER_NAME \
        --namespace emeris \
        --set imagePullPolicy=Never \
        .api-server/helm \
        &> /dev/null

    ### Ensure rpcwatcher image
    if [ "$(docker images -q emeris/rpcwatcher 2> /dev/null)" != "" ] && [ "$BUILD" = "false" ]
        then
            echo -e "${green}\xE2\x9C\x94${reset} Image emeris/rpcwatcher already exists"
        else
            if [ ! -d .rpcwatcher/.git ]
            then
                echo -e "${green}\xE2\x9C\x94${reset} Cloning emeris/rpcwatcher repo"
                git clone $RPC_WATCHER_REPO .rpcwatcher &> /dev/null
            else
                echo -e "${green}\xE2\x9C\x94${reset} Fetching rpcwatcher latest changes"
                cd .rpcwatcher
                git pull $RPC_WATCHER_REPO &> /dev/null
                cd ..
            fi
            cd .rpcwatcher
            echo -e "${green}\xE2\x9C\x94${reset} Re-building emeris/rpcwatcher image"
            docker build -t emeris/rpcwatcher --build-arg GIT_TOKEN=$GITHUB_TOKEN -f Dockerfile .
            cd ..
        fi
        echo -e "${green}\xE2\x9C\x94${reset} Pushing emeris/rpcwatcher image to cluster"
        kind load docker-image emeris/rpcwatcher --name $CLUSTER_NAME &> /dev/null

        helm upgrade rpcwatcher \
            --install \
            --kube-context kind-$CLUSTER_NAME \
            --namespace emeris \
            --set imagePullPolicy=Never \
            .rpcwatcher/helm \
            &> /dev/null

    ### Ensure ticket-watcher image
    if [ "$(docker images -q emeris/ticket-watcher 2> /dev/null)" != "" ] && [ "$BUILD" = "false" ]
    then
        echo -e "${green}\xE2\x9C\x94${reset} Image emeris/ticket-watcher already exists"
    else
        if [ ! -d .ticket-watcher/.git ]
        then
            echo -e "${green}\xE2\x9C\x94${reset} Cloning ticket-watcher repo"
            git clone $TICKET_WATCHER_REPO .ticket-watcher &> /dev/null
        else
            echo -e "${green}\xE2\x9C\x94${reset} Fetching ticket-watcher latest changes"
            cd .ticket-watcher
            git pull $TICKET_WATCHER_REPO &> /dev/null
            cd ..
        fi
        cd .ticket-watcher
        echo -e "${green}\xE2\x9C\x94${reset} Re-building emeris/ticket-watcher image"
        docker build -t emeris/ticket-watcher --build-arg GIT_TOKEN=$GITHUB_TOKEN -f Dockerfile .
        cd ..
    fi
    echo -e "${green}\xE2\x9C\x94${reset} Pushing emeris/ticket-watcher image to cluster"
    kind load docker-image emeris/ticket-watcher --name $CLUSTER_NAME &> /dev/null

    helm upgrade ticket-watcher \
        --install \
        --kube-context kind-$CLUSTER_NAME \
        --namespace emeris \
        --set imagePullPolicy=Never \
        .ticket-watcher/helm \
        &> /dev/null

    ### Ensure price-oracle-server image
    if [ "$(docker images -q emeris/price-oracle-server 2> /dev/null)" != "" ] && [ "$BUILD" = "false" ]
    then
        echo -e "${green}\xE2\x9C\x94${reset} Image emeris/price-oracle-server already exists"
    else
        if [ ! -d .price-oracle/.git ]
        then
            echo -e "${green}\xE2\x9C\x94${reset} Cloning price-oracle repo"
            git clone $PRICE_ORACLE_REPO .price-oracle &> /dev/null
        else
            echo -e "${green}\xE2\x9C\x94${reset} Fetching price-oracle latest changes"
            cd .price-oracle
            git pull $PRICE_ORACLE_REPO &> /dev/null
            cd ..
        fi
        cd .price-oracle
        echo -e "${green}\xE2\x9C\x94${reset} Re-building emeris/price-oracle-server image"
        docker build -t emeris/price-oracle-server --build-arg GIT_TOKEN=$GITHUB_TOKEN -f Dockerfile .
        cd ..
    fi
    echo -e "${green}\xE2\x9C\x94${reset} Pushing emeris/price-oracle-server image to cluster"
    kind load docker-image emeris/price-oracle-server --name $CLUSTER_NAME &> /dev/null

    helm upgrade price-oracle \
        --install \
        --kube-context kind-$CLUSTER_NAME \
        --namespace emeris \
        --set imagePullPolicy=Never \
        .price-oracle/helm \
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
