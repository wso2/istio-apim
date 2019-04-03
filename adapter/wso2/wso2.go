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
//go:generate $GOPATH/src/istio.io/istio/bin/mixer_codegen.sh -a mixer/adapter/wso2/config/config.proto -x "-s=false -n wso2 -t authorization"

package wso2

import (
	"context"
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"io/ioutil"
	policy "istio.io/api/policy/v1beta1"
	"istio.io/api/mixer/adapter/model/v1beta1"
	"istio.io/istio/mixer/adapter/wso2/config"
	"istio.io/istio/mixer/pkg/status"
	"istio.io/istio/mixer/template/authorization"
	"istio.io/istio/pkg/log"
	"net"
	"strings"
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
		listener net.Listener
		server   *grpc.Server
		caCert  []byte
		apimServerToken string
		apimUrl string
	}
)

var _ authorization.HandleAuthorizationServiceServer = &Wso2{}
var UnauthorizedError = errors.New("Invalid access token")

// Handle authorization
func (s *Wso2) HandleAuthorization(ctx context.Context, r *authorization.HandleAuthorizationRequest) (*v1beta1.CheckResult, error) {

	cfg := &config.Params{}

	if r.AdapterConfig != nil {
		if err := cfg.Unmarshal(r.AdapterConfig.Value); err != nil {
			log.Errorf("Error while unmarshalling adapter config: %v", err)
			return nil, err
		}
	}


	decodeValue := func(in interface{}) interface{} {
		switch t := in.(type) {
		case *policy.Value_StringValue:
			return t.StringValue
		case *policy.Value_Int64Value:
			return t.Int64Value
		case *policy.Value_DoubleValue:
			return t.DoubleValue
		default:
			return fmt.Sprintf("%v", in)
		}
	}

	decodeValueMap := func(in map[string]*policy.Value) map[string]interface{} {
		out := make(map[string]interface{}, len(in))
		for k, v := range in {
			out[k] = decodeValue(v.GetValue())
		}
		return out
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

		log.Infof("Failure.. due to missing credentials")
		return &v1beta1.CheckResult{
			Status: status.WithUnauthenticated("Missing Credentials..."),
		}, nil
	}
	headerValues := strings.Split(authHeaderValue, " ")

	if len(headerValues) < 2 {

		log.Infof("Failure.. due to invalid credentials")
		return &v1beta1.CheckResult{
			Status: status.WithUnauthenticated("Missing Credentials..."),
		}, nil
	}

	accessToken := headerValues[1]
	validateSubscription := cfg.ValidateSubscription

	tokenContent := strings.Split(accessToken, ".")
	var result bool
	var err error

	var requestAttributes = map[string]string{
		"api-name":  apiName,
		"api-version": apiVersion,
		"api-context": apiContext,
		"request-resource": requestResource,
		"request-method": requestMethod,
		"access-token": accessToken,
		"request-scope": requestScope,
	}

	if len(tokenContent) == 1 {
		serverToken := "Basic " + s.apimServerToken
		result, err = HandleOauth2AccessToken(serverToken, s.caCert, s.apimUrl, requestAttributes )
	} else {
		result, err = HandleJWT(validateSubscription, s.caCert, requestAttributes)
	}


	if err != nil {

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
	caCert, err := ioutil.ReadFile("/etc/wso2/server-cert/server.pem")
	if err != nil {
		log.Fatalf("Error in reading the Server Cert file: ", err)
		return nil, err
	}

	//reading the server cert file
	serverToken, err := ioutil.ReadFile("/etc/wso2/server-credentials/server-token")
	if err != nil {
		log.Warnf("Error in reading the server token: ", err)
	}

	//reading the server cert file
	apimUrl, err := ioutil.ReadFile("/etc/wso2/server-credentials/apim-url")
	if err != nil {
		log.Warnf("Error in reading the apim-url: ", err)
	}

	s := &Wso2{
		listener: listener,
		caCert: caCert,
		apimServerToken: string(serverToken),
		apimUrl: string(apimUrl),
	}
	log.Infof("listening on \"%v\"\n", s.Addr())
	s.server = grpc.NewServer()
	authorization.RegisterHandleAuthorizationServiceServer(s.server, s)
	return s, nil
}
