// Package pushnotifier provides primitives for interactiving with the pushnotifier.de api.
package pushnotifier

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

type (
	// Client represents a client interfact to the haveibeenpwned.com API.
	Client struct {
		client         *http.Client
		APIToken       string
		BaseURL        *url.URL
		PackageName    string
		UserName       string
		AppToken       string
		AppTokenExpiry int64
		Devices        []string
	}

	User struct {
		UserName  string `json:"username"`
		Avatar    string `json:"avatar"`
		AppToken  string `json:"app_token"`
		ExpiresAt int64  `json:"expires_at"`
	}

	Device struct {
		ID    string   `json:"id"`
		Title string   `json:"title"`
		Model string   `json:"model"`
		Image *url.URL `json:"image"`
	}

	serverResp struct {
		Success []string `json:"success"`
		Error   []string `json:"error"`
	}
)

const (
	endpoint = "https://api.pushnotifier.de/v2/"
)

/*
var (
	accountErrors = map[int]error{
		403: errors.New("incorrect credentials"),
		404: errors.New("user not found"),
	}

	notificationSendErrors = map[int]error{
		400: errors.New("request is malformed, i.e. missing content, url, filename"),
		404: errors.New("a device could not be found"),
		413: errors.New("payload too large (> 5 MB)"),
	}
)
*/

// A NewClient creates a new PushNotifier API client. It expects 3 arguments
// 1) a `http.Client`
// 2) an API token
// 3) a package name
//
// Currently, the 1st argument will default to `http.DefaultClient` if no
// arguments are given. For more information: https://api.pushnotifier.de/v2/doc/
func NewClient(httpClient *http.Client, packageName, token, appToken string) *Client {
	if httpClient == nil {
		httpClient = &http.Client{}
	}

	baseURL, _ := url.Parse(endpoint)

	if appToken != "" {
		return &Client{
			client:         httpClient,
			APIToken:       token,
			BaseURL:        baseURL,
			PackageName:    packageName,
			UserName:       "",
			AppToken:       appToken,
			AppTokenExpiry: 0,
		}
	}
	return &Client{
		client:         httpClient,
		APIToken:       token,
		BaseURL:        baseURL,
		PackageName:    packageName,
		UserName:       "",
		AppToken:       "",
		AppTokenExpiry: 0,
	}
}

func (c *Client) request(method, resource string, formData io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, resource, formData)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	if c.AppToken != "" {
		req.Header.Set("X-AppToken", c.AppToken)
	}

	req.SetBasicAuth(c.PackageName, c.APIToken)

	req.Close = true

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}

	return resp, nil
}

// Login is used to login on behalf of a user. Logging in means to obtain a so-called "Appp Token" which is used to identify your requests.
func (c *Client) Login(username, password string) {
	resource, err := c.BaseURL.Parse("login")
	if err != nil {
		log.Fatal(err)
	}

	if username == "" || password == "" {
		log.Fatal("[Login] username and password is required to obtain App Token")
	}

	c.UserName = username
	loginData := map[string]string{"username": c.UserName, "password": password}

	formData, err := json.Marshal(loginData)
	if err != nil {
		log.Fatal("[Login] unable to create form data to send")
	}

	resp, err := c.request("POST", resource.String(), bytes.NewBuffer(formData))
	if err != nil {
		log.Fatal(err)
	}

	var user *User
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		log.Fatal(err)
	}

	c.AppToken = user.AppToken
	c.AppTokenExpiry = user.ExpiresAt
	log.Println("[Login] App Token for user obtained")
}

func (c *Client) shouldRefresh() bool {
	timeNow := time.Now().UTC().Unix()

	timeLeft := c.AppTokenExpiry - timeNow

	// if the token expiry time is lower than 1_000 seconds, it is time to refresh the token.
	return timeLeft < 1_000
}

// RefreshToken is used to refresh your obtain App Token.
func (c *Client) RefreshToken() {
	resource, err := c.BaseURL.Parse("user/refresh")
	if err != nil {
		log.Fatal(err)
	}

	resp, err := c.request("GET", resource.String(), nil)
	if err != nil {
		log.Fatal(err)
	}

	var user *User
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		log.Fatal(err)
	}

	c.AppToken = user.AppToken
	c.AppTokenExpiry = user.ExpiresAt
	log.Println("[RefreshToken] App Token for user obtained")
}

// GetDevices get all devices a user has registered and that are available for sending.
func (c *Client) GetDevices() {
	resource, err := c.BaseURL.Parse("devices")
	if err != nil {
		log.Fatal(err)
	}

	if c.shouldRefresh() {
		c.RefreshToken()
	}

	resp, err := c.request("GET", resource.String(), nil)
	if err != nil {
		log.Fatal(err)
	}

	var devices *[]Device
	err = json.NewDecoder(resp.Body).Decode(&devices)
	if err != nil {
		log.Fatal(err)
	}

	// Append obtained devices' IDs to the Client struct for future use.
	for _, device := range *devices {
		c.Devices = append(c.Devices, device.ID)
	}

	log.Println("[GetDevices] Registered devices for user obtained")
	log.Printf("%#v", devices)
}

// SendText sends a notification to all registered clients with a simple text.
func (c *Client) SendText(content string, silent bool) {
	resource, err := c.BaseURL.Parse("notitification/text")
	if err != nil {
		log.Fatal(err)
	}

	if c.shouldRefresh() {
		c.RefreshToken()
	}

	if content == "" {
		log.Fatalln("[SendText] Requires content to send as notification")
	}

	sendData := struct {
		Devices []string `json:"devices"`
		Content string   `json:"content"`
		Silent  bool     `json:"silent"`
	}{
		Devices: c.Devices,
		Content: content,
		Silent:  silent,
	}

	formData, err := json.Marshal(sendData)
	if err != nil {
		log.Fatal("[SendText] unable to create form data to send")
	}

	resp, err := c.request("PUT", resource.String(), bytes.NewBuffer(formData))
	if err != nil {
		log.Fatal(err)
	}

	var sResp serverResp
	err = json.NewDecoder(resp.Body).Decode(&sResp)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("[SendText]", sResp.Success)
}

// SendURL sends a notification to all registered clients with a URL.
func (c *Client) SendURL(contentURL string, silent bool) {
	resource, err := c.BaseURL.Parse("notitification/text")
	if err != nil {
		log.Fatal(err)
	}

	if c.shouldRefresh() {
		c.RefreshToken()
	}

	if contentURL == "" {
		log.Fatalln("[SendURL] Requires content to send as notification")
	}

	parsedContentURL, err := url.Parse(contentURL)
	if err != nil {
		log.Fatal(err)
	}

	sendData := struct {
		Devices []string `json:"devices"`
		URL     *url.URL `json:"url"`
		Silent  bool     `json:"silent"`
	}{
		Devices: c.Devices,
		URL:     parsedContentURL,
		Silent:  silent,
	}

	formData, err := json.Marshal(sendData)
	if err != nil {
		log.Fatal("[SendURL] unable to create form data to send")
	}

	resp, err := c.request("PUT", resource.String(), bytes.NewBuffer(formData))
	if err != nil {
		log.Fatal(err)
	}

	var sResp serverResp
	err = json.NewDecoder(resp.Body).Decode(&sResp)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("[SendText]", sResp.Success)
}

// SendNotification sends a notification to all registered clients with content or URL.
func (c *Client) SendNotification(content, contentURL string, silent bool) {
	resource, err := c.BaseURL.Parse("notitification/text")
	if err != nil {
		log.Fatal(err)
	}

	if c.shouldRefresh() {
		c.RefreshToken()
	}

	if content == "" || contentURL == "" {
		log.Fatalln("[SendNotification] Requires content or url to send as notification")
	}

	parsedContentURL, err := url.Parse(contentURL)
	if err != nil {
		log.Fatal(err)
	}

	sendData := struct {
		Devices []string `json:"devices"`
		Content string   `json:"content"`
		URL     *url.URL `json:"url"`
		Silent  bool     `json:"silent"`
	}{
		Devices: c.Devices,
		Content: content,
		URL:     parsedContentURL,
		Silent:  silent,
	}

	formData, err := json.Marshal(sendData)
	if err != nil {
		log.Fatal("[SendNotification] unable to create form data to send")
	}

	resp, err := c.request("PUT", resource.String(), bytes.NewBuffer(formData))
	if err != nil {
		log.Fatal(err)
	}

	var sResp serverResp
	err = json.NewDecoder(resp.Body).Decode(&sResp)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("[SendNotification]", sResp.Success)
}

// SendImage sends a notification to all registered clients with an Image.
func (c *Client) SendImage(contentFile string, silent bool) {
	resource, err := c.BaseURL.Parse("notitification/text")
	if err != nil {
		log.Fatal(err)
	}

	if c.shouldRefresh() {
		c.RefreshToken()
	}

	if contentFile == "" {
		log.Fatalln("[SendNotification] Requires content or url to send as notification")
	}

	osStat, err := os.Stat(contentFile)
	if err != nil {
		log.Fatal(err)
	}

	// check if file path exists or not
	if os.IsNotExist(err) {
		log.Fatal(err)
	}

	// check if given path is to a file
	if osStat.IsDir() {
		log.Fatal(errors.New("given path is a directory and not a file"))
	}

	// check if file size is greater than 5_000_000 bytes or 5 Megabytes (MB)
	if osStat.Size() > 5_000_000 {
		log.Fatal(errors.New("file size is greater than 5 MegaBytes (MB)"))
	}

	fileRaw, err := ioutil.ReadFile(contentFile)
	if err != nil {
		log.Fatal(err)
	}
	encodedContent := base64.StdEncoding.EncodeToString(fileRaw)

	sendData := struct {
		Devices  []string `json:"devices"`
		Content  string   `json:"content"`
		Filename string   `json:"filename"`
		Silent   bool     `json:"silent"`
	}{
		Devices:  c.Devices,
		Content:  encodedContent,
		Filename: osStat.Name(),
		Silent:   silent,
	}

	formData, err := json.Marshal(sendData)
	if err != nil {
		log.Fatal("[SendImage] unable to create form data to send")
	}

	resp, err := c.request("PUT", resource.String(), bytes.NewBuffer(formData))
	if err != nil {
		log.Fatal(err)
	}

	var sResp serverResp
	err = json.NewDecoder(resp.Body).Decode(&sResp)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("[SendImage]", sResp.Success)
}
