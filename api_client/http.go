package api_client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/time/rate"
)

type HttpJsonClient[REQUEST_T any, REPLY_T any] struct {
	ApiEndpoint   string
	headers       map[string]string
	IsCSVResponse bool
}

func NewHttpJsonClient[REQUEST_T any, REPLY_T any](apiEndpoint string) *HttpJsonClient[REQUEST_T, REPLY_T] {
	return &HttpJsonClient[REQUEST_T, REPLY_T]{
		ApiEndpoint: apiEndpoint,
		headers:     map[string]string{},
	}
}

func (cl *HttpJsonClient[REQUEST_T, REPLY_T]) WithHeader(key string, value string) *HttpJsonClient[REQUEST_T, REPLY_T] {
	cl.headers[key] = value
	return cl
}

func (cl *HttpJsonClient[REQUEST_T, REPLY_T]) SetHeaders(headers map[string]string) *HttpJsonClient[REQUEST_T, REPLY_T] {
	for k, v := range headers {
		cl.headers[k] = v
	}
	return cl
}

func (cl *HttpJsonClient[REQUEST_T, REPLY_T]) Post(request REQUEST_T) (*REPLY_T, error) {
	return cl.Do("POST", request)
}

func (cl *HttpJsonClient[REQUEST_T, REPLY_T]) Get(request REQUEST_T) (*REPLY_T, error) {
	return cl.Do("GET", request)
}

func (cl *HttpJsonClient[REQUEST_T, REPLY_T]) Delete(request REQUEST_T) (*REPLY_T, error) {
	return cl.Do("DELETE", request)
}

var ErrEmptyResponseBody = fmt.Errorf("response body is empty")

func (cl *HttpJsonClient[REQUEST_T, REPLY_T]) Do(method string, request REQUEST_T) (*REPLY_T, error) {
	jsonStr, err := JsonSerializer[REQUEST_T]{}.ToJsonString(request)
	if err != nil {
		return nil, err
	}

	// DEBUG: Pretty print request
	// prettyJsonStr, err := serialization.JsonSerializer[REQUEST_T]{}.ToPrettyJsonString(request)
	// if err != nil {
	// 	return nil, err
	// }
	// fmt.Printf("%s %s\n%s\n", method, cl.ApiEndpoint, prettyJsonStr)

	var reqBody io.Reader
	// Allow the HttpJsonClient to be used for endpoints that don't expect a request body
	// by not sending one if the json representation of the request parameter is "null"
	if jsonStr != "null" {
		reqBody = bytes.NewBuffer([]byte(jsonStr))
	}

	req, err := http.NewRequest(method, cl.ApiEndpoint, reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	for k, v := range cl.headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if !(resp.StatusCode == 200 || resp.StatusCode == 201 || resp.StatusCode == 202) {
		reply, err := JsonSerializer[REPLY_T]{}.FromJsonString(string(body))
		err_text := fmt.Errorf("response: status=%d, body=%s", resp.StatusCode, body)
		if err != nil {
			return nil, err_text
		}
		return &reply, err_text
	}

	if cl.IsCSVResponse {
		var t any = &body
		conv, ok := t.(*REPLY_T)
		if !ok {
			return nil, errors.New("Failed to cast byte slice to empty interface")
		}
		return conv, nil
	} else {
		// this will occur when the success response is an empty body but sends "" instead of "{}" and (Hubspot API) and json.Unmarshal would error
		if len(body) == 0 {
			return nil, ErrEmptyResponseBody
		}

		reply, err := JsonSerializer[REPLY_T]{}.FromJsonString(string(body))
		return &reply, err
	}
}

type RateLimitedSecureHttpJsonClient[REQUEST_T any, REPLY_T any] struct {
	client      *HttpJsonClient[REQUEST_T, REPLY_T]
	rateLimiter *rate.Limiter
}

func NewRateLimitedSecureHttpJsonClient[REQUEST_T any, REPLY_T any](apiEndpoint string, rateLimiter *rate.Limiter) *RateLimitedSecureHttpJsonClient[REQUEST_T, REPLY_T] {
	return &RateLimitedSecureHttpJsonClient[REQUEST_T, REPLY_T]{
		client:      NewHttpJsonClient[REQUEST_T, REPLY_T](apiEndpoint),
		rateLimiter: rateLimiter,
	}
}

func (rlc *RateLimitedSecureHttpJsonClient[REQUEST_T, REPLY_T]) WithHeader(key string, value string) *RateLimitedSecureHttpJsonClient[REQUEST_T, REPLY_T] {
	rlc.client = rlc.client.WithHeader(key, value)
	return rlc
}

// This call will block until the rate limiter allows it through, avoid calling while holding a lock
func (rlc *RateLimitedSecureHttpJsonClient[REQUEST_T, REPLY_T]) Post(request REQUEST_T, ctx context.Context) (*REPLY_T, error) {
	err := rlc.rateLimiter.Wait(ctx)
	if err != nil {
		return nil, err
	}

	return rlc.client.Post(request)
}

func (rlc *RateLimitedSecureHttpJsonClient[REQUEST_T, REPLY_T]) Get(request REQUEST_T, ctx context.Context) (*REPLY_T, error) {
	err := rlc.rateLimiter.Wait(ctx)
	if err != nil {
		return nil, err
	}

	return rlc.client.Do("GET", request)
}

type JsonSerializer[T any] struct {
}

func (js JsonSerializer[T]) ToJsonString(x T) (string, error) {
	j, err := json.Marshal(x)
	if err != nil {
		return "", err
	}
	return string(j), nil
}

func (js JsonSerializer[T]) FromJsonString(s string) (T, error) {
	var x T
	err := json.Unmarshal([]byte(s), &x)
	if err != nil {
		return x, fmt.Errorf("failed json.Unmarshal.  Input: %s, err: %w", s, err)
	}
	return x, nil
}
