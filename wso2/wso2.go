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
// names (metric in this case), and whether it is session or no-session based.
//go:generate $GOPATH/src/istio.io/istio/bin/mixer_codegen.sh -a mixer/adapter/wso2/config/config.proto -x "-s=false -n wso2 -t authorization"


package wso2

import (
	"context"
	"net"
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	logf "log"
	"net/http"
	"os"
	"google.golang.org/grpc"
	"istio.io/api/mixer/adapter/model/v1beta1"
	policy "istio.io/api/policy/v1beta1"
	"istio.io/istio/mixer/adapter/wso2/config"
	"istio.io/istio/mixer/pkg/status"
	"istio.io/istio/mixer/template/authorization"
	"istio.io/istio/pkg/log"
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
	}
)

type RequestData struct {
	XMLName xml.Name   `xml:"soap:Envelope"`
	Version string     `xml:"xmlns:soap,attr"`
	Var2    string     `xml:"xmlns:xsd,attr"`
	Head    string     `xml:"soap:Header"`
	Svs     []soapBody `xml:"soap:Body"`
}

type soapBody struct {
	Context                     string `xml:"xsd:validateKey>xsd:context"`
	ApiVersion                  string `xml:"xsd:validateKey>xsd:version"`
	AccessToken                 string `xml:"xsd:validateKey>xsd:accessToken"`
	RequiredAuthenticationLevel string `xml:"xsd:validateKey>xsd:requiredAuthenticationLevel"`
	ClientDomain                string `xml:"xsd:validateKey>xsd:clientDomain"`
	MatchingResource            string `xml:"xsd:validateKey>xsd:matchingResource"`
	HttpVerb                    string `xml:"xsd:validateKey>xsd:httpVerb"`
}

var _ authorization.HandleAuthorizationServiceServer = &Wso2{}

// HandleMetric records metric entries
func (s *Wso2) HandleAuthorization(ctx context.Context, r *authorization.HandleAuthorizationRequest) (*v1beta1.CheckResult, error) {

	log.Infof("received request - test123445 %v\n", *r)

	cfg := &config.Params{}

	if r.AdapterConfig != nil {
		if err := cfg.Unmarshal(r.AdapterConfig.Value); err != nil {
			log.Errorf("error unmarshalling adapter config: %v", err)
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

	// Calling km APIM start
	tlsConfig := tls.Config{}

	// Load client cert
	cert, err := tls.LoadX509KeyPair("/etc/client-cert/client.cer.pem", "/etc/client-key/client.key.pem")
	if err != nil {
		logf.Fatal("error load client cert :", err)
	}
	tlsConfig.Certificates = []tls.Certificate{cert}



	//Load CA cert
	caCert, err := ioutil.ReadFile("/etc/server-cert/server.cer.pem")
	if err != nil {
		logf.Fatalf("error file read")
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	tlsConfig.RootCAs = caCertPool

	tlsConfig.BuildNameToCertificate()

	if err != nil {
		logf.Fatal(err)
	}

	url := "https://test.wso2.com:9443/services/APIKeyValidationService.APIKeyValidationServiceHttpsSoap12Endpoint";
	name := &RequestData{Version: "http://www.w3.org/2003/05/soap-envelope", Var2: "http://org.apache.axis2/xsd"}
	name.Svs = append(name.Svs, soapBody{"/mock/v1", "v1", "801cc400-0f67-3e41-98cb-8ba19df1b8d1", "Any", "?", "/test", "GET"})

	output, err := xml.MarshalIndent(name, "  ", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	os.Stdout.Write(output)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		},
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(output)))

	if err != nil {
		fmt.Println(err)
	}

	// req.Header.Set("Accept-Encoding","gzip,deflate")
	req.Header.Set("Content-Type", "application/soap+xml;charset=UTF-8")
	req.Header.Set("SOAPAction", "urn:validateKey")
	// req.Header.Set("Content-Length","697")
	// req.Header.Set("Connection","Keep-Alive")
	// req.Header.Set("User-Agent","Apache-HttpClient/4.1.1 (java 1.5)")
	req.Header.Set("Authorization", "Basic YWRtaW46YWRtaW4=")
	// req.Header.Set("keyManager", "verifyHostname")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	fmt.Printf("%s\n",string(body));
	if err != nil {
		panic(err.Error())
	}
	// Calling km APIM end


	props := decodeValueMap(r.Instance.Subject.Properties)
	log.Infof("%v", props)

	log.Infof("cfg.ApimUrl")
	log.Infof(cfg.ApimUrl)

	for k, v := range props {
		fmt.Println("k:", k, "v:", v)
		if (k == "custom_token_header") {
			log.Infof("success!! and done!")
			return &v1beta1.CheckResult{
				Status: status.OK,
			}, nil
		}

	}

	log.Infof("failure; header not provided")
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
	s := &Wso2{
		listener: listener,
	}
	fmt.Printf("listening on \"%v\"\n", s.Addr())
	s.server = grpc.NewServer()
	authorization.RegisterHandleAuthorizationServiceServer(s.server, s)
	return s, nil
}
