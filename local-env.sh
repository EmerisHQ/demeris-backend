#!/bin/bash

CLUSTER_NAME=emeris
REBUILD=false
NO_CHAINS=false
MONITORING=false
ENABLE_FRONTEND=false
STARPORT_OPERATOR_REPO=git@github.com:allinbits/starport-operator.git
STARPORT_OPERATOR_VERSION=v0.0.1-alpha.45
TRACELISTENER_REPO=git@github.com:allinbits/tracelistener.git
PRICE_ORACLE_REPO=git@github.com:allinbits/emeris-price-oracle.git
CNS_SERVER_REPO=git@github.com:allinbits/emeris-cns-server.git
SDK_SERVICE_REPO=git@github.com:allinbits/sdk-service.git
TICKET_WATCHER_REPO=git@github.com:allinbits/emeris-ticket-watcher.git
RPC_WATCHER_REPO=git@github.com:allinbits/emeris-rpcwatcher.git
API_SERVER_REPO=git@github.com:allinbits/demeris-api-server.git
FRONTEND_REPO=git@github.com:allinbits/demeris.git
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
    echo -e "  -fe, --enable-frontend Enable frontend"
    echo -e "  -n,  --cluster-name \t Kind cluster name"
    echo -e "  -b,  --rebuild \t Whether to (re)build docker images"
    echo -e "  -nc, --no-chains \t Do not deploy chains inside the cluster"
    echo -e "  -m,  --monitoring \t Setup monitoring infrastructure"
    echo -e "  -h,  --help \t\t Show this menu\n"
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
    -fe|--enable-frontend)
    ENABLE_FRONTEND=true
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
    hostPort: 8443
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
        echo -e "${green}\xE2\x9C\x94${reset} Fetching starport-operator $STARPORT_OPERATOR_VERSION"
        cd .starport-operator
        git pull &> /dev/null
        git checkout $STARPORT_OPERATOR_VERSION
        cd ..
    fi

    echo -e "${green}\xE2\x9C\x94${reset} Ensure starport-operator is installed"
    helm upgrade starport-operator \
        --install \
        --create-namespace \
        --kube-context kind-$CLUSTER_NAME \
        --namespace starport-system \
        --set webHooksEnabled=false \
        --set workerCount=5 \
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
            git checkout main
        else
            echo -e "${green}\xE2\x9C\x94${reset} Fetching tracelistener latest changes"
            cd .tracelistener
            git checkout main
            git pull &> /dev/null
            cd ..
        fi
        echo -e "${green}\xE2\x9C\x94${reset} Building emeris/tracelistener image"
        cd .tracelistener
        docker build -t emeris/tracelistener --build-arg GIT_TOKEN=$GITHUB_TOKEN -f Dockerfile .
        cd ..
    fi
    echo -e "${green}\xE2\x9C\x94${reset} Pushing emeris/tracelistener image to cluster"
    kind load docker-image emeris/tracelistener --name $CLUSTER_NAME &> /dev/null

    ### Ensure tracelistener44 image
    if [ "$(docker images -q emeris/tracelistener44 2> /dev/null)" != "" ] && [ "$REBUILD" = "false" ]
    then
        echo -e "${green}\xE2\x9C\x94${reset} Image emeris/tracelistener44 already exists"
    else
        if [ ! -d .tracelistener/.git ]
        then
            echo -e "${green}\xE2\x9C\x94${reset} Cloning tracelistener repo"
            git clone $TRACELISTENER_REPO .tracelistener &> /dev/null
            git checkout sdk-44-support
        else
            echo -e "${green}\xE2\x9C\x94${reset} Fetching tracelistener44 latest changes"
            cd .tracelistener
            git checkout sdk-44-support
            git pull &> /dev/null
            cd ..
        fi
        echo -e "${green}\xE2\x9C\x94${reset} Building emeris/tracelistener44 image"
        cd .tracelistener
        docker build -t emeris/tracelistener44 --build-arg GIT_TOKEN=$GITHUB_TOKEN -f Dockerfile .
        cd ..
    fi
    echo -e "${green}\xE2\x9C\x94${reset} Pushing emeris/tracelistener44 image to cluster"
    kind load docker-image emeris/tracelistener44 --name $CLUSTER_NAME &> /dev/null

    ### Setup nodes
    if [ "$NO_CHAINS" = "false" ]; then
      echo -e "${green}\xE2\x9C\x94${reset} Create/update nodes"
      kubectl apply \
          --context kind-$CLUSTER_NAME \
          -f local-env/nodes
    fi

    ### Ensure cns-server image
    if [ "$(docker images -q emeris/cns-server 2> /dev/null)" != "" ] && [ "$REBUILD" = "false" ]
    then
        echo -e "${green}\xE2\x9C\x94${reset} Image emeris/cns-server already exists"
    else
        if [ ! -d .cns-server/.git ]
        then
            echo -e "${green}\xE2\x9C\x94${reset} Cloning cns-server repo"
            git clone $CNS_SERVER_REPO .cns-server &> /dev/null
        else
            echo -e "${green}\xE2\x9C\x94${reset} Fetching cns-server latest changes"
            cd .cns-server
            git pull &> /dev/null
            cd ..
        fi
        cd .cns-server
        echo -e "${green}\xE2\x9C\x94${reset} Building emeris/cns-server image"
        docker build -t emeris/cns-server --build-arg GIT_TOKEN=$GITHUB_TOKEN -f Dockerfile .
        cd ..
    fi
    echo -e "${green}\xE2\x9C\x94${reset} Pushing emeris/cns-server image to cluster"
    kind load docker-image emeris/cns-server --name $CLUSTER_NAME &> /dev/null

    echo -e "${green}\xE2\x9C\x94${reset} Deploying emeris/cns-server"
    helm upgrade cns-server \
        --install \
        --kube-context kind-$CLUSTER_NAME \
        --namespace emeris \
        --set imagePullPolicy=Never \
        --set redirectURL=http://localhost:3000/login \
        --set test=true \
        --set resources=null \
        .cns-server/helm/ \
        &> /dev/null

    ### Ensure api-server image
    if [ "$(docker images -q emeris/api-server 2> /dev/null)" != "" ] && [ "$REBUILD" = "false" ]
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
            git pull &> /dev/null
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
        --set replicas=1 \
        --set imagePullPolicy=Never \
        --set serviceMonitorEnabled=$MONITORING \
        --set resources=null \
        .api-server/helm \
        &> /dev/null

    ### Ensure rpcwatcher image
    if [ "$(docker images -q emeris/rpcwatcher 2> /dev/null)" != "" ] && [ "$REBUILD" = "false" ]
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
            git pull &> /dev/null
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
        --set resources=null \
        .rpcwatcher/helm \
        &> /dev/null

    ### Ensure ticket-watcher image
    if [ "$(docker images -q emeris/ticket-watcher 2> /dev/null)" != "" ] && [ "$REBUILD" = "false" ]
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
            git pull &> /dev/null
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
        --set resources=null \
        .ticket-watcher/helm \
        &> /dev/null

    ### Ensure price-oracle-server image
    if [ "$(docker images -q emeris/price-oracle-server 2> /dev/null)" != "" ] && [ "$REBUILD" = "false" ]
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
            git pull &> /dev/null
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
        --set replicas=1 \
        --set imagePullPolicy=Never \
        --set resources=null \
        --set fixerKey=$FIXER_KEY \
        .price-oracle/helm \
        &> /dev/null

    ### Ensure sdk-service-42 image
    if [ "$(docker images -q emeris/sdk-service-42 2> /dev/null)" != "" ] && [ "$REBUILD" = "false" ]
    then
        echo -e "${green}\xE2\x9C\x94${reset} Image emeris/sdk-service-42 already exists"
    else
        if [ ! -d .sdk-service/.git ]
        then
            echo -e "${green}\xE2\x9C\x94${reset} Cloning sdk-service repo"
            git clone $SDK_SERVICE_REPO .sdk-service &> /dev/null
        else
            echo -e "${green}\xE2\x9C\x94${reset} Fetching sdk-service latest changes"
            cd .sdk-service
            git pull &> /dev/null
            cd ..
        fi
        cd .sdk-service
        echo -e "${green}\xE2\x9C\x94${reset} Building emeris/sdk-service-42 image"
        docker build -t emeris/sdk-service-42 --build-arg SDK_TARGET=v42 --build-arg GIT_TOKEN=$GITHUB_TOKEN -f Dockerfile .
        cd ..
    fi
    echo -e "${green}\xE2\x9C\x94${reset} Pushing emeris/sdk-service-42 image to cluster"
    kind load docker-image emeris/sdk-service-42 --name $CLUSTER_NAME &> /dev/null

    echo -e "${green}\xE2\x9C\x94${reset} Deploying emeris/sdk-service-42"
    helm upgrade sdk-service-v42 \
        --install \
        --kube-context kind-$CLUSTER_NAME \
        --namespace emeris \
        --set imagePullPolicy=Never \
        --set image=emeris/sdk-service-42 \
        --set resources=null \
        .sdk-service/helm/ \
        &> /dev/null

    ### Ensure sdk-service-44 image
    if [ "$(docker images -q emeris/sdk-service-44 2> /dev/null)" != "" ] && [ "$REBUILD" = "false" ]
    then
        echo -e "${green}\xE2\x9C\x94${reset} Image emeris/sdk-service-44 already exists"
    else
        if [ ! -d .sdk-service/.git ]
        then
            echo -e "${green}\xE2\x9C\x94${reset} Cloning sdk-service repo"
            git clone $SDK_SERVICE_REPO .sdk-service &> /dev/null
        else
            echo -e "${green}\xE2\x9C\x94${reset} Fetching sdk-service latest changes"
            cd .sdk-service
            git pull &> /dev/null
            cd ..
        fi
        cd .sdk-service
        echo -e "${green}\xE2\x9C\x94${reset} Building emeris/sdk-service-44 image"
        docker build -t emeris/sdk-service-44 --build-arg SDK_TARGET=v44 --build-arg GIT_TOKEN=$GITHUB_TOKEN -f Dockerfile .
        cd ..
    fi
    echo -e "${green}\xE2\x9C\x94${reset} Pushing emeris/sdk-service-44 image to cluster"
    kind load docker-image emeris/sdk-service-44 --name $CLUSTER_NAME &> /dev/null

    echo -e "${green}\xE2\x9C\x94${reset} Deploying emeris/sdk-service-44"
    helm upgrade sdk-service-v44 \
        --install \
        --kube-context kind-$CLUSTER_NAME \
        --namespace emeris \
        --set imagePullPolicy=Never \
        --set image=emeris/sdk-service-44 \
        --set resources=null \
        .sdk-service/helm/ \
        &> /dev/null

    ## Ensure Emeris ingress
    echo -e "${green}\xE2\x9C\x94${reset} Deploy emeris ingress"
    kubectl apply \
        --context kind-$CLUSTER_NAME \
        --namespace emeris \
        -f local-env/ingress.yaml

    ## Setup monitoring infrastructure
    if [ "$MONITORING" = "true" ]
    then
      echo -e "${green}\xE2\x9C\x94${reset} Deploying monitoring"
      ### Apply CRDs
      kubectl --context kind-$CLUSTER_NAME --namespace emeris apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/release-0.43/example/prometheus-operator-crd/monitoring.coreos.com_alertmanagerconfigs.yaml
      kubectl --context kind-$CLUSTER_NAME --namespace emeris apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/release-0.43/example/prometheus-operator-crd/monitoring.coreos.com_alertmanagers.yaml
      kubectl --context kind-$CLUSTER_NAME --namespace emeris apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/release-0.43/example/prometheus-operator-crd/monitoring.coreos.com_podmonitors.yaml
      kubectl --context kind-$CLUSTER_NAME --namespace emeris apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/release-0.43/example/prometheus-operator-crd/monitoring.coreos.com_probes.yaml
      kubectl --context kind-$CLUSTER_NAME --namespace emeris apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/release-0.43/example/prometheus-operator-crd/monitoring.coreos.com_prometheuses.yaml
      kubectl --context kind-$CLUSTER_NAME --namespace emeris apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/release-0.43/example/prometheus-operator-crd/monitoring.coreos.com_prometheusrules.yaml
      kubectl --context kind-$CLUSTER_NAME --namespace emeris apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/release-0.43/example/prometheus-operator-crd/monitoring.coreos.com_servicemonitors.yaml
      kubectl --context kind-$CLUSTER_NAME --namespace emeris apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/release-0.43/example/prometheus-operator-crd/monitoring.coreos.com_thanosrulers.yaml
      
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

    ### Setup relayer
    if [ "$NO_CHAINS" = "false" ]; then
      echo -e "${green}\xE2\x9C\x94${reset} Create/update relayer"
      kubectl apply \
          --context kind-$CLUSTER_NAME \
          -f local-env/relayer.yaml
    fi

    ### Setup chains
    if [ "$NO_CHAINS" = "false" ]; then
      echo -e "${green}\xE2\x9C\x94${reset} Waiting for CNS to be ready"
      kubectl wait pod \
        --context kind-$CLUSTER_NAME \
        --namespace ingress-nginx \
        --for=condition=ready \
        --selector=app.kubernetes.io/name=emeris-cns-server \
        --timeout=90s \
        &> /dev/null
      for f in local-env/chains/*
      do
      echo -e "${green}\xE2\x9C\x94${reset} Create/update $(basename $f .json)"
      curl -X POST -d @$f http://localhost:8000/v1/cns/add
      done
    fi

    ### Start frontend
    if [ "$ENABLE_FRONTEND" = "true" ]; then
        if [ ! -d .emeris-frontend/.git ]
        then
            echo -e "${green}\xE2\x9C\x94${reset} Cloning emeris-frontend repo"
            git clone $FRONTEND_REPO .emeris-frontend &> /dev/null
        else
            echo -e "${green}\xE2\x9C\x94${reset} Fetching emeris-frontend latest changes"
            cd .emeris-frontend
            git pull &> /dev/null
            cd ..
        fi

        echo -e "${green}\xE2\x9C\x94${reset} Starting emeris-frontend"
        docker run -d --name emeris-frontend --rm \
            -v ${PWD}/.emeris-frontend:/app \
            -w /app \
            -p 3000:3000 \
            -e PORT=3000 \
            -e VUE_APP_HUB_CHAIN=cosmos-hub-testnet \
            -e VUE_APP_EMERIS_PROD_ENDPOINT=http://localhost:8000/v1 \
            -e VUE_APP_EMERIS_PROD_LIQUIDITY_ENDPOINT=http://localhost:8000/v1/liquidity \
            --entrypoint /bin/bash \
            node:16 -c \
            "npm ci && npm run serve"

        echo -e "${green}\xE2\x9C\x94${reset} Waiting for emeris-frontend to be ready..."
        while [[ "$(curl -s -o /dev/null -w ''%{http_code}'' localhost:3000)" != "200" ]]; do sleep 3; done
    fi
    echo -e "${green}\xE2\x9C\x94${reset} All done"
fi

if [ "$COMMAND" = "down" ]
then
    if kind get clusters | grep $CLUSTER_NAME &> /dev/null
    then
        echo -e "${green}\xE2\x9C\x94${reset} Deleting cluster $CLUSTER_NAME"
        kind delete cluster --name $CLUSTER_NAME &> /dev/null
    fi
    docker stop emeris-frontend
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
