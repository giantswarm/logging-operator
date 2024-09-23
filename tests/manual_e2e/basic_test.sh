#!/bin/bash

# Helper function - prints an error message and exits
exit_error() {
  echo "Error: $*"
  exit 1
}

# Helper function - clean up the WC
clean_wc() {
  kubectl delete -f grizzly-e2e-wc.yaml
  rm grizzly-e2e-wc.yaml
}

# Helper function - checks the status of the daemonset
check_daemonset_status() {
  local daemonsetDesiredReplicas=$(kubectl get daemonset -n kube-system --context teleport.giantswarm.io-$1-loggingoperatortest $2 -o yaml | yq .status.desiredNumberScheduled)
  local daemonsetReadyReplicas=$(kubectl get daemonset -n kube-system --context teleport.giantswarm.io-$1-loggingoperatortest $2 -o yaml | yq .status.numberReady)

  [[ "$daemonsetDesiredReplicas" != "$daemonsetReadyReplicas" ]] \
    && echo "$2 app deployed but some daemonset's pods aren't in a ready state" || echo "$2 app is deployed and all daemonset's pods are ready"
}

# Helper function - checks the existence of the cm and secret for either alloy or prometheus-agent
check_configs() {
  echo "Checking if the corresponding $1-$2 has been created"
  local config

  [[ "$2" == "config" ]] \
    && config=$(kubectl get configmap -n org-giantswarm loggingoperatortest-$1-$2) || config=$(kubectl get secret -n org-giantswarm loggingoperatortest-$1-$2)

  [[ -z "$config" ]] && echo "$1-$2 not found" || echo "$1-$2 found. Test succeeded"
}

main() {
  [[ -z "$1" ]] && exit_error "Please provide the installation name as an argument"

  # Logging into the specified installation to perform the tests
  tsh kube login $1

  echo "Checking if logging-operator app is in deployed state"

  deployed=$(kubectl get app -n giantswarm logging-operator -o yaml | yq .status.release.status)

  [[ "$deployed" != "deployed" ]] \
    && exit_error "logging-operator app is not in deployed state. Please fix the app before retrying" || echo "logging-operator app is indeed in deployed state"

  echo "Creating WC"

  # Getting latest Giant Swarm release version
  toUseRelease="$(kubectl gs get releases -o template='{{range .items}}{{.status.ready}}/{{.metadata.name}}
  {{end}}' | sed -ne 's/false\/aws-//p' | sort -V | tail -1)"

  # Creating WC template and applying it
  kubectl gs template cluster --provider capa --name loggingoperatortest --organization giantswarm --description "logging-operator e2e tests" --release $toUseRelease > grizzly-e2e-wc.yaml
  kubectl create -f grizzly-e2e-wc.yaml

  echo "WC named 'loggingoperatortest' created. Waiting for it to be ready"

  sleep 1200

  echo "Checking if the logging agent is up and running on the WC"

  # Logging into the WC to get the context into the kubeconfig
  tsh kube login $1-loggingoperatortest
  tsh kube login $1

  promtail=$(kubectl get apps -n org-giantswarm | grep loggingoperatortest-promtail)
  alloy=$(kubectl get apps -n org-giantswarm | grep loggingoperatortest-alloy-logs)

  if [[ ! -z "$promtail" ]]; then
    check_daemonset_status $1 "promtail"
  elif [[ ! -z "$alloy" ]]; then
    check_daemonset_status $1 "alloy-logs"
  else
    echo "No metrics agent app found. Cleaning the WC"
    clean_wc
    exit 1
  fi

  
  check_configs "grafana-agent" "config"
  check_configs "grafana-agent" "secret"
  check_configs "logging" "config"
  check_configs "logging" "secret"

  echo "cleaning WC"

  clean_wc
}

main "$@"