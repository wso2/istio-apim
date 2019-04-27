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

package org.wso2.carbon.apimgt.istio.crd;

import com.fasterxml.jackson.databind.JsonDeserializer;
import com.fasterxml.jackson.databind.annotation.JsonDeserialize;
import io.fabric8.kubernetes.api.model.KubernetesResource;

import java.util.ArrayList;

@JsonDeserialize(
        using = JsonDeserializer.None.class
)
public class HTTPAPISpecBindingSpec implements KubernetesResource {
    private ArrayList<APISpec> api_specs;
    private ArrayList<Service> services;


    @Override
    public String toString() {
        return "HTTPAPISpecBindingSpec{" +
                "api_specs='" + api_specs + '\'' +
                ", services='" + services + '\'' +
                '}';
    }


    public ArrayList<APISpec> getApi_specs() {
        return api_specs;
    }

    public void setApi_specs(ArrayList<APISpec> api_specs) {
        this.api_specs = api_specs;
    }

    public ArrayList<Service> getServices() {
        return services;
    }

    public void setServices(ArrayList<Service> services) {
        this.services = services;
    }

}