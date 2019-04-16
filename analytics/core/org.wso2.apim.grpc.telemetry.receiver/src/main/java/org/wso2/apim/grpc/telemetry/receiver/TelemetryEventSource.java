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


import io.grpc.Server;
import io.grpc.ServerBuilder;
import org.apache.log4j.Logger;
import org.wso2.apim.grpc.telemetry.receiver.internal.ReportServiceImpl;
import org.wso2.siddhi.annotation.Example;
import org.wso2.siddhi.annotation.Extension;
import org.wso2.siddhi.annotation.Parameter;
import org.wso2.siddhi.annotation.util.DataType;
import org.wso2.siddhi.core.config.SiddhiAppContext;
import org.wso2.siddhi.core.exception.ConnectionUnavailableException;
import org.wso2.siddhi.core.stream.input.source.Source;
import org.wso2.siddhi.core.stream.input.source.SourceEventListener;
import org.wso2.siddhi.core.util.config.ConfigReader;
import org.wso2.siddhi.core.util.transport.OptionHolder;

import java.io.IOException;
import java.util.Map;

/**
 * This class implements the event source, where the received telemetry attributes can be injected to streams.
 */
@Extension(name = "telemetry-receiver", namespace = "source", description = "Telemetry Receiver for WSO2 APIM " +
        "Analytics",
        parameters = {
                @Parameter(name = "port",
                        description = "The port which the telemetry service should be started on. Default is 9091",
                        type = {DataType.INT},
                        optional = true,
                        defaultValue = "9091"),
        },
        examples = {
                @Example(syntax = "this is synatax",
                        description = "some desc")
        }
)
public class TelemetryEventSource extends Source {
    private static final Logger log = Logger.getLogger(TelemetryEventSource.class);

    private SourceEventListener sourceEventListener;
    private Server server;
    private int port;

    @Override
    public void init(SourceEventListener sourceEventListener, OptionHolder optionHolder, String[] strings,
                     ConfigReader configReader, SiddhiAppContext siddhiAppContext) {
        this.sourceEventListener = sourceEventListener;
        this.port = Integer.parseInt(optionHolder.validateAndGetStaticValue(Constants.PORT_EVENT_SOURCE_OPTION_KEY,
                Constants.DEFAULT_RECEIVER_PORT));
    }

    @Override
    public Class[] getOutputEventClasses() {
        return new Class[]{Map.class};
    }

    @Override
    public void connect(ConnectionCallback connectionCallback) throws ConnectionUnavailableException {
        try {
            this.server = ServerBuilder.forPort(this.port)
                    .addService(new ReportServiceImpl(this.sourceEventListener))
                    .build()
                    .start();
            log.info("Telemetry GRPC Server started, listening on " + port);
        } catch (IOException e) {
            throw new ConnectionUnavailableException("Unable to start the Telemetry gRPC service on port: " + port, e);
        }
    }

    @Override
    public void disconnect() {
        this.stopServer();
    }

    @Override
    public void destroy() {

    }

    @Override
    public void pause() {

    }

    @Override
    public void resume() {

    }

    @Override
    public Map<String, Object> currentState() {
        return null;
    }

    @Override
    public void restoreState(Map<String, Object> map) {

    }

    private void stopServer() {
        if (this.server != null && !this.server.isShutdown()) {
            log.info("Shutting down telemetry service");
            this.server.shutdown();
        }
    }
}
