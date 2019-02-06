# OAuth2 Key Validation with WSO2 API Manager

### Deploy adapter as a cluster service

#### Prerequisites

- Istio 
- WSO2 API Manager deployment in Kubernetes or any accessible deployment from K8s cluster
- Public certificate of WSO2 API Manager

Note: Docker image is available in the docker hub.

##### 1. Create a K8s secret in istio-system for the public certificate of WSO2 API Manager as follows.
```
kubectl create secret generic server-cert --from-file=./server.cer.pem -n istio-system
```
##### 2. Deploy the adapter as a cluster service
```
kubectl create -f cluster_service.yaml
```
##### 3. Deploy the adapter artifacts
```
kubectl apply -f wso2/adapter-artifacts/
```

## Developer Guide

This guide is to create the adapter for key validation.

##### 1. Clone WSO2 Istio-apim repo and setup environment variables

```
git clone https://github.com/wso2/istio-apim.git
cd istio-apim/oauth2-key-validation-adapter

mkdir src
export ROOT_FOLDER=`pwd`
export GOPATH=`pwd`
export MIXER_REPO=$GOPATH/src/istio.io/istio/mixer
export ISTIO=$GOPATH/src/istio.io
```

##### 2. Clone the Istio source code

```
mkdir -p $GOPATH/src/istio.io/
cd $GOPATH/src/istio.io/
git clone https://github.com/istio/istio
```

##### 3. Build mixer server,client binary

```
pushd $ISTIO/istio && make mixs
pushd $ISTIO/istio && make mixc
```

##### 4. Setup the wso2 adapter and copy the Configuration .proto file

This file contains the runtime parameters.

```
mkdir -p $MIXER_REPO/adapter/wso2/config
cp $ROOT_FOLDER/wso2/config/config.proto $MIXER_REPO/adapter/wso2/config/config.proto
```

##### 5. Copy Adapter implementation source code and build

wso2.go file contains handler business logic.

```
cp $ROOT_FOLDER/wso2/wso2.go $MIXER_REPO/adapter/wso2/wso2.go
cp $ROOT_FOLDER/wso2/keyValidationHandler.go $MIXER_REPO/adapter/wso2/keyValidationHandler.go
cd $MIXER_REPO/adapter/wso2

go generate ./...
go build ./...
```

##### 6. Copy the generated files


```
mkdir -p $MIXER_REPO/adapter/wso2/adapter-artifacts
cp $MIXER_REPO/adapter/wso2/config/wso2.yaml $MIXER_REPO/adapter/wso2/adapter-artifacts
cp $MIXER_REPO/testdata/config/attributes.yaml $MIXER_REPO/adapter/wso2/adapter-artifacts
cp $MIXER_REPO/template/authorization/template.yaml $MIXER_REPO/adapter/wso2/adapter-artifacts
cp $ROOT_FOLDER/wso2/adapter-artifacts/wso2_operator_cfg.yaml $MIXER_REPO/adapter/wso2/adapter-artifacts
```

Note: attributes.yaml and template.yaml is taken from the Istio repository. wso2_operator_cfg is taken from istio-apim repo.

##### 7. Create Adapter Starter

This app launches the adapter gRPC server:

```
mkdir -p $MIXER_REPO/adapter/wso2/cmd
cp  $ROOT_FOLDER/wso2/cmd/main.go $MIXER_REPO/adapter/wso2/cmd/
```

##### 8. Create a Adapter docker image

```
cd $ROOT_FOLDER

docker build -t wso2/wso2adapter:v1 .
```

Push this to a docker registry which can be accessed from the Kubernetes cluster.

##### 9. Create a K8s secret in istio-system for the public certificate of WSO2 API Manager as follows.
```
kubectl create secret generic server-cert --from-file=./server.cer.pem -n istio-system
```
##### 10. Deploy the adapter as a cluster service
```
kubectl create -f $ROOT_FOLDER/cluster_service.yaml
```
##### 11. Deploy the adapter artifacts
```
kubectl apply -f $MIXER_REPO/adapter/wso2/adapter-artifacts/
```