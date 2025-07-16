// Copyright 2025 Comcast Cable Communications Management, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
package http

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/rdkcentral/xconfadmin/common"
	"github.com/rdkcentral/xconfadmin/util"

	"github.com/go-akka/configuration"
	log "github.com/sirupsen/logrus"
)

var idpServiceName string

const (
	defaultClientId     = "clientId"
	defaultClientSecret = "clientSecret"
	getTokenUrl         = "%s/oauth/token?code=%s"
	fullLoginUrl        = "%s/protected/auth?continue=%s&client_id=%s&response_type=code"
	fullLogoutUrl       = "%s/logout?continue=%s&client_id=%s"
	odpServiceName      = "odp"
)

type IdpServiceConfig struct {
	ClientId        string
	ClientSecret    string
	KidMap          sync.Map // map[string]JsonWebKey
	AuthHeaderValue string
}

type JsonWebKeyResponse struct {
	Keys []JsonWebKey `json:"keys"`
}

type JsonWebKey struct {
	KeyType string `json:"kty"`
	E       string `json:"e"`
	Use     string `json:"use"`
	Kid     string `json:"kid"`
	Alg     string `json:"alg"`
	N       string `json:"n"`
}

type IdpServiceConnector interface {
	IdpServiceHost() string
	SetIdpServiceHost(host string)
	GetFullLoginUrl(continueUrl string) string
	GetJsonWebKeyResponse(url string) *JsonWebKeyResponse
	GetFullLogoutUrl(continueUrl string) string
	GetToken(code string) string
	Logout(url string) error
	GetIdpServiceConfig() *IdpServiceConfig
}
type DefaultIdpService struct {
	host string
	*HttpClient
	*IdpServiceConfig
}

func NewIdpServiceConnector(conf *configuration.Config, externalIdpService IdpServiceConnector) IdpServiceConnector {
	if externalIdpService != nil {
		return externalIdpService
	} else {
		idpServiceName = conf.GetString("xconfwebconfig.xconf.idp_service_name")
		confKey := fmt.Sprintf("xconfwebconfig.%v.host", idpServiceName)
		host := conf.GetString(confKey)
		if util.IsBlank(host) {
			panic(fmt.Errorf("%s is required", confKey))
		}
		clientId := os.Getenv("IDP_CLIENT_ID")
		if util.IsBlank(clientId) {
			confKey := fmt.Sprintf("xconfwebconfig.%v.client_id", idpServiceName)
			clientId = conf.GetString(confKey)
			if util.IsBlank(clientId) {
				panic("No env IDP_CLIENT_ID")
			}
		}
		clientSecret := os.Getenv("IDP_CLIENT_SECRET")
		if util.IsBlank(clientSecret) {
			confKey := fmt.Sprintf("xconfwebconfig.%v.client_secret", idpServiceName)
			clientSecret = conf.GetString(confKey)
			if util.IsBlank(clientSecret) {
				panic("No env IDP_CLIENT_SECRET")
			}
		}
		auth := fmt.Sprintf("%s:%s", clientId, clientSecret)
		authHeader := fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(auth)))

		idpServiceConfig := &IdpServiceConfig{
			ClientId:        clientId,
			ClientSecret:    clientSecret,
			KidMap:          sync.Map{},
			AuthHeaderValue: authHeader,
		}

		return &DefaultIdpService{
			host:             host,
			HttpClient:       NewHttpClient(conf, odpServiceName, nil),
			IdpServiceConfig: idpServiceConfig,
		}
	}
}

func (xc *DefaultIdpService) IdpServiceHost() string {
	return xc.host
}

func (xc *DefaultIdpService) GetIdpServiceConfig() *IdpServiceConfig {
	return xc.IdpServiceConfig
}

func (xc *DefaultIdpService) SetIdpServiceHost(host string) {
	xc.host = host
}

func (xc *DefaultIdpService) GetFullLoginUrl(continueUrl string) string {
	return fmt.Sprintf(fullLoginUrl, xc.host, continueUrl, xc.ClientId)
}

func (xc *DefaultIdpService) GetFullLogoutUrl(continueUrl string) string {
	return fmt.Sprintf(fullLogoutUrl, xc.host, continueUrl, xc.ClientId)
}

func (xc *DefaultIdpService) GetToken(code string) string {
	url := fmt.Sprintf(getTokenUrl, xc.IdpServiceHost(), code)
	header := map[string]string{
		common.HeaderAuthorization: xc.AuthHeaderValue,
	}
	rrbytes, err := xc.DoWithRetries("POST", url, header, nil, nil, idpServiceName)
	// make async?
	if err != nil {
		log.Errorf("error getting token from IdpService: %s", err.Error())
		return ""
	}
	return string(rrbytes)
}

func (xc *DefaultIdpService) GetJsonWebKeyResponse(url string) *JsonWebKeyResponse {
	rrbytes, err := xc.DoWithRetries("GET", url, nil, nil, nil, idpServiceName)
	if err != nil {
		return nil
	}
	var jsonWebKeyResponse JsonWebKeyResponse
	err = json.Unmarshal(rrbytes, &jsonWebKeyResponse)
	if err != nil {
		return nil
	}
	return &jsonWebKeyResponse
}

func (xc *DefaultIdpService) Logout(url string) error {
	_, err := xc.DoWithRetries("GET", url, nil, nil, nil, idpServiceName)
	return err
}
