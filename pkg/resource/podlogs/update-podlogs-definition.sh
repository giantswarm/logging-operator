#!/bin/bash
#
# This scripts fetches the Go definitions for PodLogs from Grafana Alloy at
# https://github.com/grafana/alloy/tree/main/internal/component/loki/source/podlogs/internal/apis/monitoring/v1alpha2
# and stored them in the apis/ directory.

SCRIPT_DIR=$(cd $(dirname "${BASH_SOURCE}") && pwd -P)

echo "Downloading PodLogs API definitions to $SCRIPT_DIR"

wget -O- "https://api.github.com/repos/grafana/alloy/tarball/$(curl -sS https://api.github.com/repos/grafana/alloy/releases/latest | grep tag_name | cut -d'"' -f4)" | \
tar -C "$SCRIPT_DIR/" -xz --strip-components=7 --wildcards '*/internal/component/loki/source/podlogs/internal/apis/*'

goimports -l -w "$SCRIPT_DIR/apis/"