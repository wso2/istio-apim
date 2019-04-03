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
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/xml"
	"io/ioutil"
	"istio.io/istio/pkg/log"
	"net/http"
	"strconv"
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

type ValidateResponseBody struct {
	XMLName xml.Name `xml:"Envelope"`
	Text    string   `xml:",chardata"`
	Soapenv string   `xml:"soapenv,attr"`
	Body    struct {
		Text                string `xml:",chardata"`
		ValidateKeyResponse struct {
			Text   string `xml:",chardata"`
			Ns     string `xml:"ns,attr"`
			Return struct {
				Text         string `xml:",chardata"`
				Ax2129       string `xml:"ax2129,attr"`
				Ax2131       string `xml:"ax2131,attr"`
				Ax2133       string `xml:"ax2133,attr"`
				Ax2134       string `xml:"ax2134,attr"`
				Ax2137       string `xml:"ax2137,attr"`
				Xsi          string `xml:"xsi,attr"`
				AttrType     string `xml:"type,attr"`
				ApiName      string `xml:"apiName"`
				ApiPublisher string `xml:"apiPublisher"`
				ApiTier      struct {
					Text string `xml:",chardata"`
					Nil  string `xml:"nil,attr"`
				} `xml:"apiTier"`
				ApplicationId     string `xml:"applicationId"`
				ApplicationName   string `xml:"applicationName"`
				ApplicationTier   string `xml:"applicationTier"`
				Authorized        string `xml:"authorized"`
				AuthorizedDomains struct {
					Text string `xml:",chardata"`
					Nil  string `xml:"nil,attr"`
				} `xml:"authorizedDomains"`
				ConsumerKey  string `xml:"consumerKey"`
				ContentAware string `xml:"contentAware"`
				EndUserName  string `xml:"endUserName"`
				EndUserToken struct {
					Text string `xml:",chardata"`
					Nil  string `xml:"nil,attr"`
				} `xml:"endUserToken"`
				IssuedTime       string   `xml:"issuedTime"`
				Scopes           []string `xml:"scopes"`
				SpikeArrestLimit string   `xml:"spikeArrestLimit"`
				SpikeArrestUnit  struct {
					Text string `xml:",chardata"`
					Nil  string `xml:"nil,attr"`
				} `xml:"spikeArrestUnit"`
				StopOnQuotaReach       string `xml:"stopOnQuotaReach"`
				Subscriber             string `xml:"subscriber"`
				SubscriberTenantDomain string `xml:"subscriberTenantDomain"`
				ThrottlingDataList     string `xml:"throttlingDataList"`
				Tier                   string `xml:"tier"`
				Type                   string `xml:"type"`
				UserType               string `xml:"userType"`
				ValidationStatus       string `xml:"validationStatus"`
				ValidityPeriod         string `xml:"validityPeriod"`
			} `xml:"return"`
		} `xml:"validateKeyResponse"`
	} `xml:"Body"`
}

const soapServiceUrl = "/services/APIKeyValidationService.APIKeyValidationServiceHttpsSoap12Endpoint"
const version = "http://www.w3.org/2003/05/soap-envelope"
const soapDefinition = "http://org.apache.axis2/xsd"

func HandleOauth2AccessToken(serverToken string, serverCert []byte, apimUrl string, requestAttributes map[string]string) (bool, error) {

	accessToken := requestAttributes["access-token"]
	apiContext := requestAttributes["api-context"]
	apiVersion := requestAttributes["api-version"]
	resource := requestAttributes["request-resource"]
	httpMethod := requestAttributes["request-method"]

	tlsConfig := tls.Config{}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(serverCert)
	tlsConfig.RootCAs = caCertPool
	tlsConfig.BuildNameToCertificate()

	url := apimUrl + soapServiceUrl //api-manager endpoint URL
	name := &RequestData{Version: version, Var2: soapDefinition}

	name.Svs = append(name.Svs, soapBody{apiContext, apiVersion, accessToken, "Any", "?", resource, httpMethod})

	output, err := xml.MarshalIndent(name, "  ", "    ")
	if err != nil {
		log.Errorf("Error in creating the soap request: ", err)
	}

	//http client initialization
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		},
	}

	//send a new POST request
	request, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(output)))
	if err != nil {
		log.Errorf("Error in sending the POST request: ", err)
	}

	request.Header.Set("Content-Type", "application/soap+xml;charset=UTF-8")
	request.Header.Set("SOAPAction", "urn:validateKey")
	request.Header.Set("Authorization", serverToken)

	response, err := client.Do(request)
	if err != nil {
		log.Errorf("Error in response: ", err)
	}

	//Read response body
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Errorf("Error in reading response body: ", err)
	}

	//validating response body
	vrb := &ValidateResponseBody{}
	unmarshalErr := xml.Unmarshal(body, vrb)
	if unmarshalErr != nil {
		log.Errorf("Error in unmarshalling: ", unmarshalErr)
	}

	//return authorized status
	isTokenAuthorized, _ := strconv.ParseBool(vrb.Body.ValidateKeyResponse.Return.Authorized)
	var oauthError error

	if !isTokenAuthorized {
		oauthError = UnauthorizedError
	}

	return isTokenAuthorized, oauthError
}