#!/usr/bin/env bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
ROOT_DIR=$SCRIPT_DIR/../..

# Version compare function
vercmp() {
  version1=$1 version2=$2 condition=$3
  IFS=. v1_array=($version1) v2_array=($version2)
  v1=$((v1_array[0] * 100 + v1_array[1] * 10 + v1_array[2]))
  v2=$((v2_array[0] * 100 + v2_array[1] * 10 + v2_array[2]))
  diff=$((v2 - v1))
  [[ $condition = '=' ]] && ((diff == 0)) && return 0
  [[ $condition = '!=' ]] && ((diff != 0)) && return 0
  [[ $condition = '<' ]] && ((diff > 0)) && return 0
  [[ $condition = '<=' ]] && ((diff >= 0)) && return 0
  [[ $condition = '>' ]] && ((diff < 0)) && return 0
  [[ $condition = '>=' ]] && ((diff <= 0)) && return 0
  return 1
}

# Make sure skaffold is installed
command -v skaffold >/dev/null 2>&1 || {
  echo "Command \'skaffold\' not installed. PLease install using the instructions here: https://skaffold.dev/docs/install/"
  exit 1
}

# Check skaffold version
ACTUAL_SKAFFOLD_VERSION=$(skaffold version | tr -d '[:space:]')
ACTUAL_SKAFFOLD_VERSION=${ACTUAL_SKAFFOLD_VERSION#"v"}
MINIMUM_SKAFFOLD_VERSION=$(cat $ROOT_DIR/.skaffold-version)
IS_ACTUAL_GT_MINIMUM=$(vercmp "$ACTUAL_SKAFFOLD_VERSION" "$MINIMUM_SKAFFOLD_VERSION" ">=" && printf 'true' || printf 'false')
if [ $IS_ACTUAL_GT_MINIMUM != 'true' ]; then
  echo "Skaffold version ($ACTUAL_SKAFFOLD_VERSION) is less than minimum version ($MINIMUM_SKAFFOLD_VERSION). PLease install using the instructions here: https://skaffold.dev/docs/install/"
  exit 1
fi

# Install helm using helmenv
command -v helmenv >/dev/null 2>&1 && helmenv install || true

# Check helm is installed
command -v helm >/dev/null 2>&1 || {
  echo "Command \'helm\' not installed. PLease install using the instructions here: https://helm.sh/docs/intro/install/"
  exit 1
}

# Check helm version
ACTUAL_HELM_VERSION=$(helm version --short| tr -d '[:space:]')
ACTUAL_HELM_VERSION=${ACTUAL_HELM_VERSION#"v"}
ACTUAL_HELM_VERSION=${ACTUAL_HELM_VERSION%"+*"}
MINIMUM_HELM_VERSION=$(cat $ROOT_DIR/.helm-version)
IS_ACTUAL_GT_MINIMUM=$(vercmp "$ACTUAL_HELM_VERSION" "$MINIMUM_HELM_VERSION" ">=" && printf 'true' || printf 'false')
if [ $IS_ACTUAL_GT_MINIMUM != 'true' ]; then
  echo "Helm version ($ACTUAL_HELM_VERSION) is less than minimum version ($MINIMUM_HELM_VERSION). PLease install using the instructions here: https://helm.sh/docs/intro/install/"
  exit 1
fi

# # Install helm repos
# helm repo add stable https://kubernetes-charts.storage.googleapis.com

# # Create local namespaces
# # create_ns="kubectl create ns"
# ${create_ns} wallclock || true
