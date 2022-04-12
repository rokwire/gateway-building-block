// Copyright 2021 Board of Trustees of the University of Illinois.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package authutils

import (
	"bytes"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

// ContainsString returns true if the provided value is in the provided slice
func ContainsString(slice []string, val string) bool {
	for _, v := range slice {
		if val == v {
			return true
		}
	}
	return false
}

// RemoveString removes the provided value from the provided slice
// 	Returns modified slice. If val is not found returns unmodified slice
func RemoveString(slice []string, val string) ([]string, bool) {
	for i, other := range slice {
		if other == val {
			return append(slice[:i], slice[i+1:]...), true
		}
	}
	return slice, false
}

// GetKeyFingerprint returns the fingerprint for a given rsa.PublicKey
func GetKeyFingerprint(key *rsa.PublicKey) (string, error) {
	if key == nil {
		return "", errors.New("key cannot be nil")
	}
	pubPkcs1 := x509.MarshalPKCS1PublicKey(key)

	hash, err := HashSha256(pubPkcs1)
	if err != nil {
		return "", fmt.Errorf("error hashing key: %v", err)
	}

	return "SHA256:" + base64.StdEncoding.EncodeToString(hash), nil
}

// GetPubKeyPem returns the PEM encoded public key
func GetPubKeyPem(key *rsa.PublicKey) (string, error) {
	if key == nil {
		return "", errors.New("key cannot be nil")
	}

	pubASN1, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return "", fmt.Errorf("error marshalling public key: %v", err)
	}

	pemdata := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: pubASN1,
		},
	)

	return string(pemdata), nil
}

// HashSha256 returns the SHA256 hash of the input
func HashSha256(data []byte) ([]byte, error) {
	if data == nil {
		return nil, fmt.Errorf("cannot hash nil data")
	}

	hasher := sha256.New()
	_, err := hasher.Write(data)
	if err != nil {
		return nil, fmt.Errorf("error writing data: %v", err)
	}
	return hasher.Sum(nil), nil
}

// ResetRequestBody sets r.Body to read from data (use to read from r.Body multiple times)
func ResetRequestBody(r *http.Request, data []byte) {
	r.Body = ioutil.NopCloser(bytes.NewReader(data))
}
