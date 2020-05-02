#!/usr/bin/env bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
ROOT_DIR=$SCRIPT_DIR/../..

(cd $ROOT_DIR && docker build -t projectjudge/wallclock-controller:latest .)
(cd $ROOT_DIR && kubectl apply -f __build__/k8s/manifests/wallclock-operator.yaml)
(cd $ROOT_DIR && kubectl apply -f __build__/k8s/manifests/timezones.yaml)
