package wso2

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"io/ioutil"
	"log"
	"strings"
	"time"
)

type Payload struct {
	Aud         string `json:"aud"`
	Sub         string `json:"sub"`
	Application struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Tier  string `json:"tier"`
		Owner string `json:"owner"`
	} `json:"application"`
	Scope          string `json:"scope"`
	Iss            string `json:"iss"`
	Keytype        string `json:"keytype"`
	SubscribedAPIs []struct {
		Name                   string `json:"name"`
		Context                string `json:"context"`
		Version                string `json:"version"`
		Publisher              string `json:"publisher"`
		SubscriptionTier       string `json:"subscriptionTier"`
		SubscriberTenantDomain string `json:"subscriberTenantDomain"`
	} `json:"subscribedAPIs"`
	ConsumerKey string `json:"consumerKey"`
	Exp         int    `json:"exp"`
	Iat         int64  `json:"iat"`
	Jti         string `json:"jti"`
}

var rsaData = []struct {
	name        string
	tokenString string
	alg         string
	claims      map[string]interface{}
	valid       bool
}{
	{
		"Basic RS256",
		"eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsIng1dCI6Ik5UQXhabU14TkRNeVpEZzNNVFUxWkdNME16RXpPREpoWldJNE5ETmxaRFUxT0dGa05qRmlNUSJ9.eyJhdWQiOiJodHRwOlwvXC9vcmcud3NvMi5hcGltZ3RcL2dhdGV3YXkiLCJzdWIiOiJhZG1pbiIsImFwcGxpY2F0aW9uIjp7ImlkIjoxLCJuYW1lIjoiRGVmYXVsdEFwcGxpY2F0aW9uIiwidGllciI6IlVubGltaXRlZCIsIm93bmVyIjoiYWRtaW4ifSwic2NvcGUiOiJhbV9hcHBsaWNhdGlvbl9zY29wZSBkZWZhdWx0IiwiaXNzIjoiaHR0cHM6XC9cL2xvY2FsaG9zdDo5NDQzXC9vYXV0aDJcL3Rva2VuIiwia2V5dHlwZSI6IlBST0RVQ1RJT04iLCJzdWJzY3JpYmVkQVBJcyI6W3sibmFtZSI6IlBpenphU2hhY2tBUEkiLCJjb250ZXh0IjoiXC9waXp6YXNoYWNrXC8xLjAuMCIsInZlcnNpb24iOiIxLjAuMCIsInB1Ymxpc2hlciI6ImFkbWluIiwic3Vic2NyaXB0aW9uVGllciI6IlVubGltaXRlZCIsInN1YnNjcmliZXJUZW5hbnREb21haW4iOiJjYXJib24uc3VwZXIifV0sImNvbnN1bWVyS2V5IjoiMlQ1ZEJqVFc0Vm1jMTE3eWZmQkViaEd5cUJ3YSIsImV4cCI6MTU0OTE2OTk4NywiaWF0IjoxNTQ5MTY5OTc3NTA5LCJqdGkiOiIwMTdjNjAxYi04NTQxLTQ4YzMtOTJlZC0wYTNmNmRjMjc1MTQifQ==.iOajg3An4V5ghEIU9Y7AZFBC4YYdOxIiwxq6V4lE1lt-uvZjN0XlXpwHegkNkfjDnQk_J1W82S2-RHo2K8fav9MN3xqVH89GWmJCyN2BkZKO2ScyoJs-KGg2SgbrPc4hNixmmEOZqHtdPWlDI5C77Sy8YGT8CF7E0fRRvIVjaJ4fR6IewfXGB7ucmoHs9RCObspoOq7JJd044h-mwRDwoJoUk6v_rn3djDIxzb0A9e6Y1T3GB6vpshs_sl7xJ0lW_d22eC5Wy7GbxbJGFrfdYHxk03NRpk53QAxtQKksiPZAN2LGfHSErryWOCjdTkkNBH6P3l-X17QzaU_fs7vC1A==",
		"RS256",
		map[string]interface{}{"foo": "bar"},
		true,
	},
}

func ValidateToken(publicCertFile string) {

	PublicKey, err := ioutil.ReadFile(publicCertFile)
	if err != nil {
		log.Println("Error in reading the cert file: ", err)
	}

	key, err := jwt.ParseRSAPublicKeyFromPEM(PublicKey)
	if err != nil {
		log.Println("Error in parse:", err)
	}

	data := rsaData[0]
	parts := strings.Split(data.tokenString, ".")
	err = jwt.SigningMethodRS256.Verify(strings.Join(parts[0:2], "."), parts[2], key)
	if err != nil {
		log.Println("Error in verifying/Incorrect token: ", data.name, err)
	}

	ts := string(parts[1])
	sDec, _ := b64.StdEncoding.DecodeString(ts)

	pld := &Payload{}
	er := json.Unmarshal(sDec, pld)
	if er != nil {
		log.Println("Error in unmarshalling payload: ", er)
	}

	nowTime := time.Now().Unix()

	issueTime := nowTime - (pld.Iat / 1000)
	expireTime := nowTime - int64(pld.Exp)
	
	if 0 > expireTime {
		fmt.Println(issueTime - expireTime)
		log.Println("Token is valid....")
	} else {
		fmt.Println(issueTime - expireTime)
		fmt.Println("Tocken is expired.....")
	}
}
