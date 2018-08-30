package main

import (
	"crypto"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/docker/libtrust"
	"github.com/golang/glog"
)

// TokenValidity is the validity time of the token in seconds
const TokenValidity = 900

// Header describes the header section of a JSON Web Token.
type Header struct {
	Type       string `json:"typ"`
	SigningAlg string `json:"alg"`
	KeyID      string `json:"kid,omitempty"`
}

// ClaimSet describes the main section of a JSON Web Token.
type ClaimSet struct {
	// Public claims
	Issuer     string `json:"iss"`
	Subject    string `json:"sub"`
	Audience   string `json:"aud"`
	Expiration int64  `json:"exp"`
	NotBefore  int64  `json:"nbf"`
	IssuedAt   int64  `json:"iat"`
	JWTID      string `json:"jti"`

	// Private claims
	Access []Access `json:"access,omitempty"`
	Scopes string   `json:"scopes,omitempty"`
}

// GenerateJTI generates JTI for token
func GenerateJTI() string {
	randomBytes := make([]byte, 15)
	if _, err := rand.Read(randomBytes); err != nil {
		glog.Errorf("unable to read random bytes for jwt id: %s", err)
		return ""
	}
	return base64.URLEncoding.EncodeToString(randomBytes)
}

// GenerateToken generate a JWS token for the specified accesses
func GenerateToken(accesses Accesses, audience string, subject string, iat time.Time, jti string) (string, error) {
	// get the private key
	privkey, err := libtrust.LoadKeyFile(AuthConfig.JWSKey)
	if err != nil {
		glog.Errorf("Could not load key file: %s", err)
		return "", err
	}
	// craft the headers
	joseHeader := &Header{
		Type:       "JWT",
		SigningAlg: "RS256",
		KeyID:      privkey.KeyID(),
	}
	claimSet := &ClaimSet{
		Issuer:     AuthConfig.Issuer,
		Subject:    subject,
		Audience:   audience,
		Expiration: iat.Unix() + TokenValidity,
		NotBefore:  iat.Unix(),
		IssuedAt:   iat.Unix(),
		JWTID:      jti,
		Access:     accesses,
	}
	// get bytes of the parts
	var joseHeaderBytes, claimSetBytes []byte
	if joseHeaderBytes, err = json.Marshal(joseHeader); err != nil {
		glog.Errorf("unable to marshal jose header: %s", err)
		return "", fmt.Errorf("unable to marshal jose header: %s", err)
	}
	if claimSetBytes, err = json.Marshal(claimSet); err != nil {
		glog.Errorf("unable to marshal claim set: %s", err)
		return "", fmt.Errorf("unable to marshal claim set: %s", err)
	}
	// generate jwt pratical payload
	encodedJoseHeader := joseBase64UrlEncode(joseHeaderBytes)
	encodedClaimSet := joseBase64UrlEncode(claimSetBytes)
	encodingToSign := fmt.Sprintf("%s.%s", encodedJoseHeader, encodedClaimSet)
	// generate signature
	var signatureBytes []byte
	if signatureBytes, _, err = privkey.Sign(strings.NewReader(encodingToSign), crypto.SHA256); err != nil {
		glog.Errorf("unable to sign jwt payload: %s", err)
		return "", fmt.Errorf("unable to sign jwt payload: %s", err)
	}
	signature := joseBase64UrlEncode(signatureBytes)
	return fmt.Sprintf("%s.%s", encodingToSign, signature), nil
}

func joseBase64UrlEncode(b []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(b), "=")
}
