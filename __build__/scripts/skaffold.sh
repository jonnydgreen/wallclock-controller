#!/usr/bin/env bash

set -e

# Vars
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
ROOT_DIR=$SCRIPT_DIR/../..
LOG_LEVEL=$1

# Exports
export DEV_CRONJOB=${2:-false}

# Command
(cd $ROOT_DIR && skaffold dev --force=false --verbosity $LOG_LEVEL)
