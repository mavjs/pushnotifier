// Package pushnotifier provides primitives for interactiving with the pushnotifier.de api.
package pushnotifier

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
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
		ID    string `json:"id"`
		Title string `json:"title"`
		Model string `json:"model"`
		Image string `json:"image"`
	}

	serverRespSuccess struct {
		Success interface{} `json:"success"`
		Error   []string    `json:"error"`
	}
)

const (
	endpoint = "https://api.pushnotifier.de/v2/"
)

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
			AppTokenExpiry: -1,
			Devices:        make([]string, 0),
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
		Devices:        make([]string, 0),
	}
}

func (c *Client) request(method, resource string, formData io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, resource, formData)
	if err != nil {
		return nil, err
	}

	if c.shouldRefresh() {
		c.RefreshToken()
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	if c.AppToken != "" {
		req.Header.Set("X-AppToken", c.AppToken)
	}

	req.SetBasicAuth(c.PackageName, c.APIToken)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		respBody, _ := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()

		return nil, fmt.Errorf("%v - %v", resp.Status, string(respBody))

	}

	return resp, nil
}

// Login is used to login on behalf of a user. Logging in means to obtain a so-called "Appp Token" which is used to identify your requests.
func (c *Client) Login(username, password string) error {
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
		return fmt.Errorf("[Login] unable to create form data to send: %v", err.Error())
	}

	resp, err := c.request("POST", resource.String(), bytes.NewBuffer(formData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var user *User
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		return fmt.Errorf("[Login] unable to decode response body as JSON: %v", err.Error())
	}

	c.AppToken = user.AppToken
	c.AppTokenExpiry = user.ExpiresAt
	log.Println("[Login] App Token for user obtained")

	return nil
}

func (c *Client) shouldRefresh() bool {
	// if we provided an appToken/APP_TOKEN to NewClient, that means we do not need to refresh token.
	if c.AppTokenExpiry == -1 {
		return false
	}

	timeNow := time.Now().UTC().Unix()

	timeLeft := c.AppTokenExpiry - timeNow

	// if the token expiry time is lower than 1_000 seconds, it is time to refresh the token.
	return timeLeft < 1_000
}

// RefreshToken is used to refresh your obtain App Token.
func (c *Client) RefreshToken() error {
	resource, err := c.BaseURL.Parse("user/refresh")
	if err != nil {
		return err
	}

	resp, err := c.request("GET", resource.String(), nil)
	if err != nil {
		return err
	}

	var user *User
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		return fmt.Errorf("[RefreshToken] unable to decode response body as JSON: %v", err.Error())
	}

	c.AppToken = user.AppToken
	c.AppTokenExpiry = user.ExpiresAt
	log.Println("[RefreshToken] App Token for user obtained")

	return nil
}

// GetDevices get all devices a user has registered and that are available for sending.
func (c *Client) GetDevices() error {
	resource, err := c.BaseURL.Parse("devices")
	if err != nil {
		return err
	}

	if c.shouldRefresh() {
		c.RefreshToken()
	}

	resp, err := c.request("GET", resource.String(), nil)
	if err != nil {
		return err
	}

	var devices *[]Device
	err = json.NewDecoder(resp.Body).Decode(&devices)
	if err != nil {
		return fmt.Errorf("[GetDevices] unable to decode response body as JSON: %v", err.Error())
	}

	log.Println("[GetDevices] Registered devices for user obtained")

	// Append obtained devices' IDs to the Client struct for future use and print to user with full details.
	// TODO: Append full metadata of devices, however, during sending notification to all devices just have a internal function that creates a slice of IDs.
	for _, device := range *devices {
		c.Devices = append(c.Devices, device.ID)
	}

	return nil
}

// SendText sends a notification to all registered clients with a simple text.
func (c *Client) SendText(content string, devices []string, silent bool) error {
	resource, err := c.BaseURL.Parse("notifications/text")
	if err != nil {
		return err
	}

	if content == "" {
		return errors.New("[SendText] content to send as notification was empty")
	}

	if len(c.Devices) == 0 && len(devices) == 0 {
		log.Println("[SendText] No devices given. Acquring devices...")
		c.GetDevices()
		devices = append(devices, c.Devices...)
		log.Println(devices)
	}

	sendData := struct {
		Devices []string `json:"devices"`
		Content string   `json:"content"`
		Silent  bool     `json:"silent"`
	}{
		Devices: devices,
		Content: content,
		Silent:  silent,
	}

	formData, err := json.Marshal(sendData)
	if err != nil {
		return errors.New("[SendText] unable to create form data to send")
	}

	resp, err := c.request("PUT", resource.String(), bytes.NewBuffer(formData))
	if err != nil {
		return err
	}

	var sResp serverRespSuccess
	err = json.NewDecoder(resp.Body).Decode(&sResp)
	if err != nil {
		return fmt.Errorf("[SendText] unable to decode response body as JSON: %v", err.Error())
	}

	log.Println("[SendText]", sResp.Success)

	return nil
}

// SendURL sends a notification to all registered clients with a URL.
func (c *Client) SendURL(contentURL string, devices []string, silent bool) error {
	resource, err := c.BaseURL.Parse("notifications/url")
	if err != nil {
		return err
	}

	if contentURL == "" {
		return errors.New("[SendURL] content URL to send as notification was empty")
	}

	parsedContentURL, err := url.Parse(contentURL)
	if err != nil {
		return err
	}

	if len(c.Devices) == 0 && devices == nil {
		log.Println("[SendURL] No devices given. Acquring devices...")
		c.GetDevices()
		copy(devices, c.Devices)
	}

	sendData := struct {
		Devices []string `json:"devices"`
		URL     string   `json:"url"`
		Silent  bool     `json:"silent"`
	}{
		Devices: devices,
		URL:     parsedContentURL.String(),
		Silent:  silent,
	}

	formData, err := json.Marshal(sendData)
	if err != nil {
		return errors.New("[SendURL] unable to create form data to send")
	}

	resp, err := c.request("PUT", resource.String(), bytes.NewBuffer(formData))
	if err != nil {
		return err
	}

	var sResp serverRespSuccess
	err = json.NewDecoder(resp.Body).Decode(&sResp)
	if err != nil {
		return fmt.Errorf("[SendURL] unable to decode response body as JSON: %v", err.Error())
	}

	log.Println("[SendText]", sResp.Success)

	return nil
}

// SendNotification sends a notification to all registered clients with content or URL.
func (c *Client) SendNotification(content, contentURL string, devices []string, silent bool) error {
	resource, err := c.BaseURL.Parse("notifications/notification")
	if err != nil {
		return err
	}

	if c.shouldRefresh() {
		c.RefreshToken()
	}

	if content == "" || contentURL == "" {
		return errors.New("[SendNotification] content text or URL to send as notification was empty")
	}

	parsedContentURL, err := url.Parse(contentURL)
	if err != nil {
		return err
	}

	if len(c.Devices) == 0 && devices == nil {
		log.Println("[SendNotification] No devices given. Acquring devices...")
		c.GetDevices()
		copy(devices, c.Devices)
	}

	sendData := struct {
		Devices []string `json:"devices"`
		Content string   `json:"content"`
		URL     string   `json:"url"`
		Silent  bool     `json:"silent"`
	}{
		Devices: devices,
		Content: content,
		URL:     parsedContentURL.String(),
		Silent:  silent,
	}

	formData, err := json.Marshal(sendData)
	if err != nil {
		return errors.New("[SendNotification] unable to create form data to send")
	}

	resp, err := c.request("PUT", resource.String(), bytes.NewBuffer(formData))
	if err != nil {
		return err
	}

	var sResp serverRespSuccess
	err = json.NewDecoder(resp.Body).Decode(&sResp)
	if err != nil {
		return fmt.Errorf("[SendNotification] unable to decode response body as JSON: %v", err.Error())
	}

	log.Println("[SendNotification]", sResp.Success)

	return nil
}

// SendImage sends a notification to all registered clients with an Image.
func (c *Client) SendImage(contentFile string, devices []string, silent bool) error {
	resource, err := c.BaseURL.Parse("notifications/image")
	if err != nil {
		return err
	}

	if c.shouldRefresh() {
		c.RefreshToken()
	}

	if contentFile == "" {
		return errors.New("[SendImage] content image file path to send as notification was empty")
	}

	osStat, err := os.Stat(contentFile)
	if err != nil {
		return err
	}

	// check if file path exists or not
	if os.IsNotExist(err) {
		return err
	}

	// check if given path is to a file
	if osStat.IsDir() {
		return errors.New("[SendImage] given image file path is a directory and not an actual file")
	}

	// check if file size is greater than 5_000_000 bytes or 5 Megabytes (MB)
	if osStat.Size() > 5_000_000 {
		return errors.New("[SendImage] given file size is greater than 5 MegaBytes (MB)")
	}

	fileRaw, err := ioutil.ReadFile(contentFile)
	if err != nil {
		return err
	}
	encodedContent := base64.StdEncoding.EncodeToString(fileRaw)

	if len(c.Devices) == 0 && devices == nil {
		log.Println("[SendImage] No devices given. Acquring devices...")
		c.GetDevices()
		copy(devices, c.Devices)
	}

	sendData := struct {
		Devices  []string `json:"devices"`
		Content  string   `json:"content"`
		Filename string   `json:"filename"`
		Silent   bool     `json:"silent"`
	}{
		Devices:  devices,
		Content:  encodedContent,
		Filename: osStat.Name(),
		Silent:   silent,
	}

	formData, err := json.Marshal(sendData)
	if err != nil {
		return errors.New("[SendImage] unable to create form data to send")
	}

	resp, err := c.request("PUT", resource.String(), bytes.NewBuffer(formData))
	if err != nil {
		return err
	}

	var sResp serverRespSuccess
	err = json.NewDecoder(resp.Body).Decode(&sResp)
	if err != nil {
		return fmt.Errorf("[SendImage] unable to decode response body as JSON: %v", err.Error())
	}

	log.Println("[SendImage]", sResp.Success)

	return nil
}
