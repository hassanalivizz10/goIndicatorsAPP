package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"indicatorsAPP/helpers"
	"net/http"
	"strings"
	"time"
)

var httpClient = http.Client{
	Timeout: 5 * time.Second, // Set a timeout for the HTTP client
}

func TrendsPostCron() {
	fmt.Println("Inside TrendsPostCron Cron")
	if len(coinListCacheData) == 0 {
		coinList, err := helpers.ListCoins()
		if err != nil {
			fmt.Println("TrendsPostCron Error", err)
			return
		}
		coinListCacheData = coinList
	}

	for _, coin := range coinListCacheData {
		coinSymbol := coin["symbol"].(string)
		if coinSymbol != "QTUMBTC" {
			continue
		}
		trend, err := helpers.CurrentCandelTrend(coinSymbol)
		fmt.Println("err", err)
		if err != nil {
			continue // something went wrong.
		}
		fmt.Println("trend", trend)
		// Check if trend is not empty
		if trendValue, ok := trend["candel_trends"].(string); ok && trendValue != "" {
			// Replace "strong_" from trend value
			trendValue = strings.Replace(trendValue, "strong_", "", 1)
			fmt.Println("trendValue", trendValue)
			// Post trend value
			if err := postTrendValue(coinSymbol, trendValue); err != nil {
				fmt.Printf("Error posting trend value for %s: %v\n", coinSymbol, err)
			}
		} else {
			fmt.Printf("Empty trend value for %s\n", coinSymbol)
		}
	}
}

func postTrendValue(coin, trend string) error {
	url := "https://rules.digiebot.com/apiEndPoint/addUpdateTrendValue"

	// Define the request body
	requestBody, err := json.Marshal(map[string]string{
		"coin":  coin,  // Example coin value
		"trend": trend, // Example trend value
	})
	fmt.Println("err", err)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}

	req.Header.Set("authorization", "trendValues#_cgA3s8VSQj")
	req.Header.Set("content-type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
