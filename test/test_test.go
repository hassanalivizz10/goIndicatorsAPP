package test

import (
	"fmt"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"
)

func MakeRequest() (interface{}, error) {
	url := "http://rules.digiebot.com/apiEndPoint/getAllCoinsHavingTradeSettings/1"
	client := http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("postman-token", "54617d62-35b8-4630-25df-7f512d389f6e")
	req.Header.Set("cache-control", "no-cache")
	req.Header.Set("mytoken", "OverLimit#_cgA3s8VSQj")
	req.Header.Set("content-type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check if the status code is 200
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Non-200 status code: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func TestHTTPRequest(t *testing.T) {
	response, err := MakeRequest()
	if err != nil {
		t.Fatalf("Error making HTTP request: %v", err)
	}

	fmt.Println("Response:")
	fmt.Println(response)
}
