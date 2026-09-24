package http

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"math/big"
	"sync"
	"testing"

	"github.com/golang-jwt/jwt/v4"
	xcommon "github.com/rdkcentral/xconfadmin/common"
	"github.com/stretchr/testify/require"
)

type authTestIdpService struct {
	config *IdpServiceConfig
	keys   *JsonWebKeyResponse
	url    string
}

func (s *authTestIdpService) IdpServiceHost() string        { return "https://trusted-idp.example" }
func (s *authTestIdpService) SetIdpServiceHost(string)      {}
func (s *authTestIdpService) GetFullLoginUrl(string) string { return "" }
func (s *authTestIdpService) GetJsonWebKeyResponse(url string) *JsonWebKeyResponse {
	s.url = url
	return s.keys
}
func (s *authTestIdpService) GetFullLogoutUrl(string) string         { return "" }
func (s *authTestIdpService) GetToken(string) string                 { return "" }
func (s *authTestIdpService) Logout(string) error                    { return nil }
func (s *authTestIdpService) GetIdpServiceConfig() *IdpServiceConfig { return s.config }

func TestValidateAndGetLoginTokenUsesConfiguredJWKSAndTrustedClaims(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	require.NoError(t, err)
	kid := "trusted-kid"
	service := &authTestIdpService{
		config: &IdpServiceConfig{
			JWKSURL:     "https://trusted-idp.example/keys/",
			Issuer:      "https://trusted-idp.example",
			Audience:    "xconfadmin",
			AllowedAlgs: []string{"RS256", "HS256"},
		},
		keys: &JsonWebKeyResponse{Keys: []JsonWebKey{{
			KeyType: "RSA",
			Kid:     kid,
			N:       base64.RawURLEncoding.EncodeToString(privateKey.PublicKey.N.Bytes()),
			E:       base64.RawURLEncoding.EncodeToString(big.NewInt(int64(privateKey.PublicKey.E)).Bytes()),
		}}},
	}
	service.config.KidMap = sync.Map{}
	previousServer := WebConfServer
	previousProvider := xcommon.AuthProvider
	defer func() {
		WebConfServer = previousServer
		xcommon.AuthProvider = previousProvider
	}()
	WebConfServer = &WebconfigServer{IdpServiceConnector: service}
	xcommon.AuthProvider = "idp"

	claims := jwt.MapClaims{
		"iss": "https://trusted-idp.example",
		"aud": "xconfadmin",
		"sub": "admin@example.com",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = kid
	token.Header["jku"] = "http://attacker.example/keys"
	signedToken, err := token.SignedString(privateKey)
	require.NoError(t, err)

	loginToken, err := ValidateAndGetLoginToken(signedToken)
	require.NoError(t, err)
	require.Equal(t, "admin@example.com", loginToken.Subject)
	require.Equal(t, service.config.JWKSURL, service.url)
}

func TestValidateAndGetLoginTokenRejectsUntrustedClaimsAndMethods(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	require.NoError(t, err)
	service := &authTestIdpService{config: &IdpServiceConfig{
		JWKSURL:     "https://trusted-idp.example/keys/",
		Issuer:      "trusted-issuer",
		Audience:    "trusted-audience",
		AllowedAlgs: []string{"RS256", "HS256"},
	}, keys: &JsonWebKeyResponse{Keys: []JsonWebKey{{
		KeyType: "RSA",
		Kid:     "trusted-kid",
		N:       base64.RawURLEncoding.EncodeToString(privateKey.PublicKey.N.Bytes()),
		E:       base64.RawURLEncoding.EncodeToString(big.NewInt(int64(privateKey.PublicKey.E)).Bytes()),
	}}}}
	service.config.KidMap = sync.Map{}
	previousServer := WebConfServer
	previousProvider := xcommon.AuthProvider
	defer func() {
		WebConfServer = previousServer
		xcommon.AuthProvider = previousProvider
	}()
	WebConfServer = &WebconfigServer{IdpServiceConnector: service}
	xcommon.AuthProvider = "idp"

	for _, testCase := range []struct {
		name   string
		claims jwt.MapClaims
		method jwt.SigningMethod
	}{
		{name: "issuer", claims: jwt.MapClaims{"iss": "wrong", "aud": "trusted-audience"}, method: jwt.SigningMethodRS256},
		{name: "audience", claims: jwt.MapClaims{"iss": "trusted-issuer", "aud": "wrong"}, method: jwt.SigningMethodRS256},
		{name: "method", claims: jwt.MapClaims{"iss": "trusted-issuer", "aud": "trusted-audience"}, method: jwt.SigningMethodHS256},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			token := jwt.NewWithClaims(testCase.method, testCase.claims)
			token.Header["kid"] = "trusted-kid"
			var signedToken string
			if testCase.method == jwt.SigningMethodHS256 {
				signedToken, err = token.SignedString([]byte("secret"))
			} else {
				signedToken, err = token.SignedString(privateKey)
			}
			require.NoError(t, err)
			_, err = ValidateAndGetLoginToken(signedToken)
			require.Error(t, err)
		})
	}
}
