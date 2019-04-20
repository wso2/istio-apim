// Copyright (c)  WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
//
// WSO2 Inc. licenses this file to you under the Apache License,
// Version 2.0 (the "License"); you may not use this file   except
// in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package wso2

import (
	"google.golang.org/grpc"
	"istio.io/istio/pkg/log"
	report "org.wso2.apim.grpc.telemetry.receiver.generated"
	"golang.org/x/net/context"
	"github.com/processout/grpc-go-pool"
	"strconv"
	"time"
)

// handle metrics requests
func HandleAnalytics(configs map[string]string, successRequests map[int]Request, faultRequests map[int]Request) (bool, error) {

	publishEvents(configs, successRequests, requestStream)
	publishEvents(configs, faultRequests, faultStream)

	return true, nil
}

// publish events via gRPC
func publishEvents(configs map[string]string, requests map[int]Request, requestType string) () {

	if len(requests) == 0 {
		return
	}

	grpcPoolSize, _ := strconv.Atoi(configs[grpcPoolSize])
	grpcPoolInitialSize, _ := strconv.Atoi(configs[grpcPoolInitialSize])
	pool := createGrpcPool(configs[requestType], grpcPoolInitialSize, grpcPoolSize)

	conn, _ := pool.Get(context.Background())
	defer conn.Close()

	c := report.NewReportServiceClient(conn.ClientConn)

	for _, request := range requests {

		_, err := c.Report(context.Background(), &report.ReportRequest{StringValues: request.stringValues,
			LongValues: request.longValues, IntValues: request.intValues, BooleanValues: request.booleanValues})
		if err != nil {
			log.Fatalf("Error when calling SayHello: %s", err)
		}

	}

	return
}

// create gRPC connection pool for the given target
func createGrpcPool(targetUrl string, grpcPoolIntialSize int, grpcPoolSize int) (*grpcpool.Pool) {

	factory1 := func() (*grpc.ClientConn, error) {
		conn, err := grpc.Dial(targetUrl, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("Failed to start gRPC connection %v: %v", targetUrl, err)
		}
		log.Infof("Connected to gRPC server at %v", targetUrl)
		return conn, err
	}

	pool, err := grpcpool.New(factory1, grpcPoolIntialSize, grpcPoolSize, time.Second)
	if err != nil {
		log.Fatalf("Failed to create gRPC pool for target %v: %v", targetUrl, err)
	}

	return pool
}
