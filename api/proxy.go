// This file is part of MinIO Operator
// Copyright (c) 2021 MinIO, Inc.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package api

import (
	"bytes"
	"crypto/sha1"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	url2 "net/url"
	"strings"
	"time"

	"github.com/minio/websocket"

	v2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/minio/operator/pkg/auth"
)

func serveProxy(responseWriter http.ResponseWriter, req *http.Request) {
	urlParts := strings.Split(req.URL.Path, "/")
	// Either proxy or hop, will decide the type of session
	proxyMethod := urlParts[2]

	if len(urlParts) < 5 {
		log.Println(len(urlParts))
		return
	}
	namespace := urlParts[3]
	tenantName := urlParts[4]

	// validate the tenantName

	token, err := auth.GetTokenFromRequest(req)
	if err != nil {
		log.Println(err)
		responseWriter.WriteHeader(401)
		return
	}
	claims, err := auth.SessionTokenAuthenticate(token)
	if err != nil {
		log.Printf("Unable to validate the session token %s: %v\n", token, err)
		responseWriter.WriteHeader(401)

		return
	}

	STSSessionToken := claims.STSSessionToken

	opClientClientSet, err := GetOperatorClient(STSSessionToken)
	if err != nil {
		log.Println(err)
		responseWriter.WriteHeader(404)
		return
	}
	opClient := operatorClient{
		client: opClientClientSet,
	}
	tenant, err := opClient.TenantGet(req.Context(), namespace, tenantName, metav1.GetOptions{})
	if err != nil {
		log.Println(err)
		responseWriter.WriteHeader(502)
		return
	}

	nsTenant := fmt.Sprintf("%s/%s", tenant.Namespace, tenant.Name)

	tenantSchema := "http"
	tenantPort := fmt.Sprintf(":%d", v2.ConsolePort)
	if tenant.AutoCert() || tenant.ExternalCert() {
		tenantSchema = "https"
		tenantPort = fmt.Sprintf(":%d", v2.ConsoleTLSPort)
	}

	tenantURL := fmt.Sprintf("%s://%s.%s.svc.%s%s", tenantSchema, tenant.ConsoleCIServiceName(), tenant.Namespace, v2.GetClusterDomain(), tenantPort)
	// for development
	// tenantURL = "http://localhost:9091"
	// tenantURL = "https://localhost:9443"

	h := sha1.New()
	h.Write([]byte(nsTenant))
	tenantCookieName := fmt.Sprintf("token-%s-%s-%x", proxyMethod, claims.AccountAccessKey, string(h.Sum(nil)))
	tenantCookie, err := req.Cookie(tenantCookieName)
	if err != nil {
		// login to tenantName
		loginURL := fmt.Sprintf("%s/api/v1/login", tenantURL)

		// get the tenant credentials
		clientSet, err := K8sClient(STSSessionToken)
		if err != nil {
			log.Println(err)
			responseWriter.WriteHeader(500)
			return
		}

		k8sClient := k8sClient{
			client: clientSet,
		}
		tenantConfiguration, err := GetTenantConfiguration(req.Context(), &k8sClient, tenant)
		if err != nil {
			log.Println(err)
			responseWriter.WriteHeader(500)
			return
		}

		data := map[string]interface{}{
			"accessKey": tenantConfiguration["accesskey"],
			"secretKey": tenantConfiguration["secretkey"],
		}
		// if this a proxy request hide the menu
		if proxyMethod == "proxy" {
			data["features"] = map[string]bool{
				"hide_menu": true,
			}
		}
		payload, _ := json.Marshal(data)

		loginReq, err := http.NewRequest(http.MethodPost, loginURL, bytes.NewReader(payload))
		if err != nil {
			log.Println(err)
			responseWriter.WriteHeader(500)
			return
		}
		loginReq.Header.Add("Content-Type", "application/json")

		// use localhost to ensure 'InsecureSkipVerify'
		client := GetConsoleHTTPClient("http://127.0.0.1")
		loginResp, err := client.Do(loginReq)
		if err != nil {
			log.Println(err)
			responseWriter.WriteHeader(500)
			return
		}

		if loginResp.StatusCode < 200 && loginResp.StatusCode <= 299 {
			log.Println(fmt.Printf("Status: %d. Couldn't complete login", loginResp.StatusCode))
			responseWriter.WriteHeader(500)
			return
		}

		for _, c := range loginResp.Cookies() {
			if c.Name == "token" {
				tenantCookie = c
				c := &http.Cookie{
					Name:     tenantCookieName,
					Value:    c.Value,
					Path:     c.Path,
					Expires:  c.Expires,
					HttpOnly: c.HttpOnly,
				}

				http.SetCookie(responseWriter, c)
				break
			}
		}
		defer loginResp.Body.Close()
	}

	if tenantCookie == nil {
		log.Println(errors.New("couldn't login to tenant and get cookie"))
		responseWriter.WriteHeader(403)
		return
	}
	// at this point we have a valid cookie ready to either route HTTP or WS
	// now we need to know if we are doing an  /api/ call (http) or /ws/ call (ws)
	callType := urlParts[5]

	targetURL, err := url2.Parse(tenantURL)
	if err != nil {
		log.Println(err)
		responseWriter.WriteHeader(500)
		return
	}
	tenantBase := fmt.Sprintf("/api/%s/%s/%s", proxyMethod, tenant.Namespace, tenant.Name)
	targetURL.Path = strings.ReplaceAll(req.URL.Path, tenantBase, "")

	proxiedCookie := &http.Cookie{
		Name:     "token",
		Value:    tenantCookie.Value,
		Path:     tenantCookie.Path,
		Expires:  tenantCookie.Expires,
		HttpOnly: tenantCookie.HttpOnly,
	}

	proxyCookieJar, _ := cookiejar.New(nil)
	proxyCookieJar.SetCookies(targetURL, []*http.Cookie{proxiedCookie})
	switch callType {
	case "ws":
		handleWSRequest(responseWriter, req, proxyCookieJar, targetURL, tenantSchema)
	default:
		handleHTTPRequest(responseWriter, req, proxyCookieJar, tenantBase, targetURL)
	}
}

func handleHTTPRequest(responseWriter http.ResponseWriter, req *http.Request, proxyCookieJar *cookiejar.Jar, tenantBase string, targetURL *url2.URL) {
	// use localhost to ensure 'InsecureSkipVerify'
	client := GetConsoleHTTPClient("http://127.0.0.1")
	client.Jar = proxyCookieJar
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	// are we proxying something with cp=y? (console proxy) then add cpb (console proxy base) so the console
	// on the other side updates the <base href="" /> to this value overriding sub path or root
	if v := req.URL.Query().Get("cp"); v == "y" {
		q := req.URL.Query()
		q.Add("cpb", tenantBase)
		req.URL.RawQuery = q.Encode()
	}
	// copy query params
	targetURL.RawQuery = req.URL.Query().Encode()

	proxRequest, err := http.NewRequest(req.Method, targetURL.String(), req.Body)
	if err != nil {
		log.Println(err)
		responseWriter.WriteHeader(500)
		return
	}

	for _, v := range req.Header.Values("Content-Type") {
		proxRequest.Header.Add("Content-Type", v)
	}

	resp, err := client.Do(proxRequest)
	if err != nil {
		log.Println(err)
		responseWriter.WriteHeader(500)
		return
	}

	for hk, hv := range resp.Header {
		if hk != "X-Frame-Options" {
			for _, v := range hv {
				responseWriter.Header().Add(hk, v)
			}
		}
	}
	// Allow iframes
	responseWriter.Header().Set("X-Frame-Options", "SAMEORIGIN")
	responseWriter.Header().Set("X-XSS-Protection", "1; mode=block")
	// Pass HTTP status code to proxy response
	responseWriter.WriteHeader(resp.StatusCode)

	io.Copy(responseWriter, resp.Body)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  0,
	WriteBufferSize: 1024,
}

func handleWSRequest(responseWriter http.ResponseWriter, req *http.Request, proxyCookieJar *cookiejar.Jar, targetURL *url2.URL, schema string) {
	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
		Jar:              proxyCookieJar,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}

	c, err := upgrader.Upgrade(responseWriter, req, nil)
	if err != nil {
		log.Print("error upgrade connection:", err)
		return
	}
	defer c.Close()
	if schema == "http" {
		targetURL.Scheme = "ws"
	} else {
		targetURL.Scheme = "wss"
	}

	// establish a websocket to the tenant
	tenantConn, _, err := dialer.Dial(targetURL.String(), nil)
	if err != nil {
		log.Println("dial:", err)
		return
	}
	defer tenantConn.Close()

	doneTenant := make(chan struct{})
	done := make(chan struct{})

	// start read pump from tenant connection
	go func() {
		defer close(doneTenant)
		for {
			msgType, message, err := tenantConn.ReadMessage()
			if err != nil {
				log.Println("error read from tenant:", err)
				return
			}
			c.WriteMessage(msgType, message)
		}
	}()

	// start read pump from tenant connection
	go func() {
		defer close(done)
		for {
			msgType, message, err := c.ReadMessage()
			if err != nil {
				log.Println("error read from client:", err)
				return
			}
			tenantConn.WriteMessage(msgType, message)
		}
	}()

	select {
	case <-done:
	case <-doneTenant:
	}
}
