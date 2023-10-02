package test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
)

func MakeRequest() (string, error) {
	url := "http://rules.digiebot.com/apiEndPoint/getAllCoinsHavingTradeSettings/1"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "OverLimit#_ubN7iC5W7D")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func TestHTTPRequest(t *testing.T) {
	response, err := MakeRequest()
	if err != nil {
		t.Fatalf("Error making HTTP request: %v", err)
	}

	fmt.Println("Response:")
	fmt.Println(response)
}
