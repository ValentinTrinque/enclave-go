package api_client

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Enclave-Markets/enclave-go/models"
)

type ApiKeyArgs struct {
	KeyId     string
	KeySecret string
	Timestamp string
	Sign      string
}

type ApiClient struct {
	ApiEndpoint string

	// Can be used to authenticate requests. Either with JWT token or an API key. The API key needs to sign
	// each request with a timestamp and signature.
	apiKeyArgs *ApiKeyArgs
	Headers    map[string]string
}

func (c *ApiClient) WithApiKey(keyId, keySecret string) {
	c.apiKeyArgs = &ApiKeyArgs{
		KeyId:     keyId,
		KeySecret: keySecret,
	}
}

func generateSignature(apiSecret string, timestamp string, method string, requestPath string, body string) []byte {
	concattedString := timestamp + method + requestPath + body
	mac := hmac.New(sha256.New, []byte(apiSecret))
	mac.Write([]byte(concattedString))
	return mac.Sum(nil)
}

func (c *ApiClient) computeApiKeyArgs(httpVerb string, path string, request any) {
	if c.apiKeyArgs == nil {
		return
	}
	jsonBody, err := json.Marshal(request)
	if err != nil {
		panic(fmt.Sprintf("Failed to marshal request body: %s", err.Error()))
	}
	body := string(jsonBody)
	if body == "null" {
		body = ""
	}

	c.apiKeyArgs.Timestamp = fmt.Sprint(time.Now().UnixMilli())
	sig := generateSignature(c.apiKeyArgs.KeySecret, c.apiKeyArgs.Timestamp, httpVerb, path, body)
	hexSig := hex.EncodeToString(sig)
	c.apiKeyArgs.Sign = hexSig

}

// getHeaders returns the headers for a request. It includes the auth headers and any extra headers set on the client.
func (c *ApiClient) getHeaders(httpVerb string, path string, request any) map[string]string {
	headers := c.getAuthHeaders(httpVerb, path, request)

	for k, v := range c.Headers {
		headers[k] = v
	}

	return headers
}

func (c *ApiClient) getAuthHeaders(httpVerb string, path string, request any) map[string]string {
	c.computeApiKeyArgs(httpVerb, path, request)
	headers := map[string]string{}
	if c.apiKeyArgs != nil {
		headers["ENCLAVE-KEY-ID"] = c.apiKeyArgs.KeyId
		headers["ENCLAVE-TIMESTAMP"] = c.apiKeyArgs.Timestamp
		headers["ENCLAVE-SIGN"] = c.apiKeyArgs.Sign
	}

	return headers
}

func NewApiClient(apiEndpoint string) *ApiClient {
	return &ApiClient{
		ApiEndpoint: apiEndpoint,
		Headers:     map[string]string{},
	}
}

func NewApiClientFromEnv(env string) (*ApiClient, error) {
	var apiUrl string
	switch strings.ToLower(env) {
	case "sandbox":
		apiUrl = "https://api-sandbox.enclave.market"
	case "prod":
		apiUrl = "https://api.enclave.market"
	default:
		return nil, fmt.Errorf("unknown env: %s", env)
	}

	return &ApiClient{
		ApiEndpoint: apiUrl,
		Headers:     map[string]string{},
	}, nil
}

func (client *ApiClient) WaitForEndpoint() {
	for {
		fmt.Println("waiting for the service to become available ......")
		if _, err := client.GetPublicStatus(); err != nil {
			time.Sleep(2 * time.Second)
			continue
		}
		return
	}
}

func (client *ApiClient) GetPublicStatus() (*models.GetPublicStatusRes, error) {
	path := models.StatusPath

	res, err := NewHttpJsonClient[any, models.GetPublicStatusRes](
		client.ApiEndpoint + path,
	).Get(nil)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (client *ApiClient) Hello() (*map[string]any, error) {
	path := models.HelloPath

	res, err := NewHttpJsonClient[any, map[string]any](
		client.ApiEndpoint + path,
	).Get(nil)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (client *ApiClient) AuthedHello() (*models.GenericResponse[string], error) {
	path := models.AuthedHelloPath

	jsonClient := NewHttpJsonClient[any, models.GenericResponse[string]](client.ApiEndpoint + path)
	jsonClient.SetHeaders(client.getHeaders("GET", path, nil))
	res, err := jsonClient.Get(nil)

	if err != nil {
		return nil, fmt.Errorf("error with http request to authed hello: %s", err)
	}

	if !res.Success {
		return res, fmt.Errorf("authed hello was not successful: %s", res.Error)
	}

	return res, err
}

func (client *ApiClient) Markets() (*models.GenericResponse[models.V1GetMarketsResult], error) {
	path := models.V1MarketsPath
	res, err := NewHttpJsonClient[any, models.GenericResponse[models.V1GetMarketsResult]](
		client.ApiEndpoint + path,
	).SetHeaders(client.getHeaders("GET", path, nil)).Get(nil)

	if err != nil {
		return nil, fmt.Errorf("error with http request to v1 markets: %w", err)
	}

	if !res.Success {
		return res, fmt.Errorf("error with getting v1 markets: %s", res.Error)
	}

	return res, nil
}

func (client *ApiClient) GetBalance(req models.GetBalanceReq) (*models.GenericResponse[models.V0GetBalanceRes], error) {
	path := models.V0GetBalancePath

	res, err := NewHttpJsonClient[models.GetBalanceReq, models.GenericResponse[models.V0GetBalanceRes]](
		client.ApiEndpoint + path,
	).SetHeaders(client.getHeaders("POST", path, req)).Post(req)

	if err != nil {
		return nil, fmt.Errorf("error with http request to get balance: %w", err)
	}

	if !res.Success {
		return nil, fmt.Errorf("error with getting balance %+v: %+v", req, res.Error)
	}

	return res, nil
}
