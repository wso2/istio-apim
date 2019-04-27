/*
 *  Copyright (c) 2019 WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
 *
 *  WSO2 Inc. licenses this file to you under the Apache License,
 *  Version 2.0 (the "License"); you may not use this file except
 *  in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing,
 *  software distributed under the License is distributed on an
 *  "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 *  KIND, either express or implied.  See the License for the
 *  specific language governing permissions and limitations
 *  under the License.
 *
 */

package org.wso2.carbon.apimgt.istio;

import io.fabric8.kubernetes.client.Config;
import io.fabric8.kubernetes.client.ConfigBuilder;
import io.fabric8.kubernetes.client.DefaultKubernetesClient;
import io.fabric8.kubernetes.client.KubernetesClient;
import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;
import org.apache.solr.common.StringUtils;
import org.bouncycastle.util.Strings;
import org.json.simple.JSONArray;
import org.json.simple.JSONObject;
import org.json.simple.parser.JSONParser;
import org.json.simple.parser.ParseException;
import org.wso2.carbon.apimgt.api.APIManagementException;
import org.wso2.carbon.apimgt.api.APIProvider;
import org.wso2.carbon.apimgt.api.FaultGatewaysException;
import org.wso2.carbon.apimgt.api.model.API;
import org.wso2.carbon.apimgt.api.model.APIIdentifier;
import org.wso2.carbon.apimgt.api.model.Tier;
import org.wso2.carbon.apimgt.api.model.URITemplate;
import org.wso2.carbon.apimgt.impl.APIConstants;
import org.wso2.carbon.apimgt.impl.APIManagerFactory;
import org.wso2.carbon.apimgt.impl.dao.ApiMgtDAO;
import org.wso2.carbon.apimgt.impl.internal.ServiceReferenceHolder;
import org.wso2.carbon.apimgt.impl.utils.APIUtil;
import org.wso2.carbon.apimgt.impl.utils.APIVersionComparator;
import org.wso2.carbon.context.CarbonContext;
import org.wso2.carbon.context.PrivilegedCarbonContext;
import org.wso2.carbon.governance.api.generic.GenericArtifactManager;
import org.wso2.carbon.governance.api.generic.dataobjects.GenericArtifact;
import org.wso2.carbon.governance.registry.extensions.aspects.utils.LifecycleConstants;
import org.wso2.carbon.governance.registry.extensions.interfaces.Execution;
import org.wso2.carbon.registry.core.Registry;
import org.wso2.carbon.registry.core.Resource;
import org.wso2.carbon.registry.core.exceptions.RegistryException;
import org.wso2.carbon.registry.core.jdbc.handlers.RequestContext;
import org.wso2.carbon.user.api.UserStoreException;
import org.wso2.carbon.utils.FileUtil;
import org.wso2.carbon.utils.multitenancy.MultitenantConstants;
import org.wso2.carbon.utils.multitenancy.MultitenantUtils;

import java.net.MalformedURLException;
import java.net.URL;
import java.util.List;
import java.util.Map;
import java.util.Set;

import io.fabric8.kubernetes.api.model.ObjectMeta;
import io.fabric8.kubernetes.api.model.apiextensions.CustomResourceDefinition;
import io.fabric8.kubernetes.api.model.apiextensions.CustomResourceDefinitionList;
import io.fabric8.kubernetes.client.dsl.NonNamespaceOperation;
import org.wso2.carbon.apimgt.istio.crd.*;

import java.io.IOException;
import java.util.*;

/**
 * Sends email to the technical owner of the api about a state change
 */
public class IstioExecutor implements Execution {

    private String istioSystemNamespace;
    private String appNamespace;
    private String kubernetesAPIServerUrl;
    private String saTokenFileName;
    private String saTokenFilePath;
    private String kubernetesServiceDomain;

    private static final String ISTIO_SYSTEM_NAMESPACE = "istio-system";
    private static final String APP_NAMESPACE = "default";
    private static final String K8S_MASTER_URL = "https://kubernetes.default";
    private static final String SATOKEN_FILE_PATH = "./";
    private static final String K8S_SERVICE_DOMAIN = "svc.cluster.local";

    private CustomResourceDefinition ruleCRD;
    private CustomResourceDefinition httpAPISpecCRD;
    private CustomResourceDefinition httpAPISpecBindingCRD;

    private static final String TRY_KUBE_CONFIG = "kubernetes.auth.tryKubeConfig";
    private static final String TRY_SERVICE_ACCOUNT = "kubernetes.auth.tryServiceAccount";
    private KubernetesClient client;
    public static String CRD_GROUP = "config.istio.io";
    public static String HTTPAPISpecBinding_CRD_NAME = "httpapispecbindings." + CRD_GROUP;
    public static String HTTPAPISpec_CRD_NAME = "httpapispecs." + CRD_GROUP;
    public static String RULE_CRD_NAME = "rules." + CRD_GROUP;

    Log log = LogFactory.getLog(IstioExecutor.class);


    /**
     * @param parameterMap Map of parameters which passes to the executor
     */
    public void init(Map parameterMap) {

        if (parameterMap != null) {

            istioSystemNamespace = (String) parameterMap.get("istioSystemNamespace");
            appNamespace = (String) parameterMap.get("appNamespace");
            kubernetesAPIServerUrl = (String) parameterMap.get("kubernetesAPIServerUrl");
            saTokenFileName = (String) parameterMap.get("saTokenFileName");
            saTokenFilePath = (String) parameterMap.get("saTokenFilePath");
            kubernetesServiceDomain = (String) parameterMap.get("kubernetesServiceDomain");
        }

        if (StringUtils.isEmpty(istioSystemNamespace)) {
            istioSystemNamespace = ISTIO_SYSTEM_NAMESPACE;
        }

        if (StringUtils.isEmpty(appNamespace)) {
            appNamespace = APP_NAMESPACE;
        }

        if (StringUtils.isEmpty(kubernetesAPIServerUrl)) {
            kubernetesAPIServerUrl = K8S_MASTER_URL;
        }

        if (StringUtils.isEmpty(saTokenFilePath)) {
            saTokenFilePath = SATOKEN_FILE_PATH;
        }

        if (StringUtils.isEmpty(kubernetesServiceDomain)) {
            kubernetesServiceDomain = K8S_SERVICE_DOMAIN;
        }
    }

    /**
     * @param context      The request context that was generated from the registry core.
     *                     The request context contains the resource, resource path and other
     *                     variables generated during the initial call.
     * @param currentState The current lifecycle state.
     * @param targetState  The target lifecycle state.
     * @return Returns whether the execution was successful or not.
     */
    public boolean execute(RequestContext context, String currentState, String targetState) {
        boolean executed = false;
        String user = PrivilegedCarbonContext.getThreadLocalCarbonContext().getUsername();
        String domain = CarbonContext.getThreadLocalCarbonContext().getTenantDomain();

        String userWithDomain = user;
        if (!MultitenantConstants.SUPER_TENANT_DOMAIN_NAME.equals(domain)) {
            userWithDomain = user + APIConstants.EMAIL_DOMAIN_SEPARATOR + domain;
        }

        userWithDomain = APIUtil.replaceEmailDomainBack(userWithDomain);

        try {
            String tenantUserName = MultitenantUtils.getTenantAwareUsername(userWithDomain);
            int tenantId = ServiceReferenceHolder.getInstance().getRealmService().getTenantManager().getTenantId(domain);

            GenericArtifactManager artifactManager = APIUtil
                    .getArtifactManager(context.getSystemRegistry(), APIConstants.API_KEY);
            Registry registry = ServiceReferenceHolder.getInstance().
                    getRegistryService().getGovernanceUserRegistry(tenantUserName, tenantId);
            Resource apiResource = context.getResource();
            String artifactId = apiResource.getUUID();
            if (artifactId == null) {
                return executed;
            }
            GenericArtifact apiArtifact = artifactManager.getGenericArtifact(artifactId);
            API api = APIUtil.getAPI(apiArtifact);
            APIProvider apiProvider = APIManagerFactory.getInstance().getAPIProvider(userWithDomain);

            String oldStatus = APIUtil.getLcStateFromArtifact(apiArtifact);
            String newStatus = (targetState != null) ? targetState.toUpperCase() : targetState;

            if (newStatus != null) { //only allow the executor to be used with default LC states transition
                //check only the newStatus so this executor can be used for LC state change from
                //custom state to default api state
                if ((APIConstants.CREATED.equals(oldStatus) || APIConstants.PROTOTYPED.equals(oldStatus))
                        && APIConstants.PUBLISHED.equals(newStatus)) {
                    Set<Tier> tiers = api.getAvailableTiers();
                    String endPoint = api.getEndpointConfig();
                    if (endPoint != null && endPoint.trim().length() > 0) {
                        if (tiers == null || tiers.size() <= 0) {
                            throw new APIManagementException("Failed to publish service to API store while executing " +
                                    "APIExecutor. No Tiers selected");
                        }
                    } else {
                        throw new APIManagementException("Failed to publish service to API store while executing"
                                + " APIExecutor. No endpoint selected");
                    }
                }

                //push the state change to gateway
                Map<String, String> failedGateways = apiProvider.propergateAPIStatusChangeToGateways(api.getId(), newStatus);

                if (log.isDebugEnabled()) {
                    String logMessage = "Publish changed status to the Gateway. API Name: " + api.getId().getApiName()
                            + ", API Version " + api.getId().getVersion() + ", API Context: " + api.getContext()
                            + ", New Status : " + newStatus;
                    log.debug(logMessage);
                }

                //update api related information for state change
                executed = apiProvider.updateAPIforStateChange(api.getId(), newStatus, failedGateways);

                // Setting resource again to the context as it's updated within updateAPIStatus method
                String apiPath = APIUtil.getAPIPath(api.getId());

                apiResource = registry.get(apiPath);
                context.setResource(apiResource);

                if (log.isDebugEnabled()) {
                    String logMessage =
                            "API related information successfully updated. API Name: " + api.getId().getApiName()
                                    + ", API Version " + api.getId().getVersion() + ", API Context: " + api.getContext()
                                    + ", New Status : " + newStatus;
                    log.debug(logMessage);
                }
            } else {
                throw new APIManagementException("Invalid Lifecycle status for default APIExecutor :" + targetState);
            }


            boolean deprecateOldVersions = false;
            boolean makeKeysForwardCompatible = false;
            //If the API status is CREATED/PROTOTYPED ,check for check list items of lifecycle
            if (APIConstants.CREATED.equals(oldStatus) || APIConstants.PROTOTYPED.equals(oldStatus)) {
                deprecateOldVersions = apiArtifact.isLCItemChecked(0, APIConstants.API_LIFE_CYCLE);
                makeKeysForwardCompatible = !(apiArtifact.isLCItemChecked(1, APIConstants.API_LIFE_CYCLE));
            }

            if ((APIConstants.CREATED.equals(oldStatus) || APIConstants.PROTOTYPED.equals(oldStatus))
                    && APIConstants.PUBLISHED.equals(newStatus)) {
                if (makeKeysForwardCompatible) {
                    apiProvider.makeAPIKeysForwardCompatible(api);
                }
                if (deprecateOldVersions) {
                    String provider = APIUtil.replaceEmailDomain(api.getId().getProviderName());

                    List<API> apiList = apiProvider.getAPIsByProvider(provider);
                    APIVersionComparator versionComparator = new APIVersionComparator();
                    for (API oldAPI : apiList) {
                        if (oldAPI.getId().getApiName().equals(api.getId().getApiName()) &&
                                versionComparator.compare(oldAPI, api) < 0 &&
                                (APIConstants.PUBLISHED.equals(oldAPI.getStatus()))) {
                            apiProvider.changeLifeCycleStatus(oldAPI.getId(), APIConstants.API_LC_ACTION_DEPRECATE);

                        }
                    }
                }
            }
        } catch (RegistryException e) {
            log.error("Failed to get the generic artifact while executing APIExecutor. ", e);
            context.setProperty(LifecycleConstants.EXECUTOR_MESSAGE_KEY,
                    "APIManagementException:" + e.getMessage());
        } catch (APIManagementException e) {
            log.error("Failed to publish service to API store while executing APIExecutor. ", e);
            context.setProperty(LifecycleConstants.EXECUTOR_MESSAGE_KEY,
                    "APIManagementException:" + e.getMessage());
        } catch (FaultGatewaysException e) {
            log.error("Failed to publish service gateway while executing APIExecutor. ", e);
            context.setProperty(LifecycleConstants.EXECUTOR_MESSAGE_KEY,
                    "FaultGatewaysException:" + e.getFaultMap());
        } catch (UserStoreException e) {
            log.error("Failed to get tenant Id while executing APIExecutor. ", e);
            context.setProperty(LifecycleConstants.EXECUTOR_MESSAGE_KEY,
                    "APIManagementException:" + e.getMessage());
        }

        if (executed) {
            return publishToIstio(context);
        }
        return executed;
    }

    /**
     * Publish Istio resources
     *
     * @param context The request context that was generated from the registry core.
     *                The request context contains the resource, resource path and other
     *                variables generated during the initial call.
     * @return Returns whether the execution was successful or not.
     */
    public boolean publishToIstio(RequestContext context) {

        GenericArtifactManager artifactManager;
        GenericArtifact apiArtifact;
        try {

            artifactManager = APIUtil
                    .getArtifactManager(context.getSystemRegistry(), APIConstants.API_KEY);
            Resource apiResource = context.getResource();
            String artifactId = apiResource.getUUID();
            apiArtifact = artifactManager.getGenericArtifact(artifactId);
            API api = APIUtil.getAPI(apiArtifact);

            APIIdentifier apiIdentifier = api.getId();
            String apiName = apiIdentifier.getApiName();

            String endPoint = getAPIEndpoint(api.getEndpointConfig());

            if (!endPoint.contains(K8S_SERVICE_DOMAIN)) {
                log.info("Istio resources are not getting deployed as the endpoint is not an Istio Service!");
                return true;
            }

            String apiVersion = apiIdentifier.getVersion();
            String apiContext = api.getContext();
            String istioServiceName = getIstioServiceNameFromAPIEndPoint(endPoint);
            Set<URITemplate> uriTemplates = api.getUriTemplates();
            HashMap<String, String> resourceScopes = ApiMgtDAO.getInstance().getResourceToScopeMapping(api.getId());

            setupClient();
            setupCRDs();
            createHTTPAPISpec(apiName, apiContext, apiVersion, uriTemplates, resourceScopes);
            createHttpAPISpecBinding(apiName, istioServiceName);
            createRule(apiName, istioServiceName);

            return true;
        } catch (RegistryException e) {
            log.error("Failed to get the generic artifact while executing IstioExecutor. ", e);
        } catch (APIManagementException e) {
            log.error("Failed to publish service to API store while executing IstioExecutor. ", e);
        } catch (ParseException e) {
            log.error("Failed to parse API endpoint config while executing IstioExecutor. ", e);
        }

        return false;
    }

    /**
     * Extract Istio Service name from the endpoint
     *
     * @param endPoint Endpoint of the API
     * @return Returns Istio Service Name
     */
    private String getIstioServiceNameFromAPIEndPoint(String endPoint) throws APIManagementException {

        try {
            URL url = new URL(endPoint);
            String hostname = url.getHost();
            String[] hostnameParts = hostname.split("\\.");

            return hostnameParts[0];
        } catch (MalformedURLException e) {
            log.error("Malformed URL found for the endpoint.", e);
        }

        throw new APIManagementException("Istio Service Name could not found for the given API endpoint");
    }


    /**
     * Setup client for publishing
     */
    public void setupClient() throws APIManagementException {

        if (client == null) {
            client = new DefaultKubernetesClient(buildConfig());
        }
    }

    /**
     * Setting up custom resources
     */
    public void setupCRDs() {

        CustomResourceDefinitionList crds = client.customResourceDefinitions().list();
        List<CustomResourceDefinition> crdsItems = crds.getItems();

        for (CustomResourceDefinition crd : crdsItems) {
            ObjectMeta metadata = crd.getMetadata();
            if (metadata != null) {
                String name = metadata.getName();
                if (RULE_CRD_NAME.equals(name)) {
                    ruleCRD = crd;
                } else if (HTTPAPISpec_CRD_NAME.equals(name)) {
                    httpAPISpecCRD = crd;
                } else if (HTTPAPISpecBinding_CRD_NAME.equals(name)) {
                    httpAPISpecBindingCRD = crd;
                }
            }
        }

    }

    /**
     * Create HTTPAPISpecBinding for the API
     *
     * @param apiName  Name of the API
     * @param endPoint Endpoint of the API
     */
    public void createHttpAPISpecBinding(String apiName, String endPoint) {

        NonNamespaceOperation<HTTPAPISpecBinding, HTTPAPISpecBindingList, DoneableHTTPAPISpecBinding,
                io.fabric8.kubernetes.client.dsl.Resource<HTTPAPISpecBinding, DoneableHTTPAPISpecBinding>> bindingClient
                = client.customResource(httpAPISpecBindingCRD, HTTPAPISpecBinding.class, HTTPAPISpecBindingList.class,
                DoneableHTTPAPISpecBinding.class).inNamespace(appNamespace);

        String bindingName = Strings.toLowerCase(apiName) + "-binding";
        String apiSpecName = Strings.toLowerCase(apiName) + "-apispec";

        HTTPAPISpecBinding httpapiSpecBinding = new HTTPAPISpecBinding();
        ObjectMeta metadata = new ObjectMeta();
        metadata.setName(bindingName);
        httpapiSpecBinding.setMetadata(metadata);
        HTTPAPISpecBindingSpec httpapiSpecBindingSpec = new HTTPAPISpecBindingSpec();

        APISpec apiSpec = new APISpec();
        apiSpec.setName(apiSpecName);
        apiSpec.setNamespace(appNamespace);
        ArrayList<APISpec> apiSpecsList = new ArrayList<>();
        apiSpecsList.add(apiSpec);
        httpapiSpecBindingSpec.setApi_specs(apiSpecsList);

        Service service = new Service();
        service.setName(endPoint);
        service.setNamespace(appNamespace);
        ArrayList<Service> servicesList = new ArrayList<>();
        servicesList.add(service);
        httpapiSpecBindingSpec.setServices(servicesList);
        httpapiSpecBinding.setSpec(httpapiSpecBindingSpec);

        bindingClient.createOrReplace(httpapiSpecBinding);
        log.info("[HTTPAPISpecBinding] " + bindingName + " Created in the [Namespace] " + appNamespace + " for the"
                + " [API] " + apiName);
    }

    /**
     * Create HTTPAPISpec for the API
     *
     * @param apiName        Name of the API
     * @param apiContext     Context of the API
     * @param apiVersion     Version of the API
     * @param uriTemplates   URI templates of the API
     * @param resourceScopes Scopes of the resources of the API
     */
    public void createHTTPAPISpec(String apiName, String apiContext, String apiVersion, Set<URITemplate> uriTemplates,
                                  HashMap<String, String> resourceScopes) {

        NonNamespaceOperation<HTTPAPISpec, HTTPAPISpecList, DoneableHTTPAPISpec,
                io.fabric8.kubernetes.client.dsl.Resource<HTTPAPISpec, DoneableHTTPAPISpec>> apiSpecClient
                = client.customResource(httpAPISpecCRD, HTTPAPISpec.class, HTTPAPISpecList.class,
                DoneableHTTPAPISpec.class).inNamespace(appNamespace);

        String apiSpecName = Strings.toLowerCase(apiName) + "-apispec";

        HTTPAPISpec httpapiSpec = new HTTPAPISpec();
        ObjectMeta metadata = new ObjectMeta();
        metadata.setName(apiSpecName);
        httpapiSpec.setMetadata(metadata);

        HTTPAPISpecSpec httpapiSpecSpec = new HTTPAPISpecSpec();
        Map<String, Map<String, String>> attributeList = new HashMap<>();

        Map<String, String> apiService = new HashMap<>();
        apiService.put("stringValue", apiName);
        attributeList.put("api.service", apiService);

        Map<String, String> apiContextValue = new HashMap<>();
        apiContextValue.put("stringValue", apiContext);
        attributeList.put("api.context", apiContextValue);

        Map<String, String> apiVersionValue = new HashMap<>();
        apiVersionValue.put("stringValue", apiVersion);
        attributeList.put("api.version", apiVersionValue);

        Attributes attributes = new Attributes();
        attributes.setAttributes(attributeList);
        httpapiSpecSpec.setAttributes(attributes);
        httpapiSpecSpec.setPatterns(getPatterns(apiContext, apiVersion, uriTemplates, resourceScopes));
        httpapiSpec.setSpec(httpapiSpecSpec);

        apiSpecClient.createOrReplace(httpapiSpec);
        log.info("[HTTPAPISpec] " + apiSpecName + " Created in the [Namespace] " + appNamespace + " for the"
                + " [API] " + apiName);
    }

    /**
     * Create Mixer rule for the API
     *
     * @param apiName     Name of the API
     * @param serviceName Istio service name
     */
    private void createRule(String apiName, String serviceName) {

        NonNamespaceOperation<Rule, RuleList, DoneableRule,
                io.fabric8.kubernetes.client.dsl.Resource<Rule, DoneableRule>> ruleCRDClient
                = client.customResource(ruleCRD, Rule.class, RuleList.class,
                DoneableRule.class).inNamespace(istioSystemNamespace);
        String ruleName = Strings.toLowerCase(apiName) + "-rule";

        Rule rule = new Rule();
        ObjectMeta metadata = new ObjectMeta();
        metadata.setName(ruleName);
        rule.setMetadata(metadata);

        Action action = new Action();
        String handlerName = "wso2-handler." + istioSystemNamespace;
        action.setHandler(handlerName);
        ArrayList<String> instances = new ArrayList<>();
        instances.add("wso2-authorization");
        instances.add("wso2-metrics");
        action.setInstances(instances);

        ArrayList<Action> actions = new ArrayList<>();
        actions.add(action);

        RuleSpec ruleSpec = new RuleSpec();
        String matchValue = "context.reporter.kind == \"inbound\" && destination.namespace == \"" + appNamespace +
                "\" && destination.service.name == \"" + serviceName + "\"";
        ruleSpec.setMatch(matchValue);
        ruleSpec.setActions(actions);
        rule.setSpec(ruleSpec);

        ruleCRDClient.createOrReplace(rule);
        log.info("[Rule] " + ruleName + " Created in the [Namespace] " + istioSystemNamespace + " for the [API] "
                + apiName);

    }

    /**
     * Build the config for the client
     */
    private Config buildConfig() throws APIManagementException {

        System.setProperty(TRY_KUBE_CONFIG, "false");
        System.setProperty(TRY_SERVICE_ACCOUNT, "true");
        ConfigBuilder configBuilder;

        configBuilder = new ConfigBuilder().withMasterUrl(kubernetesAPIServerUrl);

        if (!StringUtils.isEmpty(saTokenFileName)) {
            String token;
            String tokenFile = saTokenFilePath + "/" + saTokenFileName;
            try {
                token = FileUtil.readFileToString(tokenFile);
            } catch (IOException e) {
                throw new APIManagementException("Error while reading the SA Token FIle " + tokenFile);
            }
            configBuilder.withOauthToken(token);
        }

        return configBuilder.build();
    }

    /**
     * Get API endpoint from the API endpoint config
     *
     * @param endPointConfig Endpoint config of the API
     * @return Returns API endpoint url
     */
    private String getAPIEndpoint(String endPointConfig) throws ParseException {

        if (endPointConfig != null) {
            JSONParser parser = new JSONParser();
            JSONObject endpointConfigJson = null;
            endpointConfigJson = (JSONObject) parser.parse(endPointConfig);

            if (endpointConfigJson.containsKey("production_endpoints")) {
                return getEPUrl(endpointConfigJson.get("production_endpoints"));
            }
        }

        return "";
    }

    /**
     * Verify whether the endpoint url is non empty
     *
     * @param endpoints Endpoint objects
     * @return Returns if endpoint is not empty
     */
    private static boolean isEndpointURLNonEmpty(Object endpoints) {
        if (endpoints instanceof JSONObject) {
            JSONObject endpointJson = (JSONObject) endpoints;
            if (endpointJson.containsKey("url") && endpointJson.get("url") != null) {
                String url = endpointJson.get("url").toString();
                if (org.apache.commons.lang.StringUtils.isNotBlank(url)) {
                    return true;
                }
            }
        } else if (endpoints instanceof org.json.simple.JSONArray) {
            org.json.simple.JSONArray endpointsJson = (JSONArray) endpoints;

            for (int i = 0; i < endpointsJson.size(); ++i) {
                if (isEndpointURLNonEmpty(endpointsJson.get(i))) {
                    return true;
                }
            }
        }

        return false;
    }

    /**
     * Extract Endpoint url from the endpoints object
     *
     * @param endpoints Endpoints object
     * @return Returns endpoint url
     */
    private String getEPUrl(Object endpoints) {

        if (endpoints instanceof JSONObject) {
            JSONObject endpointJson = (JSONObject) endpoints;
            if (endpointJson.containsKey("url") && endpointJson.get("url") != null) {
                String url = endpointJson.get("url").toString();
                if (org.apache.commons.lang.StringUtils.isNotBlank(url)) {
                    return url;
                }
            }
        } else if (endpoints instanceof org.json.simple.JSONArray) {
            org.json.simple.JSONArray endpointsJson = (JSONArray) endpoints;

            for (int i = 0; i < endpointsJson.size(); ++i) {
                if (isEndpointURLNonEmpty(endpointsJson.get(i))) {
                    return endpointsJson.get(i).toString();
                }
            }
        }

        return "";
    }


    /**
     * Get patterns for the given api
     *
     * @param apiContext     Context of the API
     * @param apiVersion     Version of the API
     * @param uriTemplates   URI templates of the API
     * @param resourceScopes Scopes of the resources of the API
     * @return Returns list of patterns
     */
    private ArrayList<Pattern> getPatterns(String apiContext, String apiVersion, Set<URITemplate> uriTemplates,
                                           HashMap<String, String> resourceScopes) {

        ArrayList<Pattern> patterns = new ArrayList<>();

        for (URITemplate uriTemplate : uriTemplates) {

            String resourceUri = uriTemplate.getUriTemplate();
            String httpMethod = uriTemplate.getHTTPVerb();
            Pattern pattern = new Pattern();
            pattern.setHttpMethod(httpMethod);
            pattern.setUriTemplate(resourceUri);

            Map<String, Map<String, String>> attributeList = new HashMap<>();

            Map<String, String> apiOperation = new HashMap<>();
            String apiOperationValue = "\"" + resourceUri + "\"";
            apiOperation.put("stringValue", apiOperationValue);
            attributeList.put("api.operation", apiOperation);

            String scopeKey = APIUtil.getResourceKey(apiContext, apiVersion, resourceUri, httpMethod);

            if (resourceScopes != null) {
                String scope = resourceScopes.get(scopeKey);

                if (scope != null) {
                    Map<String, String> resourceScope = new HashMap<>();
                    String resourceScopeValue = scope;
                    resourceScope.put("stringValue", resourceScopeValue);
                    attributeList.put("resource.scope", resourceScope);
                }
            }

            Attributes attributes = new Attributes();
            attributes.setAttributes(attributeList);
            pattern.setAttributes(attributes);

            patterns.add(pattern);
        }

        return patterns;
    }
}

