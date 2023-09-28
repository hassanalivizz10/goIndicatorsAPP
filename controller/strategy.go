package controller
import (
	"fmt"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"indicatorsAPP/helpers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)
var coinListCacheData []bson.M
type StrategyResponse struct {
	Status  int         `json:"status"`
	Result  []CoinDataSet  `json:"result"`
	Message string      `json:"message"`
}

type CoinDataSet struct {
	Coin  string `json:"coin"`
	Value string `json:"value"`
}
func StrategyCron(){
	fmt.Println("Inside Strategy Cron")
	strategies, err := getStrategies()
	if err != nil {
		fmt.Println("Error fetching strategies:", err)
		return
	}

	if len(coinListCacheData) ==0 {
		coinList , err := helpers.ListCoins()
		if err!=nil{
			fmt.Println("StrategyCron Error",err)
			return
		}
		coinListCacheData = coinList
	}
	fmt.Println("strategies",strategies)
	for _, coin := range coinListCacheData {
		strategyValue := "none"
		coinSymbol := coin["symbol"].(string)
		val, exists := getStrategyValue(coinSymbol, strategies.Result)
		if exists{
			strategyValue = val
			
		} 
		fmt.Printf("Strategy value for %s: %s\n", coinSymbol, strategyValue)
		// get the latest candle data
		data , err := helpers.GetBodyMoveAverage(coinSymbol)
		if err!=nil{
			fmt.Println("StrategyCron Error on Getting CUrrent Candle for Coin"+coinSymbol,err)
			continue
		}
		if len(data) == 0 || len(data[0]) == 0 {
			fmt.Println("StrategyCron Data Empty For CUrrent Candle for Coin"+coinSymbol)
			continue
		}	

		if val , ok := data[0]["_id"].(primitive.ObjectID); ok{
			err := helpers.UpdateStrategyValue(val,strategyValue,coinSymbol)
			if err!=nil{
				fmt.Println("StrategyCron Updation Error Coin"+coinSymbol,err.Error())
				continue
			}
		} else {
			fmt.Println("StrategyCron Data ID is null for Coin"+coinSymbol)
			continue
		}
	
	}
}

func getStrategies() (*StrategyResponse, error) {
	// url := "http://rules.digiebot.com/apiEndPoint/getAllCoinsHavingTradeSettings/1"
	// client := http.Client{}
	// req, err := http.NewRequest("POST", url, nil)
	// if err != nil {
	// 	fmt.Println("err",err)
	// 	return nil, err
	// }

	// req.Header.Add("postman-token", "54617d62-35b8-4630-25df-7f512d389f6e")
	// req.Header.Add("cache-control", "no-cache")
	// req.Header.Add("Authorization", "OverLimit#_cgA3s8VSQj")
	// req.Header.Add("content-type", "application/json")

	// resp, err := client.Do(req)
	// if err != nil {
	// 	fmt.Println("err",err)
	// 	return nil, err
	// }
	// defer resp.Body.Close()
	// body, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	return nil, err
	// }
	// fmt.Println("body",resp.Body)

	// // Check if the status code is 200
	// if resp.StatusCode != http.StatusOK {
	// 	return nil, fmt.Errorf("Non-200 status code: %d", resp.StatusCode)
	// }

	url := "http://rules.digiebot.com/apiEndPoint/getAllCoinsHavingTradeSettings/1"
	method := "POST"

	client := &http.Client {
	}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return nil , err
	} 
	   // Set the authorization header to one of the expected values
	   req.Header.Add("Authorization", "OverLimit#_cgA3s8VSQj") // Or use "OverLimit#_ubN7iC5W7D"
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil , err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return nil , err
	}
	fmt.Println(string(body))
	var result StrategyResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func getStrategyValue(coinToCheck string, strategies []CoinDataSet) (string, bool) {
	for _, coin := range strategies {
		if coin.Coin == coinToCheck {
			return coin.Value, true
		}
	}
	return "", false
}