# OAuth2 Key Validation with WSO2 API Manager

### Deploy adapter as a cluster service

#### Prerequisites

- Istio 
- WSO2 API Manager deployment in Kubernetes or any accessible deployment from K8s cluster
- Public certificate of WSO2 API Manager

1. Create a K8s secret in istio-system for the public certificate of WSO2 API Manager as follows.

kubectl create secret generic server-cert --from-file=./server.cer.pem -n istio-system

2. Deploy the adapter as a cluster service

kubectl create -f cluster_service.yaml

3. Deploy the adapter artifacts

kubectl apply -f wso2/adapter-artifacts/