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
package org.wso2.apim.grpc.telemetry.receiver;

import com.google.gson.Gson;
import com.google.protobuf.ByteString;
import com.google.protobuf.Duration;
import com.google.protobuf.Timestamp;

import java.util.HashMap;
import java.util.Map;
import java.util.logging.Logger;

/**
 * This class holds the list of decoded attributes that was received by the Telemetry service.
 */
public class AttributesBag {

    private static final Logger logger = Logger.getLogger(AttributesBag.class.getName());
    private Map<String, Object> attribute = new HashMap<>();
    private Map<String, String> requestHeaders = new HashMap<>();
    private Map<String, String> responseHeaders = new HashMap<>();

    public void put(String key, String value) {
        putAttribute(key, value);
    }

    public void put(String key, Map<String, String> value) {
        if (key.equalsIgnoreCase(Constants.REQUEST_HEADER_FIELDS_ATTRIBUTE)) {
            requestHeaders = value;
        } else if (key.equalsIgnoreCase(Constants.RESPONSE_HEADER_FIELDS_ATTRIBUTE)) {
            responseHeaders = value;
        }
        putAttribute(key, Utils.toString(value));
    }

    public void put(String key, Long value) {
        putAttribute(key, value);
    }

    public void put(String key, int value) {
        putAttribute(key, value);
    }

    public void put(String key, Boolean value) {
        putAttribute(key, value);
    }

    public void put(String key, Double value) {
        putAttribute(key, value);
    }

    public void put(String key, ByteString value) {
        putAttribute(key, Utils.toString(value));
    }

    public void put(String key, Timestamp value) {
        putAttribute(key + Constants.SECONDS_KEY, value.getSeconds());
        putAttribute(key + Constants.NANO_SECONDS_KEY, value.getNanos());
    }

    public void put(String key, Duration value) {
        putAttribute(key + Constants.SECONDS_KEY, value.getSeconds());
        putAttribute(key + Constants.NANO_SECONDS_KEY, value.getNanos());
    }

    private void putAttribute(String key, Object value) {
        Object existingValue = this.attribute.putIfAbsent(key, value);
        if (existingValue != null) {
            logger.warning("Attribute key - " + key + " already present with value - " + existingValue
                    + " , therefore cannot replace with new value - " + value);
        }
    }

    public Map<String, Object> getAttributes() {
        return this.attribute;
    }

    public Map<String, String> getRequestHeaders() {
        return requestHeaders;
    }

    public Map<String, String> getResponseHeaders() {
        return responseHeaders;
    }

    public String toString() {
        Gson gson = new Gson();
        return gson.toJson(this.attribute);
    }

}
