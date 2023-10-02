package test

import (
	"fmt"
	"time"
	"io/ioutil"
	"net/http"
	"testing"
)

func MakeRequest() (string, error) {
	url := "http://rules.digiebot.com/apiEndPoint/getAllCoinsHavingTradeSettings/1"
	//url = "http://rules.digiebot.com/goTest"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	token:= "OverLimit#_cgA3s8VSQj"
		// Set headers to replicate POSTMAN headers
		req.Header.Add("myToken", token) // Replace with your actual token
		req.Header.Add("User-Agent", "PostmanRuntime/7.33.0")
		req.Header.Add("Accept", "*/*")
		req.Header.Add("Accept-Encoding", "gzip, deflate, br")
		req.Header.Add("Cache-Control", "no-cache")
		req.Header.Add("Connection", "keep-alive")
		req.Header.Add("Host", "localhost:3001") // Replace with your Node.js server's hostname and port
	//req.Header.Set("Content-Type", "application/json")
	//req.Header.Set("Authorization", "OverLimit#_ubN7iC5W7D")
	// Print the request headers for debugging
	fmt.Println("Request Headers:")
    fmt.Println(req.Header)
	// Set a timeout for the request
	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	fmt.Println("Response Status Code:", resp.Status)
fmt.Println("Response Headers:")
fmt.Println(resp.Header)

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
