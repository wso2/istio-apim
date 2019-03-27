# Istio Mixer adapter for WSO2 API Manager

Using WSO2 adapter, users can validate JWT tokens along with the API subscriptions.

### Deploy wso2 adapter as a cluster service

#### Prerequisites

- Istio 
- WSO2 API Manager running in anywhere
- Public certificate of WSO2 API Manager

Note: Docker image is available in the docker hub.

##### 1. Create a K8s secret in istio-system namespace for the public certificate of WSO2 API Manager as follows.
```
kubectl create secret generic server-cert --from-file=./server.pem -n istio-system
```
##### 2. Deploy the wso2-adapter as a cluster service
```
kubectl create -f samples/adapter-artifacts/
```
##### 3. [Create and publish](https://docs.wso2.com/display/AM260/Create+and+Publish+an+API) an API in WSO2 API Manager Publisher

If you are exposing a service called httpbin, you can create and publish an API with the name httpbinAPI 
in WSO2 API Manager.

##### 4. Update the API name, version and service for subscription validation

```
Open the samples/api.yaml and update api name, version and the service mesh service which needs to map the API and the service.

Then deploy the api as follows.

kubectl create -f samples/api.yaml

```

##### 5. Deploy the rule to apply the mixer adapter for incoming requests

```
This rule applies for any incoming request in the default namespace. 

kubectl create -f samples/rule.yaml
```

##### 6. Create an application in WSO2 API Manager Store, subscribe to the API and generate an access token

```
- Create an application and select JWT for the Token Type.
- Subscribe to the API by selecting the application recreated
- Generate an access token
```

##### 7. Access the Service 

When accessing the service, provide a header as follows.

```
Authorization: Bearer ACCESS_TOKEN
```


## Developer Guide

This guide is to create the adapter for key validation.

##### 1. Clone WSO2 Istio-apim repo and setup environment variables

```
git clone https://github.com/wso2/istio-apim.git
cd istio-apim/adapter

mkdir src
export ROOT_FOLDER=`pwd`
export GOPATH=`pwd`
export MIXER_REPO=$GOPATH/src/istio.io/istio/mixer
export ISTIO=$GOPATH/src/istio.io
```

##### 2. Clone the Istio source code & checkout to 1.1.0 version

```
mkdir -p $GOPATH/src/istio.io/
cd $GOPATH/src/istio.io/
git clone https://github.com/istio/istio
cd $ISTIO/istio
git checkout 1.1.0
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
cp $ROOT_FOLDER/wso2/jwtValidationHandler.go $MIXER_REPO/adapter/wso2/jwtValidationHandler.go
cd $MIXER_REPO/adapter/wso2

go generate ./...
go build ./...
```

##### 6. Copy adapter artifacts

```
mkdir -p $MIXER_REPO/adapter/wso2/adapter-artifacts
cp $MIXER_REPO/adapter/wso2/config/wso2.yaml $MIXER_REPO/adapter/wso2/adapter-artifacts
cp $MIXER_REPO/testdata/config/attributes.yaml $MIXER_REPO/adapter/wso2/adapter-artifacts
cp $MIXER_REPO/template/authorization/template.yaml $MIXER_REPO/adapter/wso2/adapter-artifacts
cp $ROOT_FOLDER/../samples/adapter-artifacts/wso2-operator-config.yaml $MIXER_REPO/adapter/wso2/adapter-artifacts
cp $ROOT_FOLDER/../samples/adapter-artifacts/wso2-adapter.yaml $MIXER_REPO/adapter/wso2/adapter-artifacts
```

Note: attributes.yaml and template.yaml is taken from the Istio repository. wso2-operator-config.yaml and wso2-adapter is taken from istio-apim repo.

##### 7. Create Adapter Starter

This app launches the adapter gRPC server:

```
mkdir -p $MIXER_REPO/adapter/wso2/cmd
cp  $ROOT_FOLDER/wso2/cmd/main.go $MIXER_REPO/adapter/wso2/cmd/
```

##### 8. Create a Adapter docker image

```
cd $ROOT_FOLDER

docker build -t pubudu/wso2adapter:v4 .
```

Push this to a docker registry which can be accessed from the Kubernetes cluster.

##### 9. Create a K8s secret in istio-system for the public certificate of WSO2 API Manager as follows.

```
kubectl create secret generic server-cert --from-file=./server.pem -n istio-system
```

##### 10. Deploy the wso2-adapter as a cluster service

```
kubectl create -f $MIXER_REPO/adapter/wso2/adapter-artifacts/
```

##### 11. Deploy the wso2-adapter as a cluster service

```
kubectl create -f $MIXER_REPO/adapter/wso2/adapter-artifacts/
```

##### 12. Deploy the api and the rule for the service

```
kubectl create -f $ROOT_FOLDER/../samples/api.yaml
kubectl create -f $ROOT_FOLDER/../samples/rule.yaml
```