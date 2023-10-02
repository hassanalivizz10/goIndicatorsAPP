package controller
import (
	"fmt"
	//"time"
	"github.com/go-resty/resty/v2"
	//"io/ioutil"
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
	

	url := "http://rules.digiebot.com/apiEndPoint/getAllCoinsHavingTradeSettings/1"
	
	// Set request headers
	// Set custom headers
    headers := map[string]string{
        "Authorization": "overlimitrequest",        // Replace with your authorization token
        "Content-Type":  "application/json",            // Add any other headers you need
    }	
	client := resty.New()

    request := client.R().SetHeaders(headers)

    
    resp, err := request.Execute("GET", url)
    if err != nil {
        return nil, err
    }
    fmt.Println("result",resp.String())	
	 // Read and parse the JSON response into a struct
	 var result StrategyResponse
	err = json.Unmarshal(resp.Body(),&result)
    if err != nil {
        fmt.Println(err)
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