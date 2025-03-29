package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func NewOpenAIClient(apiKey string) *OpenAIClient {
	return &OpenAIClient{APIKey: apiKey}
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
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))

	httpResp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return resp, err
	}
	defer httpResp.Body.Close()

	var respBody []byte
	respBody, err = io.ReadAll(httpResp.Body)
	if err != nil {
		return resp, err
	}

	// for debugging
	// fmt.Printf("%s", respBody)
	// fmt.Printf("Status Code: %d\n", httpResp.StatusCode)

	switch httpResp.StatusCode {
	case http.StatusBadRequest:
		var errorResponse map[string]string
		err = json.Unmarshal(respBody, &errorResponse)
		if err != nil {
			return resp, fmt.Errorf("failed to unmarshal 400 error response: %v", err)
		}
		return resp, fmt.Errorf("400 Bad Request: %v", errorResponse["error"])
	case http.StatusOK:
		var incompleteResponse map[string]interface{}
		err = json.Unmarshal(respBody, &incompleteResponse)
		if err != nil {
			return resp, err
		}
		if status, ok := incompleteResponse["status"].(string); ok && status == "incomplete" {
			return resp, fmt.Errorf("incomplete response: %v", incompleteResponse["incomplete_details"])
		}
		err = json.Unmarshal(respBody, &resp)
		if err != nil {
			return resp, err
		}
		return resp, nil
	default:
		return resp, fmt.Errorf("non-OK response code: %d", httpResp.StatusCode)
	}
}
