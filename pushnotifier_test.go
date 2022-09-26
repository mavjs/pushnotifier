/*
Copyright © 2022 Maverick Kaung <mavjs01@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package pushnotifier

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	mockHandler *http.ServeMux
	mockServer  *httptest.Server
)

func init() {
	mockHandler = http.NewServeMux()
	mockServer = httptest.NewServer(mockHandler)
}

func TestNewClient(t *testing.T) {
	assert := assert.New(t)

	packageName := "dev.myapp.pn"
	apiToken := "aabbccdd112233"
	var appToken = ""

	pn := NewClient(nil, packageName, apiToken, appToken)

	assert.Equal(apiToken, pn.APIToken, "[TestNewClient] Expected provided and client returned API Token to be equal")
}

func TestNewClientWithMockServer(t *testing.T) {
	assert := assert.New(t)

	want, _ := url.Parse(mockServer.URL)
	packageName := "dev.myapp.pn"
	apiToken := "aabbccdd112233"
	var appToken = ""

	pn := NewClient(nil, packageName, apiToken, appToken)
	pn.BaseURL, _ = url.Parse(mockServer.URL)

	assert.Equal(want, pn.BaseURL, "[TestNewClientWithMockServer] Expected provided and client returned base URL to be equal")
}

func TestLogin(t *testing.T) {
	assert := assert.New(t)

	packageName := "dev.myapp.pn"
	apiToken := "aabbccdd112233"
	wantAppToken := "ZZXX11ff"

	test_user := "aUser"
	test_password := "aUserPassword"

	mockHandler.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		reqUsername, reqPassword, _ := r.BasicAuth()

		if reqUsername != packageName {
			http.Error(w, "package name is invalid", http.StatusUnauthorized)
			return
		}
		if reqPassword != apiToken {
			http.Error(w, "api token is invalid", http.StatusUnauthorized)
			return
		}

		reqBody, err := io.ReadAll(r.Body)
		if err != nil {
			t.Log(err)
		}
		var user struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		json.Unmarshal(reqBody, &user)

		givenUsername := user.Username
		log.Printf("[TestLogin] givenUsername: %#v", givenUsername)
		givenPassword := user.Password
		log.Printf("[TestLogin] givenPassword: %#v", givenPassword)

		if givenUsername != test_user {
			http.Error(w, "user not be found", http.StatusNotFound)
			return
		}
		if givenPassword != test_password {
			http.Error(w, "invalid credentials", http.StatusForbidden)
			return
		}

		timeNowSec := time.Now().UTC()
		expiryTime := timeNowSec.AddDate(0, 0, 30).Unix()

		fmt.Fprint(w, `{
			"username": "aUser",
			"avatar": "https://example.com/avatar/00000000000000000000000000000000",
			"app_token": "`+wantAppToken+`",
			"expires_at": `+strconv.FormatInt(expiryTime, 10)+`}`)
	})

	pn := NewClient(nil, packageName, apiToken, "")
	pn.BaseURL, _ = url.Parse(mockServer.URL)

	if err := pn.Login("aUser", "aUserPassword"); err != nil {
		t.Error(err)
	}

	assert.Equal(wantAppToken, pn.AppToken, "[TestLogin] Expected wanted and recevied APP Token to be equal")
}
