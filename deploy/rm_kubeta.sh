#!/bin/bash

kubectl delete -f kubeta_deployment.yaml
kubectl delete ns kubetatester
kubectl delete crd testers.mytester.kubeta.github.io
kubectl delete clusterroles crd-creator
kubectl delete clusterrolebindings crd-creator 

