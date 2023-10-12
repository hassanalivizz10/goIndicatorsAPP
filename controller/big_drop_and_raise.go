package controller
import (
	"fmt"
	"sync"
	"time"
	"net/http"
	"errors"
	"io/ioutil"
	//"reflect"
	"encoding/json"
	"strconv"
	"indicatorsAPP/helpers"
	//"indicatorsAPP/mongohelpers"
	
	"go.mongodb.org/mongo-driver/bson"
)


var lastHourReset time.Time
var openPrices []CoinData
var coinListCache []bson.M
var bigDropRaiseMutex sync.Mutex
// Defaults ....
var bigDropFactorValue  float64  = 3
var bigRaiseFactorValue float64 = 3.5

type CoinData struct {
	Symbol                 string 
	OpenPrice              float64
	PerMove                float64
	DropPercValue          float64
	RaisePercvalue         float64
	DropTrailingPrice      float64
	RaiseTrailingPrice     float64
	PricesFound            int
	DropFound			   bool
	RaiseFound			   bool
	Date                   time.Time   
}



/*
bigDropRaiseMutex.Lock() // Acquire the mutex
// Access the shared resource
bigDropRaiseMutex.Unlock() // Release the mutex when done
*/
//ruleType := "buy"
func RunBigRaiseAndBigDrop(){
	currentDateTime := time.Now().UTC()
	currentHourDate := time.Date(currentDateTime.Year(), currentDateTime.Month(), currentDateTime.Day(), currentDateTime.Hour(), 0, 0, 0, currentDateTime.Location())

	if lastHourReset.IsZero() || currentHourDate.After(lastHourReset) {
		fmt.Println("Listing coins")
		openPrices = nil
		lastHourReset = currentHourDate
	}

	if len(coinListCache) ==0 {
		coinList , err := helpers.ListCoins()
		if err!=nil{
			fmt.Println("Big Raise and Drop Error",err)
			return
		}
		coinListCache = coinList
	}
	//fmt.Println("coinListCache",coinListCache)
	//coinList a bson.M document
	for _, currentCoin := range coinListCache {
		coinSymbol := currentCoin["symbol"].(string)
		fmt.Println("coinSymbol"+coinSymbol,time.Now().UTC())
		bigDropRaiseMutex.Lock()
		foundPriceObject := findPriceObject(coinSymbol, currentHourDate)
		if foundPriceObject != nil {
			//fmt.Println("Big Drop Found in foundPriceObject for coin:", coinSymbol, *foundPriceObject)
			bigDropRaiseMutex.Unlock()
			continue
		}
		bodyMoveAverage, err := helpers.GetBodyMoveAverage(coinSymbol)
		//fmt.Println("bodyMoveAverage",bodyMoveAverage)
		if err!=nil{
			fmt.Println("bodyMoveAverage Error "+coinSymbol,err)
			bigDropRaiseMutex.Unlock()
			continue
		}
		if len(bodyMoveAverage) == 0 || len(bodyMoveAverage[0]) == 0 {
			fmt.Println("bodyMoveAverage is empty "+coinSymbol,bodyMoveAverage)
			bigDropRaiseMutex.Unlock()
			continue
		}
		
		var body_move_average float64
		if val , ok := bodyMoveAverage[0]["body_move_average"].(float64); ok{
			body_move_average = val
		} else {
			fmt.Println("bodyMoveAverage is missing ",bodyMoveAverage[0])
			continue;
		}
		DropPercValue  := body_move_average *  bigDropFactorValue
		RaisePercvalue := body_move_average * bigRaiseFactorValue
		currentOpenPrice, err := getOpenPrice(coinSymbol)
		if err != nil {
			fmt.Println("Error fetching open price for", coinSymbol, ":", err)
			bigDropRaiseMutex.Unlock()
		}

		DropTrailingPrice  := currentOpenPrice - ((currentOpenPrice * DropPercValue) / 100)
		RaiseTrailingPrice :=  currentOpenPrice + ((currentOpenPrice * RaisePercvalue) / 100)
		priceObject := CoinData{
			Symbol:      coinSymbol,
			OpenPrice:   currentOpenPrice,
			PerMove:     body_move_average,
			DropPercValue:     DropPercValue,
			RaisePercvalue : RaisePercvalue,
			DropTrailingPrice : DropTrailingPrice,
			RaiseTrailingPrice : RaiseTrailingPrice,
			Date:        currentHourDate,
			PricesFound: 0,
			RaiseFound : false,
			DropFound : false,
			
		}

		openPrices = append(openPrices, priceObject)
		bigDropRaiseMutex.Unlock()
	} // ends coinListCache forLoop
	fmt.Println("Time Now Data",time.Now().UTC())
	
	for _, priceData := range openPrices {
		dataToParse := priceData
		coin := dataToParse.Symbol
		
		raisePrice := dataToParse.RaiseTrailingPrice
		dropPrice := dataToParse.DropTrailingPrice
		bodyMoveValue := dataToParse.PerMove
		OpenPrice := dataToParse.OpenPrice
		fmt.Println("openPrices started"+coin,"raisePrice",raisePrice,"dropPrice",dropPrice,"OpenPrice",OpenPrice)
		dateNow := time.Now().UTC()
		startTime := getStartTime(dateNow)
		RaiseFound :=  false
		DropFound :=  false

		if hasTimeChanged(lastHourReset) {
			fmt.Println("Breaking big Drop for", lastHourReset, time.Now().UTC())
			break
		}
		pricesData , err := helpers.FetchMarketPrices(coin, startTime)
		if err!=nil{
			fmt.Println("ERROR ON FetchMarketPrices"+coin,err)
		}
		if len(pricesData) == 0  || len(pricesData[0]) == 0 {
			fmt.Println("Found Empty on FetchMarketPrices"+coin)
		}
		var currentPrice float64
	
		//fmt.Println("pricess",pricesData[0]["price"])
		currentPrice , ok := helpers.ToFloat64(pricesData[0]["price"])
		if !ok{
			fmt.Println("currentPrice Unsupported numeric type errored")
			continue;
		}
		//currentPrice := pricesData[0]["price"].(float64)
		if currentPrice > raisePrice {
			RaiseFound = true
			raiseFilters := bson.M{
				"coin":coin,
				"type":"big_raise_pull_back",
			}
			upsert:= true
			raiseUpdate := bson.M{
				"$set":bson.M{
					"coin":coin,
					"type":"big_raise_pull_back",
					"open_price":OpenPrice,
					"pull_back_price": currentPrice - ((currentPrice*bodyMoveValue)/100),
					"created_date":time.Now().UTC(),
					"raise_price":currentPrice,
					"trailing_price":raisePrice,
					"move":bodyMoveValue,
				},
			}
			err:= helpers.AddRaiseDropEntry(raiseFilters, raiseUpdate , upsert)
			if err!=nil{
				fmt.Println("AddRaiseDropEntry has ERRORED for BIG RAISE FOUND"+coin,err)
			}
		} 

		if currentPrice < dropPrice {
			DropFound = true
			raiseFilters := bson.M{
				"coin":coin,
				"type":"big_drop_pull_back",
			}
			upsert:= true
			raiseUpdate := bson.M{
				"$set":bson.M{
					"coin":coin,
					"type":"big_drop_pull_back",
					"open_price":OpenPrice,
					"pull_back_price": currentPrice + ((currentPrice*bodyMoveValue)/100),
					"created_date":time.Now().UTC(),
					"drop_price":currentPrice,
					"trailing_price":dropPrice,
					"move":bodyMoveValue,
				},
			}
			err:= helpers.AddRaiseDropEntry(raiseFilters, raiseUpdate , upsert)
			if err!=nil{
				fmt.Println("AddRaiseDropEntry has ERRORED for BIG RAISE FOUND"+coin,err)
			}
		} 
		fmt.Println("RaiseFound"+coin,RaiseFound)
		fmt.Println("DropFound"+coin,DropFound)
		
	} // ends openPrices loop
}

func hasTimeChanged(lastHourReset time.Time) bool {
	return time.Now().UTC().Hour() != lastHourReset.Hour()
}

func getOpenPrice(coinSymbol string) (float64, error) {
	url := fmt.Sprintf("https://api.binance.com/api/v3/klines?symbol=%s&interval=1h&limit=1", coinSymbol)

	client := http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, errors.New(fmt.Sprintf("Request failed with status code: %d", resp.StatusCode))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var jsonData [][]interface{}
	if err := json.Unmarshal(body, &jsonData); err != nil {
		return 0, err
	}

	if len(jsonData) > 0 && len(jsonData[0]) > 1 {
		op, err := strconv.ParseFloat(jsonData[0][1].(string), 64)
		if err != nil {
			return 0, err
		}
		return op, nil
	}
	return 0, errors.New("Invalid response format")
}

func findPriceObject(coinSymbol string, currentHourDate time.Time) *CoinData {
	for i := range openPrices {
		priceData := &openPrices[i]
		objectDate := priceData.Date
		if priceData.Symbol == coinSymbol && objectDate.Equal(currentHourDate) {
			return priceData
		}
	}
	return nil
}


func getStartTime(dateNow time.Time) time.Time {
	var startTime time.Time
	switch {
	case dateNow.Minute() >= 0 && dateNow.Minute() <= 15:
		startTime = time.Date(dateNow.Year(), dateNow.Month(), dateNow.Day(), dateNow.Hour(), 0, 0, 0, dateNow.Location())
	case dateNow.Minute() > 15 && dateNow.Minute() <= 30:
		startTime = time.Date(dateNow.Year(), dateNow.Month(), dateNow.Day(), dateNow.Hour(), 15, 0, 0, dateNow.Location())
	case dateNow.Minute() > 30 && dateNow.Minute() <= 45:
		startTime = time.Date(dateNow.Year(), dateNow.Month(), dateNow.Day(), dateNow.Hour(), 30, 0, 0, dateNow.Location())
	case dateNow.Minute() > 45 && dateNow.Minute() <= 59:
		startTime = time.Date(dateNow.Year(), dateNow.Month(), dateNow.Day(), dateNow.Hour(), 45, 0, 0, dateNow.Location())
	}
	return startTime
}