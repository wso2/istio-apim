# Introduction

WSO2 API Manager is a full lifecycle API Management solution which has an API Gateway and a Microgateway. Istio is a service mesh solution which helps users to deploy and manage a collection of microservices. Service meshes in their native form have an “API Management Gap” that requires to be filled. These are related to exposing services to external consumers (advanced security, discovery, governance, etc), business insights, policy enforcement, and monetization. This explains how WSO2 API Manager plans to integrate with Istio and manage services deployed in Istio as APIs. 

# Background

When users move towards microservice architecture from monolithic app architecture, it can result in a considerable number of fine-grained microservices. So, it was a challenge to manage all these microservices. As a solution, Istio was able to provide a platform to connect, manage and secure all these microservices while reducing the complexity of deployments. In addition, Istio includes APIs that let it integrate into any logging platform, or telemetry or policy system.

However when users need to expose these microservices to outside in a secured controlled manner API Management comes in to picture. Most of the time we need to create APIs (for microservices) and share them with other developers who might be part of your organization or external. So API Management within service mesh solution is required to operate successfully. With this capability, the user can expose one or more services from an Istio service mesh as APIs by adding API management capabilities. 

# Approach

When it comes to enabling API Management for Istio, the first iteration of this solution will be designed in such a way that the control plane of the service mesh communicates with the API Manager for token validation, authentication..etc. The API Manager will be responsible for discovery, policy declaration and enforcement, security token service (STS), rate limiting and business insights. The Istio Mixer will be the main point of integration when it comes to run-time security, policy checking, and analytics.

![alt text](https://raw.githubusercontent.com/wso2/istio-apim/master/component_diagram.png)

#### Role of the Istio Mixer plugin

Mixer is a core Istio component which runs in the control plane of the service mesh. Mixer's plugin model enables new rules and policies to be added to groups of services in the mesh without modifying the individual services or the nodes where they run. API management policies such as authentication (by API key validation), rate-limiting, etc can be deployed and managed at API Manager without doing any changes to the actual microservice or sidecar proxy.

#### Create APIs for Services Created

Whenever a user deploys a service, Istio injects a sidecar to the particular service as a proxy. For each request sent to the service, the sidecar proxy will capture a set of data and publish it to the Mixer. If the user needs to expose this service to outside in a managed way, an API should be created in API Manager. This can be done via different methods:

- Automated process - When a user deploys a service which is required to be exposed, an API will be created in API Manager automatically. This can be done via an extension to Kubernetes, using a custom controller which listens for services to be exposed.
- Manual process - Once a service is deployed, the user can go to the API Manager developer portal and create API by giving service data and swagger file.

####  Route of a Successful Request

Let us now see how service calls work with this solution and at which point API related quality of services gets applied. As you can see in the diagram below, when a request comes from outside it first goes to the Istio proxy (Envoy) and then it will communicate with the mixer for performing policy checks. Based on the outcome of the policy checks, the request may be routed to the service or an error should be sent back to the client. Please see the diagram and steps listed below.

![alt text](https://raw.githubusercontent.com/wso2/istio-apim/master/request_flow.png)

1. The client sends the request to the service (Istio capture the request and redirect to the Istio-proxy). This enters the Kubernetes cluster via an ingress point.
2. Proxy captures a wealth of signal and sends to the Mixer as attributes.
3. Mixer adapter then calls the API Manager for various types of policy checks and verifications.
4. API Manager performs the policy checks and responds back to the mixer.
5. Mixer communicates the outcome of the policy checks to the Istio proxy.
6. Since in this case there are no policy validation failures the request is routed to the microservice.
7. The microservice executes the service logic and sends the response.
8. The response is sent out to the client.


---
## Istio mixer adapter for WSO2 API Manager

Using WSO2 adapter, users can validate JWT tokens along with the API subscriptions.

### Installation of the mixer adapter

##### Prerequisites

- [Istio 1.1 or above](https://istio.io/docs/setup/kubernetes/install/) 
- [WSO2 API Manager 2.6.0 or above](https://wso2.com/api-management/)

Notes: 

- The docker image of the WSO2 mixer adapter is available in the docker hub.
- In the default profile of Istio installation, policy check is disabled by default. To use the mixer adapter, policy check has to enable explicitly. Please follow [Enable Policy Enforcement](https://istio.io/docs/tasks/policy-enforcement/enabling-policy/)

##### Enable Istio side car injection for the default namespace 

```
kubectl label namespace default istio-injection=enabled
```

##### Create a K8s secret in istio-system namespace for the public certificate of WSO2 API Manager as follows.

```
kubectl create secret generic server-cert --from-file=./install/server.pem -n istio-system
```

Note: The public certificate of WSO2 API Manager 2.6.0 GA can be found in install/server.pem

##### Deploy the wso2-adapter as a cluster service

```
kubectl apply -f install/
```

### Deploy the httpbin sample


- Deploy httpbin sample 

```
kubectl create -f samples/httpbin/httpbin.yaml
```

- Expose httpbin via Istio ingress gateway to access from outside

```
kubectl create -f samples/httpbin/httpbin-gw.yaml
```

- Access httpbin via Istio ingress gateway

```
curl http://${INGRESS_GATEWAY_IP}/31380/headers
```

### Secure the service and validate subscriptions


##### Create and publish an API in WSO2 API Manager Publisher

Log into WSO2 API Manager publisher and create an API with the following details.

- API Name : httpbinAPI
- API Context : /httpbin
- API Version: 1.0.0 

##### Deploy the API in Istio for subscription validation

```
kubectl create -f samples/httpbin/api.yaml
```

Note: You can map the API with the service mesh service by changing the following values in samples/httpbin/api.yaml

- api.service : name of the API
- api.version : version of the API
- service : mesh service 

##### Deploy the rule to apply the mixer adapter for incoming requests

```
kubectl create -f samples/httpbin/rule.yaml
```

Note: This rule applies for any incoming request in the default namespace. 

##### Create an application in WSO2 API Manager Store, subscribe to the API and generate an access token

- Create an application and select JWT for the Token Type.

- Subscribe to the API httpbinAPI by selecting the application created

- Generate an access token


##### Access the Service 

When accessing the service, provide the authorization header as follows.

```
curl http://${INGRESS_GATEWAY_IP}/31380/headers -H "Authorization: Bearer ACCESS_TOKEN"
```

### Cleanup

```
kubectl delete -f samples/httpbin
kubectl delete -f install/
kubectl delete secrets server-cert -n istio-system
```