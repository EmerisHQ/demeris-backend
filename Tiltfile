load('ext://helm_resource', 'helm_resource', 'helm_repo')
load('ext://namespace', 'namespace_create', 'namespace_inject')

update_settings(
    k8s_upsert_timeout_secs = 300,
)

print("Working on context: ", k8s_context())

k8s_yaml('local-env/namespace.yaml')

helm_repo('ingress-nginx-repo', 'https://kubernetes.github.io/ingress-nginx')
helm_resource('ingress-nginx', 'ingress-nginx-repo/ingress-nginx', namespace='ingress-nginx', flags=[
    '--create-namespace'
])

nginx_patch = read_file('./local-env/nginx-patch.yaml')
local_resource('nginx_patch',
    'kubectl patch \
      --context %s \
      --namespace ingress-nginx \
      svc ingress-nginx-controller \
      --patch "%s"' % (k8s_context(), nginx_patch),
    resource_deps=[])

k8s_yaml('local-env/ingress.yaml')

helm_repo('cockroachdb-repo', 'https://charts.cockroachdb.com/')
helm_resource('cockroachdb', 'cockroachdb-repo/cockroachdb', namespace='emeris', flags=[
    "--version", "7.0.0",
    "--set", "tls.enabled=false",
    "--set", "config.single-node=true",
    "--set", "statefulset.replicas=1",
])
k8s_resource('cockroachdb', port_forwards=[
  port_forward(26257, 26257),
  port_forward(65001, 8080, name='cockroachdb admin ui'),
])

helm_repo('bitnami', 'https://charts.bitnami.com/bitnami')
helm_resource('redis', 'bitnami/redis', namespace='emeris', flags=[
    "--set", "auth.enabled=false",
    "--set", "auth.sentinel=false",
    "--set", "architecture=standalone",
])

# starport-operator
namespace_create(
    'starport-system',
    allow_duplicates=True,
)
k8s_yaml(helm(
    '../starport-operator/helm/starport-operator',
    name='starport-operator',
    namespace='starport-system',
    set=[
        'webHooksEnabled=false',
        'workerCount=5',
    ]
))

GITHUB_TOKEN = os.getenv("GITHUB_TOKEN")
if not GITHUB_TOKEN:
    fail("⚠️  you should export GITHUB_TOKEN with a valid GitHub token (with at least repository access) to build images.")

k8s_kind('NodeSet', image_json_path='{.spec.config.nodes.traceStoreContainer.image}')
k8s_kind('Relayer')

# tracelistener images
docker_build(
    'emeris/tracelistener-v42',
    '../tracelistener',
    build_args={'GIT_TOKEN': GITHUB_TOKEN, 'SDK_TARGET': 'v42'},
)

docker_build(
    'emeris/tracelistener-v44',
    '../tracelistener',
    build_args={'GIT_TOKEN': GITHUB_TOKEN, 'SDK_TARGET': 'v44'},
)

# chains
k8s_yaml('local-env/nodes/cosmos-hub.yaml')
k8s_yaml('local-env/nodes/akash.yaml')
k8s_yaml('local-env/relayer.yaml')

# our services
# CNS
docker_build(
    'emeris/cns-server',
    '../emeris-cns-server',
    build_args={'GIT_TOKEN': GITHUB_TOKEN},
)
k8s_yaml(helm(
    '../emeris-cns-server/helm',
    name='cns-server',
    namespace='emeris',
    set=[
        'imagePullPolicy=Never',
        'redirectURL=http://localhost:3000/login',
        'test=true',
        'resources=null',
    ]
))

def add_cns_chain(chain_name):
    local_resource(
        'cns-add-chain-%s' % chain_name, 
        'curl -f -X POST -d @./ci/dev/chains/%s.json http://localhost:8000/v1/cns/add' % chain_name,
        deps=['ingress-nginx', 'cns-server'],
    )

add_cns_chain('cosmos-hub')
add_cns_chain('akash')

# API SERVER
docker_build(
    'emeris/api-server',
    '../demeris-api-server',
    build_args={'GIT_TOKEN': GITHUB_TOKEN},
)
k8s_yaml(helm(
    '../demeris-api-server/helm',
    name='api-server',
    namespace='emeris',
    set=[
        'replicas=1',
        'imagePullPolicy=Never',
        'serviceMonitorEnabled=false', # TODO make it parametric
        'resources=null',
    ]
))

# RPC WATCHER
docker_build(
    'emeris/rpcwatcher',
    '../emeris-rpcwatcher',
    build_args={'GIT_TOKEN': GITHUB_TOKEN},
)
k8s_yaml(helm(
    '../emeris-rpcwatcher/helm',
    name='rpcwatcher',
    namespace='emeris',
    set=[
        'imagePullPolicy=Never',
        'resources=null',
    ]
))

# DEXINFO
docker_build(
  'emeris/dexinfo',
  '../emeris-dexinfo',
)
k8s_yaml(helm(
    '../emeris-dexinfo/helm',
    name='dexinfo',
    namespace='emeris',
    set=[
        'image=emeris/dexinfo:latest',
        'imagePullPolicy=Never',
        'resources=null',
    ]
))

# DAGGREGATION
docker_build(
  'emeris/daggregation',
  '../daggregation',
)
k8s_yaml(helm(
    '../daggregation/helm',
    name='daggregation',
    namespace='emeris',
    set=[
        'imagePullPolicy=Never',
        'resources=null',
        'dexInfoSwapsUrl=http://dexinfo:8080/swaps',
        'dexInfoVerifiedDenoms=http://api-server:8000/verified_denoms',
    ]
))

# TICKET WATCHER
docker_build(
    'emeris/ticket-watcher',
    '../emeris-ticket-watcher',
    build_args={'GIT_TOKEN': GITHUB_TOKEN},
)
k8s_yaml(helm(
    '../emeris-ticket-watcher/helm',
    name='ticket-watcher',
    namespace='emeris',
    set=[
        'imagePullPolicy=Never',
        'resources=null',
    ]
))

# PRICE ORACLE
FIXER_KEY = os.getenv("FIXER_KEY")
if not FIXER_KEY:
    print("⚠️ Set FIXER_KEY env variable for enabling price-oracle")
else:
    docker_build(
        'emeris/price-oracle-server',
        '../emeris-price-oracle',
        build_args={'GIT_TOKEN': GITHUB_TOKEN},
    )
    k8s_yaml(helm(
        '../emeris-price-oracle/helm',
        name='price-oracle',
        namespace='emeris',
        set=[
            'replicas=1',
            'fixerKey=%s' % FIXER_KEY,
            'imagePullPolicy=Never',
            'resources=null',
        ]
    ))

# SDK-SERVICE-v42
docker_build(
    'emeris/sdk-service-42',
    '../sdk-service',
    build_args={
        'GIT_TOKEN': GITHUB_TOKEN,
        'SDK_TARGET': 'v42',
    },
)
k8s_yaml(helm(
    '../sdk-service/helm',
    name='sdk-service-v42',
    namespace='emeris',
    set=[
        'replicas=1',
        'image=emeris/sdk-service-42',
        'resources=null',
    ]
))

# SDK-SERVICE-v44
docker_build(
    'emeris/sdk-service-44',
    '../sdk-service',
    build_args={
        'GIT_TOKEN': GITHUB_TOKEN,
        'SDK_TARGET': 'v44',
    },
)
k8s_yaml(helm(
    '../sdk-service/helm',
    name='sdk-service-v44',
    namespace='emeris',
    set=[
        'replicas=1',
        'image=emeris/sdk-service-44',
        'resources=null',
    ]
))

# FRONTEND
# TODO: move Dockerfile and k8s pod yaml to separate files
# NOTE(tb): we can't use docker_build tilt instruction for the front because
# it implicitely discards the .git folder from the docker context, and vite
# needs it for the front build. The workaround is to replace docker_build with
# custom_build, but first we need to create the front Dockerfile.
local("""echo 'FROM node:16
WORKDIR /app
COPY . /app
RUN --mount=type=cache,target=/root/.npm npm ci
CMD ["npm", "run", "dev", "--", "--host", "localhost"]' > /tmp/Dockerfile.front
"""
)
custom_build(
    'emeris/frontend',
    # TODO(tb) A bug in Docker Buildkit forbids us to use `-f /tmp/Dockerfile.front`
    # so we use an alternative.
    # See https://github.com/docker/cli/issues/2249
    'docker build -t $EXPECTED_REF -f - ../demeris < /tmp/Dockerfile.front',
    ['../demeris'],
    live_update=[
      sync('../demeris/src/', '/app/src/'),
    ]
)

k8s_yaml(blob("""
apiVersion: v1
kind: Pod
metadata:
  name: frontend
  namespace: emeris
spec:
  containers:
  - name: app
    image: emeris/frontend
    env:
      - name: VUE_APP_GIT_VERSION
        value: development
      - name: VITE_FEATURE_USE_DEV
        value: "true"
      - name: VITE_EMERIS_DEV_ENDPOINT
        value: http://localhost:8000/v1
      - name: VITE_EMERIS_DEV_LIQUIDITY_ENDPOINT
        value: http://localhost:8000/v1/liquidity
      - name: VITE_EMERIS_DEV_WEBSOCKET_ENDPOINT
        value: ws://localhost:8000/v1
"""))
k8s_resource('frontend', port_forwards=[8080])

# vim: set syntax=python:
