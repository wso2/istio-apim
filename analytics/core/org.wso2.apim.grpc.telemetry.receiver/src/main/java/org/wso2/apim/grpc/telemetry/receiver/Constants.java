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

/**
 * This class hold the Constants for the Telemetry Receiver component.
 */
public class Constants {

    static final String DEFAULT_RECEIVER_PORT = "9091";
    static final String PORT_EVENT_SOURCE_OPTION_KEY = "port";
    static final String SECONDS_KEY = "_sec";
    static final String NANO_SECONDS_KEY = "_nanosec";
    static final String REQUEST_HEADER_FIELDS_ATTRIBUTE = "request.headers";
    static final String RESPONSE_HEADER_FIELDS_ATTRIBUTE = "response.headers";
    public static final String UNKNOWN_ATTRIBUTE = "Unknown";

}
