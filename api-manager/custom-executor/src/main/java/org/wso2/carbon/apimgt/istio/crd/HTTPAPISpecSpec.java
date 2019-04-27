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
import java.util.Map;

@JsonDeserialize(
        using = JsonDeserializer.None.class
)
public class HTTPAPISpecSpec implements KubernetesResource {

    private ArrayList<Map<String, String>> apiKeys;
    private Attributes attributes;
    private ArrayList<Pattern> patterns;

    public void setAttributes(Attributes attributes) {
        this.attributes = attributes;
    }

    public Attributes getAttributes() {
        return attributes;
    }

    public ArrayList<Pattern> getPatterns() {
        return patterns;
    }

    public void setPatterns(ArrayList<Pattern> patterns) {
        this.patterns = patterns;
    }

    @Override
    public String toString() {
        return "HTTPAPISpecSpec{" +
                "apiKeys='" + apiKeys + '\'' +
                ", attributes='" + attributes + '\'' +
                ", patterns='" + patterns + '\'' +
                '}';
    }

    public ArrayList<Map<String, String>> getApiKeys() {
        return apiKeys;
    }

    public void setApiKeys(ArrayList<Map<String, String>> apiKeys) {
        this.apiKeys = apiKeys;
    }
}