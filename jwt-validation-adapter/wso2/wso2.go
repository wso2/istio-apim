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

func ValidateToken(publicCertFile string, accessToken string) {

	PublicKey, err := ioutil.ReadFile(publicCertFile)
	if err != nil {
		log.Println("Error in reading the cert file: ", err)
	}

	key, err := jwt.ParseRSAPublicKeyFromPEM(PublicKey)
	if err != nil {
		log.Println("Error in parse:", err)
	}

	parts := strings.Split(accessToken, ".")
	err = jwt.SigningMethodRS256.Verify(strings.Join(parts[0:2], "."), parts[2], key)
	if err != nil {
		log.Println("Error in verifying/Incorrect token: ", err)
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
