/*
Copyright 2014 Google Inc. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package client

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
)

type basicAuthRoundTripper struct {
	username string
	password string
	rt       http.RoundTripper
}

func NewBasicAuthRoundTripper(username, password string, rt http.RoundTripper) http.RoundTripper {
	return &basicAuthRoundTripper{username, password, rt}
}

func (rt *basicAuthRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req = cloneRequest(req)
	req.SetBasicAuth(rt.username, rt.password)
	return rt.rt.RoundTrip(req)
}

type bearerAuthRoundTripper struct {
	bearer string
	rt     http.RoundTripper
}

func NewBearerAuthRoundTripper(bearer string, rt http.RoundTripper) http.RoundTripper {
	return &bearerAuthRoundTripper{bearer, rt}
}

func (rt *bearerAuthRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req = cloneRequest(req)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", rt.bearer))
	return rt.rt.RoundTrip(req)
}

func NewClientCertTLSTransport(certData, keyData, caData []byte) (*http.Transport, error) {
	cert, err := tls.X509KeyPair(certData, keyData)
	if err != nil {
		return nil, err
	}
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(caData)
	return &http.Transport{
		TLSClientConfig: &tls.Config{
			// Change default from SSLv3 to TLSv1.0 (because of POODLE vulnerability)
			MinVersion: tls.VersionTLS10,
			Certificates: []tls.Certificate{
				cert,
			},
			RootCAs:    certPool,
			ClientCAs:  certPool,
			ClientAuth: tls.RequireAndVerifyClientCert,
		},
	}, nil
}

func NewTLSTransport(caData []byte) (*http.Transport, error) {
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(caData)
	return &http.Transport{
		TLSClientConfig: &tls.Config{
			// Change default from SSLv3 to TLSv1.0 (because of POODLE vulnerability)
			MinVersion: tls.VersionTLS10,
			RootCAs:    certPool,
		},
	}, nil
}

func NewUnsafeTLSTransport() *http.Transport {
	return &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
}

// cloneRequest returns a clone of the provided *http.Request.
// The clone is a shallow copy of the struct and its Header map.
func cloneRequest(r *http.Request) *http.Request {
	// shallow copy of the struct
	r2 := new(http.Request)
	*r2 = *r
	// deep copy of the Header
	r2.Header = make(http.Header)
	for k, s := range r.Header {
		r2.Header[k] = s
	}
	return r2
}
