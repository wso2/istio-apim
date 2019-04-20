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
	"strconv"
	"strings"
	"time"
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

type Subscription struct {
	name                   string
	context                string
	version                string
	publisher              string
	subscriptionTier       string
	subscriberTenantDomain string
}

var Unknown = "__unknown__"

func HandleJWT(validateSubscription bool, publicCert []byte, requestAttributes map[string]string) (bool, TokenData, error) {

	accessToken := requestAttributes["access-token"]
	apiName := requestAttributes["api-name"]
	apiVersion := requestAttributes["api-version"]
	requestScope := requestAttributes["request-scope"]

	tokenContent := strings.Split(accessToken, ".")
	var tokenData TokenData

	if len(tokenContent) != 3 {
		log.Errorf("Invalid JWT token received, token must have 3 parts")
		return false, tokenData, UnauthorizedError
	}

	signedContent := tokenContent[0] + "." + tokenContent[1]
	err := validateSignature(publicCert, signedContent, tokenContent[2])
	if err != nil {
		log.Errorf("Error in validating the signature: %v", err)
		return false, tokenData, UnauthorizedError
	}

	jwtData, err := decodePayload(string(tokenContent[1]))
	if jwtData == nil {
		log.Errorf("Error in decoding the payload: %v", err)
		return false, tokenData, UnauthorizedError
	}

	if isTokenExpired(jwtData) {
		return false, tokenData, UnauthorizedError
	}

	if !isRequestScopeValid(jwtData, requestScope) {
		return false, tokenData, UnauthorizedError
	}

	if validateSubscription {

		subscription := getSubscription(jwtData, apiName, apiVersion)

		if &subscription == nil {
			return false, tokenData, errors.New("Resource forbidden")
		}

		return true, getTokenDataForJWT(jwtData, apiName, apiVersion), nil
	}

	return true, getTokenDataForJWT(jwtData, apiName, apiVersion), nil
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

// do resource scope validation
func isRequestScopeValid(jwtData *JWTData, requestScope string) bool {

	if len(requestScope) > 0 {

		tokenScopes := strings.Split(jwtData.Scope, " ")

		for _, tokenScope := range tokenScopes {
			if requestScope == tokenScope {
				log.Infof("Matching scopes found!")
				return true
			}

		}
		log.Infof("No matching scopes found!")
		return false
	}

	log.Infof("No scopes defined")
	return true
}

// get the subscription
func getSubscription(jwtData *JWTData, apiName string, apiVersion string) Subscription {

	var subscription Subscription
	for _, api := range jwtData.SubscribedAPIs {

		if (strings.ToLower(apiName) == strings.ToLower(api.Name)) && apiVersion == api.Version {
			subscription.name = apiName
			subscription.version = apiVersion
			subscription.context = api.Context
			subscription.publisher = api.Publisher
			subscription.subscriptionTier = api.SubscriptionTier
			subscription.subscriberTenantDomain = api.SubscriberTenantDomain
			return subscription
		}
	}

	log.Infof("Subscription is not valid!")
	return subscription
}

func getTokenDataForJWT(jwtData *JWTData, apiName string, apiVersion string) TokenData {

	var tokenData TokenData

	tokenData.authorized = true
	tokenData.meta_clientType = jwtData.Keytype
	tokenData.applicationConsumerKey = jwtData.ConsumerKey
	tokenData.applicationName = jwtData.Application.Name
	tokenData.applicationId = strconv.Itoa(jwtData.Application.ID)
	tokenData.applicationOwner = jwtData.Application.Owner

	subscription := getSubscription(jwtData, apiName, apiVersion)

	if &subscription == nil {
		tokenData.apiCreator = Unknown
		tokenData.apiCreatorTenantDomain = Unknown
		tokenData.apiTier = Unknown
		tokenData.userTenantDomain = Unknown
	} else {
		tokenData.apiCreator = subscription.publisher
		tokenData.apiCreatorTenantDomain = subscription.subscriberTenantDomain
		tokenData.apiTier = subscription.subscriptionTier
		tokenData.userTenantDomain = subscription.subscriberTenantDomain
	}

	tokenData.username = jwtData.Sub
	tokenData.throttledOut = false

	return tokenData
}
