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
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"istio.io/istio/pkg/log"
	"strings"
	"time"
	"strconv"
)

type JWTData struct {
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

var UnauthorizedError = errors.New("Invalid access token")

func HandleJWT(validateSubscription string, apiName string, apiVersion string, publicCert []byte, accessToken string) (bool , error) {


	tokenContent := strings.Split(accessToken, ".")

	if len(tokenContent) != 3 {
		log.Errorf("Invalid JWT token received, token must have 3 parts")
		return false, UnauthorizedError
	}

	signedContent := tokenContent[0] + "." + tokenContent[1]
	err := validateSignature(publicCert, signedContent, tokenContent[2] )
	if err != nil {
		log.Errorf("Error in validating the signature: %v", err)
		return false, UnauthorizedError
	}

	jwtData, err := decodePayload(string(tokenContent[1]))
	if jwtData == nil {
		log.Errorf("Error in decoding the payload: %v", err)
		return false, UnauthorizedError
	}

	if isTokenExpired(jwtData) {
		return false, UnauthorizedError
	}

	validateSub, _ := strconv.ParseBool(validateSubscription)
	if validateSub {
		if isSubscriptionValid(jwtData, apiName, apiVersion) {
			return true, nil
		}
		return false, errors.New("Resource forbidden")
	}

	return true, nil
}

// validate the signature
func validateSignature(publicCert []byte, signedContent string, signature string) error {

	key, err := jwt.ParseRSAPublicKeyFromPEM(publicCert)

	if err != nil {
		log.Errorf("Error in parsing the public key: %v", err)
		return err
	}

	return jwt.SigningMethodRS256.Verify(signedContent, signature, key)
}

// decode the payload
func decodePayload(payload string) (*JWTData, error) {

	data, _ := base64.StdEncoding.DecodeString(payload)

	jwtData := JWTData{}
	err := json.Unmarshal(data, &jwtData)
	if err != nil {
		log.Errorf("Error in unmarshalling payload: %v", err)
		return nil, err
	}

	return &jwtData, nil
}

// check whether the token has expired
func isTokenExpired(jwtData *JWTData) bool {

	nowTime := time.Now().Unix()
	expireTime := int64(jwtData.Exp)

	if expireTime < nowTime {
		log.Infof("Token is expired!")
		return true
	}

	return false
}

// do the subscription validation
func isSubscriptionValid(jwtData *JWTData, apiName string, apiVersion string) bool {

	for _, api := range jwtData.SubscribedAPIs {

		if (strings.ToLower(apiName) == strings.ToLower(api.Name)) && apiVersion == api.Version {
			return true
		}
	}

	log.Infof("Subscription is not valid!")
	return false
}
