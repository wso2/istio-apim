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

// nolint:lll
// Generates the wso2 adapter's resource yaml. It contains the adapter's configuration, name, supported template
// names (Authorization in this case), and whether it is session or no-session based.
//go:generate $GOPATH/src/istio.io/istio/bin/mixer_codegen.sh -a mixer/adapter/wso2/config/config.proto -x "-s=false -n wso2 -t authorization -t metric"

package wso2

import (
	"context"
	"errors"
	"fmt"
	"github.com/patrickmn/go-cache"
	"google.golang.org/grpc"
	"io/ioutil"
	"istio.io/api/mixer/adapter/model/v1beta1"
	policy "istio.io/api/policy/v1beta1"
	"istio.io/istio/mixer/adapter/wso2/config"
	"istio.io/istio/mixer/pkg/status"
	"istio.io/istio/mixer/template/authorization"
	"istio.io/istio/mixer/template/metric"
	"istio.io/istio/pkg/log"
	"net"
	"strconv"
	"strings"
	"time"
)

type (
	// Server is basic server interface
	Server interface {
		Addr() string
		Close() error
		Run(shutdown chan error)
	}

	// WSO2 supports authorization template.
	Wso2 struct {
		listener        net.Listener
		server          *grpc.Server
		caCert          []byte
		apimServerToken string
		apimUrl         string
	}
)

type TokenData struct {
	meta_clientType        string
	applicationConsumerKey string
	applicationName        string
	applicationId          string
	applicationOwner       string
	apiCreator             string
	apiCreatorTenantDomain string
	apiTier                string
	username               string
	userTenantDomain       string
	throttledOut           bool
	serviceTime            int64
	authorized             bool
}

type Request struct {
	stringValues  map[string]string
	longValues    map[string]int64
	intValues     map[string]int32
	booleanValues map[string]bool
}

var _ authorization.HandleAuthorizationServiceServer = &Wso2{}
var _ metric.HandleMetricServiceServer = &Wso2{}

const (
	requestStream       string = "request-stream"
	faultStream         string = "fault-stream"
	throttleStream      string = "throttle-stream"
	grpcPoolSize        string = "grpc-pool-size"
	grpcPoolInitialSize string = "grpc-pool-initial-size"
)

var UnauthorizedError = errors.New("Invalid access token")
var GlobalCache = cache.New(5*time.Minute, 10*time.Minute)

// Handle authorization
func (s *Wso2) HandleAuthorization(ctx context.Context, r *authorization.HandleAuthorizationRequest) (*v1beta1.CheckResult, error) {

	startTime := time.Now().UnixNano() / int64(time.Millisecond)

	cfg := &config.Params{}

	if r.AdapterConfig != nil {
		if err := cfg.Unmarshal(r.AdapterConfig.Value); err != nil {
			log.Errorf("Error while unmarshalling adapter config: %v", err)
			return nil, err
		}
	}

	props := decodeValueMap(r.Instance.Subject.Properties)

	authHeaderValue := props["auth_header_value"].(string)
	apiName := props["api_name"].(string)
	apiVersion := props["api_version"].(string)
	apiContext := props["api_context"].(string)
	requestResource := props["request_resource"].(string)
	requestMethod := props["request_method"].(string)
	requestScope := props["request_scope"].(string)

	if len(strings.TrimSpace(authHeaderValue)) == 0 {

		log.Errorf("Failure.. due to missing credentials")
		return &v1beta1.CheckResult{
			Status: status.WithUnauthenticated("Missing Credentials..."),
		}, nil
	}
	headerValues := strings.Split(authHeaderValue, " ")

	if len(headerValues) < 2 {

		log.Errorf("Failure.. due to invalid credentials")
		return &v1beta1.CheckResult{
			Status: status.WithUnauthenticated("Missing Credentials..."),
		}, nil
	}

	accessToken := headerValues[1]
	validateSubscription, _ := strconv.ParseBool(cfg.ValidateSubscription)
	disableHostnameVerification, _ := strconv.ParseBool(cfg.DisableHostnameVerification)

	tokenContent := strings.Split(accessToken, ".")
	var result bool
	var err error
	var tokenData TokenData

	var requestAttributes = map[string]string{
		"api-name":         apiName,
		"api-version":      apiVersion,
		"api-context":      apiContext,
		"request-resource": requestResource,
		"request-method":   requestMethod,
		"access-token":     accessToken,
		"request-scope":    requestScope,
	}

	if len(tokenContent) == 1 {
		serverToken := "Basic " + s.apimServerToken
		result, tokenData, err = HandleOauth2AccessToken(serverToken, s.caCert, s.apimUrl, requestAttributes,
			disableHostnameVerification)
	} else {
		result, tokenData, err = HandleJWT(validateSubscription, s.caCert, requestAttributes)
	}

	endTime := time.Now().UnixNano() / int64(time.Millisecond)
	serviceTime := endTime - startTime
	tokenData.serviceTime = serviceTime
	GlobalCache.Set(accessToken, &tokenData, cache.DefaultExpiration)

	if err != nil {

		log.Infof("Failure..")
		if err == UnauthorizedError {
			return &v1beta1.CheckResult{
				Status: status.WithUnauthenticated("Unauthorized!"),
			}, nil

		} else {
			return &v1beta1.CheckResult{
				Status: status.WithPermissionDenied(err.Error()),
			}, nil
		}
	}

	if result {
		log.Infof("success!!")
		return &v1beta1.CheckResult{
			Status: status.OK,
		}, nil
	}

	log.Infof("Failure..")
	return &v1beta1.CheckResult{
		Status: status.WithPermissionDenied("Unauthorized..."),
	}, nil
}

// HandleMetric records metric entries
func (s *Wso2) HandleMetric(ctx context.Context, r *metric.HandleMetricRequest) (*v1beta1.ReportResult, error) {

	cfg := &config.Params{}

	if r.AdapterConfig != nil {
		if err := cfg.Unmarshal(r.AdapterConfig.Value); err != nil {
			log.Errorf("Error unmarshalling adapter config: %v", err)
			return nil, err
		}
	}

	successRequests, faultRequests := getRequests(r.Instances)
	_, _ = HandleAnalytics(getAnalyticsConfigs(cfg), successRequests, faultRequests)

	return &v1beta1.ReportResult{}, nil
}

// get analytics configurations
func getAnalyticsConfigs(cfg *config.Params) map[string]string {

	var configs = make(map[string]string)

	configs[grpcPoolSize] = cfg.GrpcPoolSize
	configs[grpcPoolInitialSize] = cfg.GrpcPoolInitialSize
	configs[requestStream] = cfg.RequestStreamAppUrl
	configs[faultStream] = cfg.FaultStreamAppUrl
	configs[throttleStream] = cfg.ThrottleStreamAppUrl

	return configs
}

// decode the value map
func decodeValueMap(in map[string]*policy.Value) map[string]interface{} {
	out := make(map[string]interface{}, len(in))
	for k, v := range in {
		out[k] = decodeValue(v.GetValue())
	}
	return out
}

// decode values
func decodeValue(in interface{}) interface{} {
	switch t := in.(type) {
	case *policy.Value_StringValue:
		return t.StringValue
	case *policy.Value_Int64Value:
		return t.Int64Value
	case *policy.Value_DoubleValue:
		return t.DoubleValue
	case *policy.Value_IpAddressValue:
		return t.IpAddressValue
	case *policy.Value_TimestampValue:
		return t.TimestampValue
	default:
		return fmt.Sprintf("%v", in)
	}
}

// get requests for analytics
func getRequests(in []*metric.InstanceMsg) (map[int]Request, map[int]Request) {

	successRequests := make(map[int]Request)
	faultRequests := make(map[int]Request)

	for i, inst := range in {

		stringValues := make(map[string]string)
		longValues := make(map[string]int64)
		intValues := make(map[string]int32)
		booleanValues := make(map[string]bool)
		dimensions := decodeValueMap(inst.Dimensions)

		authHeaderValue := dimensions["auth_header_value"].(string)

		if len(strings.TrimSpace(authHeaderValue)) == 0 {
			continue
		}
		headerValues := strings.Split(authHeaderValue, " ")
		if len(headerValues) < 2 {
			continue
		}
		accessToken := strings.Split(authHeaderValue, " ")

		var tokenDataValues *TokenData
		var serviceTime = int64(0)

		if tokenData, found := GlobalCache.Get(accessToken[1]); found {
			tokenDataValues = tokenData.(*TokenData)

			if !tokenDataValues.authorized {
				continue
			}

			stringValues["meta_clientType"] = tokenDataValues.meta_clientType
			stringValues["applicationConsumerKey"] = tokenDataValues.applicationConsumerKey
			stringValues["applicationName"] = tokenDataValues.applicationName
			stringValues["applicationId"] = tokenDataValues.applicationId
			stringValues["applicationOwner"] = tokenDataValues.applicationOwner
			stringValues["apiCreator"] = tokenDataValues.apiCreator
			stringValues["apiCreatorTenantDomain"] = tokenDataValues.apiCreatorTenantDomain
			stringValues["apiTier"] = tokenDataValues.apiTier
			stringValues["username"] = tokenDataValues.username
			stringValues["userTenantDomain"] = tokenDataValues.userTenantDomain

			booleanValues["throttledOut"] = tokenDataValues.throttledOut
			serviceTime = tokenDataValues.serviceTime
			longValues["serviceTime"] = serviceTime
			longValues["securityLatency"] = serviceTime
		}

		apiContext := dimensions["api_context"].(string)
		apiName := dimensions["api_name"].(string)
		apiVersion := dimensions["api_version"].(string)
		apiResourcePath := dimensions["resource_path"].(string)
		apiResourceTemplate := dimensions["resource_path_template"].(string)
		apiMethod := dimensions["request_method"].(string)
		apiHostname := dimensions["request_host"].(string)
		// ipaddress
		ipAddr := dimensions["user_ip"]
		ipValue := ipAddr.(*policy.IPAddress).Value
		userIp := net.IP(ipValue).String()

		userAgent := dimensions["user_agent"].(string)

		stringValues["apiContext"] = apiContext
		stringValues["apiName"] = apiName
		stringValues["apiVersion"] = apiVersion
		stringValues["apiResourcePath"] = apiResourcePath
		stringValues["apiResourceTemplate"] = apiResourceTemplate
		stringValues["apiMethod"] = apiMethod
		stringValues["apiHostname"] = apiHostname
		stringValues["userIp"] = userIp
		stringValues["userAgent"] = userAgent

		requestTimestamp := getUnixTime("request_timestamp", dimensions)
		longValues["requestTimestamp"] = requestTimestamp

		responseTimestamp := getUnixTime("response_timestamp", dimensions)
		responseTime := responseTimestamp - requestTimestamp
		longValues["responseTime"] = responseTime

		if responseTime < serviceTime {
			serviceTime = 0
		}

		longValues["backendLatency"] = responseTime - serviceTime
		longValues["backendTime"] = responseTime - serviceTime

		responseCacheHit := false
		booleanValues["responseCacheHit"] = responseCacheHit

		responseSize := dimensions["response_size"].(int64)
		longValues["responseSize"] = responseSize

		protocol := dimensions["api_protocol"].(string)
		stringValues["protocol"] = protocol

		responseCode := dimensions["response_code"].(int64)
		intValues["responseCode"] = int32(responseCode)

		destination := dimensions["destination"].(string)
		stringValues["destination"] = destination

		throttlingLatency := 0
		longValues["throttlingLatency"] = int64(throttlingLatency)

		requestMedLat := 0
		longValues["requestMedLat"] = int64(requestMedLat)

		responseMedLat := 0
		longValues["responseMedLat"] = int64(responseMedLat)

		otherLatency := 0
		longValues["otherLatency"] = int64(otherLatency)

		gatewayType := "ISTIO"
		label := "ISTIO"
		stringValues["gatewayType"] = gatewayType
		stringValues["label"] = label

		var request Request
		request.stringValues = stringValues
		request.longValues = longValues
		request.intValues = intValues
		request.booleanValues = booleanValues

		requestType := getRequestType(responseCode)

		if requestType == requestStream {
			successRequests[i] = request
		} else if requestType == faultStream {
			stringValues["hostname"] = apiHostname
			stringValues["errorCode"] = strconv.FormatInt(responseCode, 10)
			stringValues["errorMessage"] = "Error"
			faultRequests[i] = request
		}

	}

	return successRequests, faultRequests
}

// get request type
func getRequestType(responseCode int64) string {

	var requestType string

	if responseCode >= 200 && responseCode < 300 {
		requestType = requestStream
	} else if responseCode >= 500 && responseCode < 600 {
		requestType = faultStream
	}

	return requestType
}

// get the unix time for the given property value
func getUnixTime(property string, dimensions map[string]interface{}) int64 {

	timeStamp := dimensions[property]
	timeStampValue := timeStamp.(*policy.TimeStamp).Value
	timeStampValueStr := fmt.Sprint(timeStampValue)
	timeValue, err := time.Parse(time.RFC3339, timeStampValueStr)

	if err != nil {
		log.Errorf("Error while parsing the timestamp %v", timeStampValueStr)
		return 0
	}

	return timeValue.UnixNano() / int64(time.Millisecond)
}

// Addr returns the listening address of the server
func (s *Wso2) Addr() string {
	return s.listener.Addr().String()
}

// Run starts the server run
func (s *Wso2) Run(shutdown chan error) {
	shutdown <- s.server.Serve(s.listener)
}

// Close gracefully shuts down the server; used for testing
func (s *Wso2) Close() error {
	if s.server != nil {
		s.server.GracefulStop()
	}

	if s.listener != nil {
		_ = s.listener.Close()
	}

	return nil
}

// NewWso2 creates a new IBP adapter that listens at provided port.
func NewWso2(addr string) (Server, error) {
	if addr == "" {
		addr = "0"
	}
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", addr))
	if err != nil {
		return nil, fmt.Errorf("unable to listen on socket: %v", err)
	}

	//reading the server cert file
	caCert, _ := readSecret("/etc/wso2/server-cert/server.pem")

	//reading the server cert file
	serverToken, _ := readSecret("/etc/wso2/server-credentials/server-token")

	//reading the server cert file
	apimUrl, _ := readSecret("/etc/wso2/server-credentials/apim-url")

	s := &Wso2{
		listener:        listener,
		caCert:          caCert,
		apimServerToken: string(serverToken),
		apimUrl:         string(apimUrl),
	}

	log.Infof("listening on \"%v\"\n", s.Addr())
	s.server = grpc.NewServer()
	authorization.RegisterHandleAuthorizationServiceServer(s.server, s)
	metric.RegisterHandleMetricServiceServer(s.server, s)

	return s, nil
}

//reading the secret
func readSecret(fileName string) ([]byte, error) {

	secretValue, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Warnf("Error in reading the secret %v: error - %v", fileName, err)
	}

	return secretValue, err
}
