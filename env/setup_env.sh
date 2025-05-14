#!/usr/bin/env bash

KUBECFG="./env/kubeconfig.yaml"
TEST_ENV_CLUSTER_NAME="memtest"

kind create cluster --name "${TEST_ENV_CLUSTER_NAME}" --kubeconfig "${KUBECFG}" --image kindest/node:v1.32.2 \
  --config "./env/kind-cfg.yaml"
export KUBECONFIG="${KUBECFG}"

helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo add stable https://charts.helm.sh/stable
helm repo update

kubectl create namespace monitoring
kubectl create -f env/configmap-dashboard.yaml

helm install --wait --timeout 5m --namespace monitoring \
  --values ./env/values.yaml \
  --repo https://prometheus-community.github.io/helm-charts kube-prometheus-stack kube-prometheus-stack

echo "Loading image memtest to ${TEST_ENV_CLUSTER_NAME} cluster"
export IMG="registry.k8s.io/memtest:test"
CMD_KIND_LOAD=("kind load docker-image ${IMG} --name ${TEST_ENV_CLUSTER_NAME}")
${CMD_KIND_LOAD} &> /dev/null

kubectl create -f env/pod.yaml
kubectl create -f env/pm.yaml

echo "grafana"
kubectl wait --for=condition=Ready -n monitoring pod -l 'app.kubernetes.io/name=grafana' --timeout=10s
GRAFANA_PF=("kubectl port-forward -n monitoring deployment/kube-prometheus-stack-grafana 3000")
${GRAFANA_PF} & #> /dev/null

echo "prometheus"
kubectl wait --for=condition=Ready -n monitoring pod -l 'app.kubernetes.io/name=prometheus'  --timeout=20s
PROMETHEUS_PF=("kubectl port-forward -n monitoring svc/kube-prometheus-stack-prometheus 9090:9090")
${PROMETHEUS_PF} &> /dev/null

echo "done"

