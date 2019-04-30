## Advanced Guide

By following this guide you can do the followings.

1. Setup WSO2 API Manager 
2. Setup WSO2 API Manager Analytics
3. Deploy Istio Mixer Adapter

### Installation

##### Prerequisites

- [Istio 1.1 or above](https://istio.io/docs/setup/kubernetes/install/) 
- [WSO2 API Manager 2.6.0 or above](https://wso2.com/api-management/)
- [WSO2 API Manager Analytics 2.6.0 or above](https://wso2.com/api-management/)
- [Istio-apim release: wso2am-istio-1.0.zip](https://github.com/wso2/istio-apim/releases/tag/1.0)

**Notes:** 

- The docker image of the WSO2 mixer adapter is available in the [docker hub](https://hub.docker.com/r/wso2/apim-istio-mixer-adapter).
- In the default profile of Istio installation, the policy check is disabled by default. To use the mixer adapter, the policy check has to enable explicitly. Please follow [Enable Policy Enforcement](https://istio.io/docs/tasks/policy-enforcement/enabling-policy/)
- wso2am-istio-1.0.zip contains installation artifacts to deploy in the Istio, WSO2 API Manager and WSO2 API Manager Analytics.

#### Install WSO2 API Manager Analytics

- Copy gRPC and supportive jars to lib directory in the analytics server

```
cp install/analytics/resources/lib/* <WSO2_API_Manager_Analytics_Server>/lib/
```

- Copy updated Siddhi files to siddhi-files directory in the analytics server

```
cp install/analytics/resources/siddhi-files/* <WSO2_API_Manager_Analytics_Server>/wso2/worker/deployment/siddhi-files/
```

- Start WSO2 API Manager Analytics server

**Note:** Make sure WSO2 API Manager Analytics server can be accessible from the K8s cluster

#### Install WSO2 API Manager

- [Enable Analytics](https://docs.wso2.com/display/AM260/Configuring+APIM+Analytics) 

- Copy custom executor to lib directory in the api manager server

```
cp install/api-manager/resources/lib/* <WSO2_API_Manager>/repository/components/lib/
```

- Copy supported osgi bundles to dropins location in api manager server

```
cp install/api-manager/resources/dropins/* <WSO2_API_Manager>/repository/components/dropins/
```
- Copy APILifeCycle.xml resource to life cycles directory in api manager server

```
cp install/api-manager/resources/lifecycles/* <WSO2_API_Manager>/repository/resources/lifecycles/
```

- Create a service account in Kubernetes for API access.

```
kubectl apply -f install/api-manager/k8s-artifacts/rbac.yaml
```

- Retrieve K8s secret name created for the service account

```
kubectl get serviceaccounts wso2svc-account -n wso2 -oyaml
```

Under secrets you can find the secret created when creating the service account.

- Retrieve K8s secret details

```
kubectl get secrets <SECRET_NAME> -n wso2 -oyaml
```

This secret contains "ca.crt" and the "token" which can be used to access the K8s API. Both are base64 for encoded.

- Copy token to WSO2 API Manager and Set K8s API Server url

Decode the token and save it as a satoken.txt. Copy the satoken.txt file to WSO2 API Manager Server home path.
```
cp satoken.txt <WSO2_API_Manager>/
```

Get the K8s API Server URL
```
kubectl cluster-info
```

Change the kubernetesAPIServerUrl value in APILifeCycle.xml in WSO2_API_Manager/repository/resources/lifecycles/ location. 
Uncomment saTokenFileName parameter APILifeCycle.xml. Do these changes in all the places in APILifeCycle.xml.

- Add ca.crt to client-truststore of WSO2 API Manager

Decode the ca.crt and save it as ca.crt file. Insert the certificate using the keytool as below.

```
keytool -import -trustcacerts -alias carbon -file ca.crt -keystore 
<WSO2_API_Manager>/repository/resources/security/client-truststore.jks -storepass wso2carbon
```

- Start WSO2 API Manager server

**Note:** Make sure WSO2 API Manager server can be accessible from the K8s cluster

#### Install WSO2 Istio Mixer Adapter

- Create a K8s secret in istio-system namespace for the public certificate of WSO2 API Manager as follows.

```
kubectl create secret generic server-cert --from-file=./install/adapter-artifacts/server.pem -n istio-system
```

**Note:** The public certificate of WSO2 API Manager 2.6.0 GA can be found in install/adapter-artifacts/server.pem.

- Update WSO2 API Manager URLs and credentials

Update API Manager URLs and credentials for OAuth2 token validation - install/adapter-artifacts/wso2-adapter.yaml

```
apim-url: https://wso2-apim:9443      
server-token: YWRtaW46YWRtaW4=  (Base 64 encoded username:password)
```

**Note:** Hostname verification is disabled by default for OAuth2 token validation service call. You can enable by changing the config in install/adapter-artifacts/wso2-operator-config.yaml
```
disable_hostname_verification: "false"
```

- Update API Manager Analytics endpoints for gRPC event publishing for analytics - install/adapter-artifacts/wso2-operator-config.yaml

```
request_stream_app_url: "wso2-apim:7575"             
fault_stream_app_url: "wso2-apim:7576"               
throttle_stream_app_url: "wso2-apim:7577"
```

**Note:** 7575, 7576, 7577 are gRPC ports used for data publishing.

- Deploy the wso2-adapter as a cluster service

```
kubectl apply -f install/adapter-artifacts/
```

### Cleanup

```
kubectl delete -f install/adapter-artifacts/
kubectl delete secrets server-cert -n istio-system
kubectl delete serviceaccounts wso2svc-account -n wso2
```
