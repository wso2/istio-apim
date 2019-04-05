# Introduction

WSO2 API Manager is a full lifecycle API Management solution which has an API Gateway and a Microgateway. Istio is a service mesh solution which helps users to deploy and manage a collection of microservices. Service meshes in their native form have an “API Management Gap” that requires to be filled. These are related to exposing services to external consumers (advanced security, discovery, governance, etc), business insights, policy enforcement, and monetization. This explains how WSO2 API Manager plans to integrate with Istio and manage services deployed in Istio as APIs. 

# Background

When users move towards microservice architecture from monolithic app architecture, it can result in a considerable number of fine-grained microservices. So, it was a challenge to manage all these microservices. As a solution, Istio was able to provide a platform to connect, manage and secure all these microservices while reducing the complexity of deployments. In addition, Istio includes APIs that let it integrate into any logging platform, or telemetry or policy system.

However, when users need to expose these microservices to outside in a secured controlled manner API Management comes in to picture. Most of the time we need to create APIs (for microservices) and share them with other developers who might be part of your organization or external. So API Management within service mesh solution is required to operate successfully. With this capability, the user can expose one or more services from an Istio service mesh as APIs by adding API management capabilities. 

# Approach

While Istio providing Data Plane and Control Plane capabilities, WSO2 API Manager provides Manage Plane capabilities to manage microservices.

![alt text](https://raw.githubusercontent.com/wso2/istio-apim/master/component_diagram.png)

#### Role of the Istio Mixer plugin

The mixer is a core Istio component which runs in the control plane of the service mesh. Mixer's plugin model enables new rules and policies to be added to groups of services in the mesh without modifying the individual services or the nodes where they run. API management policies such as authentication (by API key validation), rate-limiting, etc can be deployed and managed at API Manager without doing any changes to the actual microservice or sidecar proxy.

#### API Management for Istio

When need to expose this service to outside in a managed way, API developer can use WSO2 API Publisher portal to create the API by attaching necessary policies like security, rate limiting etc. The Publisher is capable of pushing all these policies into Envoy proxy via Pilot and Mixer for them to take action of policy enforcement. After publishing this API, it will appear in the WSO2 API Developer portal. Now app developer can discover these APIs and use in their application along with all the capabilities provided by developer portal like getting a subscription plan, adding application security etc. The business user can use API Analytics to get more business insights by looking at API Analytics.

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

Using WSO2 adapter, users can do the following.

- Secure service with JWT and OAuth2 tokens
- Validate API subscriptions
- Validate scopes

### Installation of the mixer adapter

##### Prerequisites

- [Istio 1.1 or above](https://istio.io/docs/setup/kubernetes/install/) 
- [WSO2 API Manager 2.6.0 or above](https://wso2.com/api-management/)
- [Istio-apim release: wso2am-istio-0.6.zip](https://github.com/wso2/istio-apim/releases/tag/0.6)

Notes: 

- The docker image of the WSO2 mixer adapter is available in the [docker hub](https://hub.docker.com/r/wso2/apim-istio-mixer-adapter).
- In the default profile of Istio installation, the policy check is disabled by default. To use the mixer adapter, policy check has to enable explicitly. Please follow [Enable Policy Enforcement](https://istio.io/docs/tasks/policy-enforcement/enabling-policy/)
- wso2am-istio-0.6.zip contains artifacts to deploy in the Istio.

##### Enable Istio side car injection for the default namespace 

```
kubectl label namespace default istio-injection=enabled
```

##### Create a K8s secret in istio-system namespace for the public certificate of WSO2 API Manager as follows.

```
kubectl create secret generic server-cert --from-file=./install/server.pem -n istio-system
```

*Note:* The public certificate of WSO2 API Manager 2.6.0 GA can be found in install/server.pem. Using this server certificate, you can do the JWT token validation. 
If you want to do the OAuth2 token validation, then deploy WSO2 API Manager in K8s or any accessible location. Use that certificate to create the secret.

##### Deploy the wso2-adapter as a cluster service

```
kubectl apply -f install/
```

*Note:* If you want to use OAuth2 token validation, then update the apim-url and server-token of the WSO2 API Manager in install/wso2-adapter.yaml file.

Sample values: 

apim-url: https://wso2apim-with-analytics-apim-service.wso2.svc:9443      
server-token: YWRtaW46YWRtaW4=  (Base 64 encoded username:password)

### Deploy a microservice in Istio

- Deploy httpbin sample service

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

### Apply API Management for microservices

We are going to secure the service and this can be done with OAuth2 tokens or JWT tokens. Also, do the subscription validation for the API and scope validation for the resources.

##### Create and publish an API in WSO2 API Manager Publisher

Log into WSO2 API Manager publisher and create an API with the following details.

- API Name : HttpbinAPI
- API Context : /httpbin
- API Version: 1.0.0 

Add the following resources with these scopes.

| Resource        | Scope            | 
|:--------------- |:---------------- |
| /ip             | scope_ip         | 
| /headers        | scope_headers    |  


##### Bind the API to the service for subscription validation and scope validation.

```
kubectl create -f samples/httpbin/api.yaml
```

*Note:* You can map the API with the service mesh service by changing the following values in samples/httpbin/api.yaml

- api.service : name of the API              # Used in JWT verification
- api.version : version of the API           # Used in JWT verification and OAuth2 verification
- api.context : context of the API           # Used in OAuth2 verification
- resource.scope : scope of the resource     # Used in JWT verification 
- service : mesh service 

The above values are used in the following verifications.

| Attribute Value | Use Case         | 
|:--------------- |:---------------- |
| api.service     | JWT              | 
| api.version     | JWT and OAuth2   | 
| api.context     | OAuth2           | 
| resource.scope  | JWT              | 

##### Deploy the rule to apply the mixer adapter for incoming requests

```
kubectl create -f samples/httpbin/rule.yaml
```

*Note:* This rule applies for any incoming request in the default namespace. 

##### Access the Service

1.) Using JWT Tokens

- Create an application by selecting JWT for the Token Type.

- Subscribe to the API httpbinAPI by selecting the application created

- When generating the token, select the relevant scopes and generate an access token


When accessing the service, provide the authorization header as follows.

```
curl http://${INGRESS_GATEWAY_IP}/31380/headers -H "Authorization: Bearer JWT_ACCESS_TOKEN"
```

2.) Using OAuth2 Tokens

- Create an application by selecting OAuth2 for the Token Type.

- Subscribe to the API httpbinAPI by selecting the application created

- When generating the token, select the relevant scopes and generate an access token

When accessing the service, provide the authorization header as follows.

```
curl http://${INGRESS_GATEWAY_IP}/31380/headers -H "Authorization: Bearer OAuth2_ACCESS_TOKEN"
```

### Cleanup

```
kubectl delete -f samples/httpbin
kubectl delete -f install/
kubectl delete secrets server-cert -n istio-system
```