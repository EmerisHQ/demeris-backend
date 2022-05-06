#!/bin/bash -e
# This script setups the port forward for the blockchain RPC endpoints that are
# required by the tests.

kubectl port-forward svc/cosmos-hub 10000:26657 &
kubectl port-forward svc/akash      10001:26657 &
kubectl port-forward svc/terra      10002:26657 &