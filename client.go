package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func NewOpenAIClient(apiKey string) *OpenAIClient {
	return &OpenAIClient{ApiKey: apiKey}
}

func (c *OpenAIClient) CreateResponse(req ResponseRequest) (ResponseBody, error) {
	var resp ResponseBody

	reqBody, err := json.Marshal(req)
	if err != nil {
		return resp, err
	}

	httpReq, err := http.NewRequest("POST", "https://api.openai.com/v1/responses", bytes.NewBuffer(reqBody))
	if err != nil {
		return resp, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.ApiKey))

	httpResp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return resp, err
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return resp, err
	}

	err = json.Unmarshal(respBody, &resp)
	if err != nil {
		return resp, err
	}

	return resp, nil
}
