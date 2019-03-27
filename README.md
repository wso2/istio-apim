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

### Deploy wso2 adapter as a cluster service

#### Prerequisites

- Istio 
- WSO2 API Manager 2.5.0 or above 
- Public certificate of WSO2 API Manager

Notes: 

- The docker image of the WSO2 mixer adapter is available in the docker hub.
- In the default profile of Istio installation, policy check is disabled by default. To use the mixer adapter, policy check has to enable explictly. Please follow [Enable Policy Enforcement](https://istio.io/docs/tasks/policy-enforcement/enabling-policy/)

##### 1. Create a K8s secret in istio-system namespace for the public certificate of WSO2 API Manager as follows.
```
kubectl create secret generic server-cert --from-file=./server.pem -n istio-system

Note: Save the WSO2 server certificate in PEM format in a file called server.pem
```
##### 2. Deploy the wso2-adapter as a cluster service
```
kubectl create -f samples/adapter-artifacts/
```
##### 3. Create and publish an API in WSO2 API Manager Publisher

If you are exposing a service called httpbin, you can create and publish an API with the name httpbinAPI 
in WSO2 API Manager.

##### 4. Update the API name, version and service for subscription validation

```
Open the samples/api.yaml and update the following.

- api.service
- api.version
- service 

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
curl http://${INGRESS_GATEWAY_URL}/ -H "Authorization: Bearer ACCESS_TOKEN"
```

##### 8. Cleanup


```
kubectl delete -f samples/ -R
kubectl delete secrets server-cert -n istio-system
```