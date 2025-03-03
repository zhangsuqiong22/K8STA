#!/bin/bash
# Delete testing-related Kubernetes resources
kubectl delete deployment.apps/tester-testing-report-manager -n kubeta
kubectl delete deployment.apps/case-controller -n kubeta
kubectl delete svc tester-testing-report-manager -n kubeta
kubectl delete ns kubeta
kubectl delete clusterrole crd-creator
