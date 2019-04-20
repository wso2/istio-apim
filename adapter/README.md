## Developer Guide

This guide is to create the adapter for token validation.

##### Prerequisites

- Docker
- Go v1.11
- Protoc

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

-----------
mkdir -p $GOPATH/src/github.com/golang/
cd $GOPATH/src/github.com/golang/
git clone https://github.com/golang/protobuf/

mkdir -p $GOPATH/src/github.com/processout/
cd $GOPATH/src/github.com/processout/
git clone https://github.com/processout/grpc-go-pool

mkdir -p $GOPATH/src/github.com/processout/
cd $GOPATH/src/github.com/processout/
git clone https://github.com/processout/grpc-go-pool

golang.org/x/net/context
---------

##### 3. Build mixer server,client binary

```
pushd $ISTIO/istio && make mixs
pushd $ISTIO/istio && make mixc
```

##### 4. Setup the wso2 adapter and copy the Configuration .proto file

This file contains the runtime parameters.

```
mkdir -p $MIXER_REPO/adapter/wso2/config
mkdir -p $ROOT_FOLDER/src/org.wso2.apim.grpc.telemetry.receiver.generated
cp $ROOT_FOLDER/wso2/config/config.proto $MIXER_REPO/adapter/wso2/config/config.proto
cp $ROOT_FOLDER/../analytics/core/org.wso2.apim.proto/src/main/proto/ReportService.proto $ROOT_FOLDER/src/org.wso2.apim.grpc.telemetry.receiver.generated/
```

##### 5. Copy Adapter implementation source code and build

wso2.go file contains handler business logic.

```
cp $ROOT_FOLDER/wso2/wso2.go $MIXER_REPO/adapter/wso2/wso2.go
cp $ROOT_FOLDER/wso2/jwtValidationHandler.go $MIXER_REPO/adapter/wso2/jwtValidationHandler.go
cp $ROOT_FOLDER/wso2/oauth2ValidationHandler.go $MIXER_REPO/adapter/wso2/oauth2jwtValidationHandler.go
cp $ROOT_FOLDER/wso2/analyticsHandler.go $MIXER_REPO/adapter/wso2/analyticsHandler.go
cd $MIXER_REPO/adapter/wso2

protoc -I $ROOT_FOLDER/src/org.wso2.apim.grpc.telemetry.receiver.generated/ -I${GOPATH}/src --go_out=plugins=grpc:$ROOT_FOLDER/src/org.wso2.apim.grpc.telemetry.receiver.generated/ $ROOT_FOLDER/src/org.wso2.apim.grpc.telemetry.receiver.generated/ReportService.proto


go generate ./...
go build ./...
```

##### 6. Copy adapter artifacts

```
mkdir -p $MIXER_REPO/adapter/wso2/install
cp $MIXER_REPO/adapter/wso2/config/wso2.yaml $MIXER_REPO/adapter/wso2/install
cp $ROOT_FOLDER/../install/attributes.yaml $MIXER_REPO/adapter/wso2/install
cp $MIXER_REPO/template/authorization/template.yaml $MIXER_REPO/adapter/wso2/install/auth-template.yaml
cp $MIXER_REPO/template/metric/template.yaml $MIXER_REPO/adapter/wso2/install/metric-template.yaml
cp $ROOT_FOLDER/../install/wso2-operator-config.yaml $MIXER_REPO/adapter/wso2/install
cp $ROOT_FOLDER/../install/wso2-adapter.yaml $MIXER_REPO/adapter/wso2/install
```

Note: template.yaml is taken from the Istio repository. attributes, wso2-operator-config.yaml and wso2-adapter is taken from istio-apim repo.

##### 7. Create Adapter Starter

This app launches the adapter gRPC server:

```
mkdir -p $MIXER_REPO/adapter/wso2/cmd
cp $ROOT_FOLDER/wso2/cmd/main.go $MIXER_REPO/adapter/wso2/cmd/
```

##### 8. Create a Adapter docker image

```
cd $ROOT_FOLDER

docker build -t pubudu/apim-istio-mixer-adapter:0.8 .
```

Note: Push this docker image to a docker registry which can be accessed from the Kubernetes cluster.

##### 9. Create a K8s secret in istio-system for the public certificate of WSO2 API Manager as follows.

```
kubectl create secret generic server-cert --from-file=$ROOT_FOLDER/../install/server.pem -n istio-system
```

##### 10. Deploy the wso2-adapter as a cluster service

```
kubectl apply -f $MIXER_REPO/adapter/wso2/install/
kubectl delete -f $MIXER_REPO/adapter/wso2/install/
```

##### 11. Deploy the api and the rule for the service

```
kubectl create -f $ROOT_FOLDER/../samples/httpbin/api.yaml
kubectl create -f $ROOT_FOLDER/../samples/httpbin/rule.yaml
```