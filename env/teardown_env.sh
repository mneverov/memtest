#!/usr/bin/env bash

if kind delete cluster --name "memtest"; then
  echo "Test cluster deleted successfully"
  rm -f "kubeconfig.yaml"
else
  echo "Test cluster deletion failed"
  exit 1
fi
