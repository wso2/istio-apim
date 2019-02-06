package wso2

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
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

func KeyValidationHandler(serverToken string, accessToken string, serverCert []byte, apimUrl string, path string, httpVerb string) (result string) {

	tlsConfig := tls.Config{}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(serverCert)
	tlsConfig.RootCAs = caCertPool
	tlsConfig.BuildNameToCertificate()

	url := apimUrl + soapServiceUrl //apim endpoint URL
	name := &RequestData{Version: version, Var2: soapDefinition}
	name.Svs = append(name.Svs, soapBody{"/pizzashack/1.0.0", "1.0.0", accessToken, "Any", "?", "/menu", httpVerb})

	output, err := xml.MarshalIndent(name, "  ", "    ")
	if err != nil {
		log.Println("Error in creating the soap request: ", err)
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
		log.Println("Error in sending the POST request: ", err)
	}

	request.Header.Set("Content-Type", "application/soap+xml;charset=UTF-8")
	request.Header.Set("SOAPAction", "urn:validateKey")
	request.Header.Set("Authorization", serverToken)

	response, err := client.Do(request)
	if err != nil {
		log.Println("Error in response: ", err)
	}

	//Read response body
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println("Error in reading response body: ", err)
	}

	//validating response body
	vrb := &ValidateResponseBody{}
	unmarshalErr := xml.Unmarshal(body, vrb)
	if unmarshalErr != nil {
		log.Println("Error in unmarshalling: ", unmarshalErr)
	}

	//return authorized status
	oAuthResult := vrb.Body.ValidateKeyResponse.Return.Authorized
	return oAuthResult
}
