package api

import (
	"fmt"
	"io/ioutil"
	"sort"
	"time"
	//"reflect"
	"strconv"
	"encoding/json"
	"indicatorsAPP/helpers"
	"net/http"
	//"indicatorsAPP/mongohelpers"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"github.com/patrickmn/go-cache"
)

type Candle struct {
    OpenTime int64  	    `json:"timestampDate"` 
    Open     float64 		`json:"open"`
    Coin     string  		`json:"coin"`
    Low      float64 		`json:"low"`
    High     float64 		`json:"high"`
    Volume   float64 		`json:"volume"`
    Close    float64 		`json:"close"`
}
var customLayout string
var myCache = cache.New(5*time.Minute, 10*time.Minute)

func init() {
	customLayout = "2006-01-02 15:04:05"
}

// convert unix TimeStamp to DateTime (start of Hour)
// Helper function to parse and format the start_date
func formatStartDate(startDate interface{}) (string, error) {
	// Check the data type of the interface variable

    switch v := startDate.(type) {
    case float64:
        ts := int64(v)
        startDate := time.Unix(ts, 0)
        startDate = startDate.Truncate(time.Hour)
        return startDate.Format(customLayout), nil
    case string:
        ts, err := strconv.ParseInt(v, 10, 64)
        if err != nil {
            return "", err
        }
        startDate := time.Unix(ts, 0)
        startDate = startDate.Truncate(time.Hour)
        return startDate.Format(customLayout), nil
    default:
        return "", fmt.Errorf("Invalid start date format")
    }
}

// Helper function to parse and calculate the new date based on the duration
func calculateNewDate(start time.Time, duration interface{}) (time.Time, error) {
    switch v := duration.(type) {
    case float64:
        // Convert the duration to an integer number of weeks
        weeks := int(v)
        // Subtract the number of weeks from the start date
        newDate := start.AddDate(0, 0, -7*weeks)
		// Set minutes and seconds to zero to represent the start of the hour
        newDate = newDate.Truncate(time.Hour)
        return newDate, nil
    case string:
        // Attempt to convert the string to an integer
        weeks, err := strconv.Atoi(v)
        if err != nil {
            return time.Time{}, err
        }
        // Subtract the number of weeks from the start date
        newDate := start.AddDate(0, 0, -7*weeks)
		// Set minutes and seconds to zero to represent the start of the hour
        newDate = newDate.Truncate(time.Hour)
        return newDate, nil
    default:
        return time.Time{}, fmt.Errorf("Invalid duration format")
    }
}

// POST METHOD TO Get Candle Charts Data for Trading View Chart
func FetchTradingDataByCoinHandler(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader != "tradingChart#(njVEkn2AEZ" && authHeader != "tradingChart#CJOMTGzhrB4" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  401,
			"allCandle":    nil,
			"error":   "Auth Failed",
			"message": "Request Blocked",
		})
		return
	}
//	var errors []string
	requestBody, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  400,
			"allCandle":    nil,
			"error":   "Failed to read request body",
			"message": "Error On Your Request",
		})
		return
	}
	// Attempt to unmarshal the request body as JSON
	var jsonData map[string]interface{}
	if err := json.Unmarshal(requestBody, &jsonData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  400,
			"allCandle":    nil,
			"error":   "Invalid JSON Data",
			"message": "Error On Your Request",
		})
		return
	}
	if len(jsonData) != 0 {
		coin, coinOK := jsonData["coin"].(string)
		requestType, typeOK := jsonData["type"].(string)
		durationStr, durationOK := jsonData["duration"].(string)
		startDateUnix, startOK := jsonData["start_date"]
		// if values not in the expected data types...
		if !coinOK || !typeOK || !durationOK || !startOK {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  400,
				"allCandle":    nil,
				"error":   "Invalid request parameters",
				"message": "Error On Your Request",
			})
			return
		}
		//  Parse and format the unixtimestamp start_date using the helper function 
		startDateStr, err := formatStartDate(startDateUnix)
		fmt.Println("startDateStr err",err)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  400,
				"allCandle":    nil,
				"error":   "Invalid start date format",
				"message": "Error On Your Request",
			})
			return
		} 
		fmt.Println("startDateStr",startDateStr)
		// Convert the startDateStr to time.Time with the custom layout
		dateNow, err := time.Parse(customLayout, startDateStr)
		fmt.Println("dateNow err",err)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  400,
				"allCandle":    nil,
				"error":   "Invalid start date format",
				"message": "Error On Your Request",
			})
			return
		}

		dateFrom, err := calculateNewDate(dateNow,durationStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  400,
				"data":    nil,
				"error":   "Invalid duration format",
				"message": "Error On Your Request",
			})
			return
		}

		fmt.Println("dateFrom",dateFrom)
		fmt.Println("dateNow",dateNow)
		
		// Create the cache key
		cacheKey := fmt.Sprintf("%s_%s_%d_%s", coin, requestType, durationStr, dateNow.Format(customLayout))	
		if cachedResponse, found := myCache.Get(cacheKey); found {
			// Return the cached response
			c.JSON(http.StatusOK, gin.H{
				"status":  200,
				"allCandle":    cachedResponse,
				"error":   "",
				"message": "Data Returned Successfully",
			})

			return
		} else {

			data , err := helpers.FetchCandlesData(coin,requestType,dateFrom,dateNow)
			if err!=nil || len(data) == 0{
				c.JSON(http.StatusBadRequest, gin.H{
					"status":  400,
					"allCandle":    nil,
					"error":   err.Error(),
					"message":  "Error On Your Request",
				})
			}
			response := buildCandlesData(data)
			myCache.Set(cacheKey, response, cache.DefaultExpiration)


			// query the data and cache it.
			c.JSON(http.StatusOK, gin.H{
				"status":  200,
				"allCandle":    response,
				"error":   "",
				"message": "Data Returned Successfully",
			})
			return
		}

	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  400,
			"allCandle":    nil,              // Replace with your data
			"error":   "Params Missing", // Handle errors if necessary
			"message": "Error On Your Request",
		})
		return

	}
}

// Build Candle response
func buildCandlesData(bsonData []bson.M) []Candle {
	var candles []Candle
	for _, doc := range bsonData {
		var candle Candle
		coin := doc["coin"].(string)
		close , err :=  helpers.ConvertToType(doc["close"],float64(0))
		if err!=nil{
			fmt.Println("ERROR on buildCandlesData for close conversion",err.Error())
			continue
		}
		high , err :=  helpers.ConvertToType(doc["high"],float64(0))
		if err!=nil{
			fmt.Println("ERROR on buildCandlesData for high conversion",err.Error())
			continue
		}
		low , err :=  helpers.ConvertToType(doc["low"],float64(0))
		if err!=nil{
			fmt.Println("ERROR on buildCandlesData for low conversion",err.Error())
			continue
		}
		open , err :=  helpers.ConvertToType(doc["open"],float64(0))
		if err!=nil{
			fmt.Println("ERROR on buildCandlesData for open conversion",err.Error())
			continue
		}
		openTime , err :=  helpers.ConvertToType(doc["openTime"],int64(0))
		if err!=nil{
			fmt.Println("ERROR on buildCandlesData for openTime conversion",err.Error())
			continue
		}
		volume , err :=  helpers.ConvertToType(doc["volume"],float64(0))
		if err!=nil{
			fmt.Println("ERROR on buildCandlesData for volume conversion",err.Error())
			continue
		}
		candle.OpenTime = openTime.(int64)
		candle.High = high.(float64)
		candle.Coin = coin
		candle.Low = low.(float64)
		candle.Close = close.(float64)
		candle.Open = open.(float64)
		candle.Volume = volume.(float64)
		
		candles = append(candles, candle)
	}
	return candles
}

// POST METHOD to Calculate Daily Indicators
func CreateUpdateDailyIndicatorsHandler(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader != "indicators#(njVEkn2AEZ" && authHeader != "indicators#CJOMTGzhrB4" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  401,
			"data":    nil,
			"error":   "Auth Failed",
			"message": "Request Blocked",
		})
		return
	}

	var errors []string
	requestBody, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(400, gin.H{
			"error": "Failed to read request body",
		})
		return
	}

	// Attempt to unmarshal the request body as JSON
	var jsonData map[string]interface{}
	if err := json.Unmarshal(requestBody, &jsonData); err != nil {
		c.JSON(400, gin.H{
			"error": "Invalid JSON data",
		})
		return
	}
	fmt.Println("jsonData", jsonData)
	fmt.Println(len(jsonData))
	if len(jsonData) != 0 {
		if coin := jsonData["coin"].(string); coin == "" {
			errors = append(errors, "Coin is Required Field")
		}

		startDateStr := jsonData["start_date"].(string) //c.Query("start_date")
		startDate, err := time.Parse(customLayout, startDateStr)
		if err != nil {
			errors = append(errors, "Start Date has Invalid Format")
		}

		endDateStr := jsonData["end_date"].(string) //c.Query("end_date")
		var diffInDays int = 0
		if endDateStr != "" {
			endDate, err := time.Parse(customLayout, endDateStr)
			if err != nil {
				errors = append(errors, "End Date has Invalid Format")
			}

			diffInDays = int(endDate.Sub(startDate).Hours() / 24)
		}
		fmt.Println("errors", errors)
		if len(errors) > 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  400,
				"data":    nil,              // Replace with your data
				"error":   "Params Missing", // Handle errors if necessary
				"message": "Error On Your Request",
			})
			return
		}
		coin := jsonData["coin"].(string)

		responseSlice := []map[string]interface{}{}
		for i := 0; i <= diffInDays; i++ {
			setStartDate := startDate.AddDate(0, 0, diffInDays) // years, months, days
			// Set the time components for the start and end dates
			fromDate := time.Date(setStartDate.Year(), setStartDate.Month(), setStartDate.Day(), 0, 0, 0, 0, time.UTC)
			toDate := time.Date(setStartDate.Year(), setStartDate.Month(), setStartDate.Day(), 23, 59, 59, 999999999, time.UTC)
			dpupDirection, _ := dpupTrendDirectionCalculations(coin, fromDate, toDate)
			getTrend, _ := calculateDailyTrend(coin, fromDate, toDate)
			response := map[string]interface{}{
				"dpup_direction": dpupDirection,
				"getTrend":       getTrend,
			}
			responseSlice = append(responseSlice, response)
		}
		c.JSON(http.StatusOK, gin.H{
			"status":  200,
			"data":    responseSlice, // Replace with your data
			"error":   nil,           // Handle errors if necessary
			"message": "Success",
		})
		return
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  400,
			"data":    nil,              // Replace with your data
			"error":   "Params Missing", // Handle errors if necessary
			"message": "Error On Your Request",
		})
		return

	}
}

// POST METHOD to Set Hourly Indicators
func SetHourlyIndicatorsHandler(c *gin.Context) {
	fmt.Println("SetHourlyIndicatorsHandler")
	authHeader := c.GetHeader("Authorization")
	if authHeader != "indicators#(njVEkn2AEZ" && authHeader != "indicators#CJOMTGzhrB4" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  401,
			"data":    nil,
			"error":   "Auth Failed",
			"message": "Request Blocked",
		})
		return
	}

	var errors []string

	// Parse the request JSON body
	var requestBody map[string]interface{}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		errors = append(errors, "Invalid JSON format")
	}
	if len(errors) == 0 {
		if coin, ok := requestBody["coin"].(string); !ok || coin == "" {
			errors = append(errors, "Coin is a required field")
		}
		if startDateStr, ok := requestBody["start_date"].(string); !ok || startDateStr == "" {
			errors = append(errors, "Start Date is a required field")
		} else {
			_, err := time.Parse("2006-01-02 15:04:05", startDateStr)
			if err != nil {
				errors = append(errors, "Start Date has an invalid format")
			}
			if endDateStr, ok := requestBody["end_date"].(string); ok && endDateStr != "" {
				_, err = time.Parse("2006-01-02 15:04:05", endDateStr)
				if err != nil {
					errors = append(errors, "End Date has an invalid format")
				}
			}
		}
		if len(errors) == 0 {
			startDateStr := requestBody["start_date"].(string)
			endDateStr := requestBody["end_date"].(string)
			coin := requestBody["coin"].(string)
			hourDifference, startDate, err := calculateHourDifference(startDateStr, endDateStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"status":  400,
					"data":    nil,
					"error":   errors,
					"message": "Error On Your Request",
				})
				return
			}
			var resArr []interface{}
			var candel_trends []bson.M
			for i := 0; i <= hourDifference; i++ {
				currentStartDate := startDate.Add(time.Hour * time.Duration(i))
				currentEndDate := currentStartDate.Add(time.Hour - time.Second)
				where := bson.M{}
				where["coin"] = coin
				where["created_date"] = bson.M{"$gte": currentStartDate, "$lte": currentEndDate}
				fmt.Println("where", where)
				chartDataFetched, err := helpers.MarketChartDataForCoin(where)
				if err != nil {
					fmt.Println("Error on MarketChartDataForCoin for Date")
					fmt.Printf("Hour %d: Start Date: %s, End Date: %s\n", i, currentStartDate.Format("2006-01-02 15:04:05"), currentEndDate.Format("2006-01-02 15:04:05"))
					continue
				}
				//fmt.Println("chartDataFetched", chartDataFetched)
				if len(chartDataFetched) == 0 {
					resArr = append(resArr, "NO Data Found")
					continue
				}
				//fmt.Println("chartDataFetched", chartDataFetched)
				currentData := chartDataFetched[0]
				open := currentData["open"].(float64)
				close := currentData["close"].(float64)
				coin := currentData["coin"].(string)
				openTime_human_readible := currentData["openTime_human_readible"].(string)
				filters := bson.M{
					"coin":coin,
					"openTime_human_readible":openTime_human_readible,
				}
				update := bson.M{}	

				DP_UP_of_Candle := CalculateDPUPOfCandle(coin, currentStartDate, open, close)
				

				
				//fmt.Println("DP_UP_of_Candle",DP_UP_of_Candle)
				
				toArr := make(map[string]interface{})
				for key, value := range DP_UP_of_Candle {
					toArr[key] = value
				}
				toArr["startCandleDate"] = currentStartDate
				toArr["endCandleDate"] = currentEndDate
				temp := map[string]interface{}{

					"openTime_human_readible": openTime_human_readible,
				}
				resArr = append(resArr, temp)
				DP_UP_of_Candleperc := CalculateDPUPPercentiles(coin,toArr,30)
				
				
				candel_trends = processCandelTrends(coin, currentStartDate, open, close)
				
				
				toUpdate := BuildHourlyUpdateArr(coin,DP_UP_of_Candle,DP_UP_of_Candleperc,candel_trends)
				update = bson.M{
					"$set":toUpdate,
				}
				err = helpers.UpdateHourlyData(filters,update)
				if err!=nil{
					fmt.Println("ERROR ON UpdateHourlyData main Call ",err.Error())
				}
				if coin == "BTCUSDT"{
					updateAllCoinsBTC := modifyAndFilterKeysForBTC(toUpdate)
					toUpdateBTCData := bson.M{"$set":updateAllCoinsBTC}
					otherCoinsFilters := bson.M{"openTime_human_readible":openTime_human_readible,"coin": bson.M{
						"$nin": []string{"BTCUSDT","POEBTC"},
					}}
					err := helpers.UpdateHourlyData(otherCoinsFilters,toUpdateBTCData)
					if err!=nil{
						fmt.Println("ERROR ON UpdateHourlyData main Call ",err.Error())
					}
				}

				resArr = append(resArr, DP_UP_of_Candle)
				resArr = append(resArr,DP_UP_of_Candleperc)

				
			}

			c.JSON(http.StatusOK, gin.H{
				"status": 200,
				"data":   requestBody,
				"response": map[string]interface{}{
					"dpup":          resArr,
					"candel_trends": candel_trends,
				},
				"error":   nil,
				"message": "Successfully Done",
			})
			return
		} else { // inner if errors
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  400,
				"data":    nil,
				"error":   errors,
				"message": "Error On Your Request",
			})
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  400,
			"data":    nil,
			"error":   errors,
			"message": "Error On Your Request",
		})
		return
	} // ends outer if no errors
}

// FOR BTC 
func modifyAndFilterKeysForBTC(input bson.M) bson.M {
	result := bson.M{}

	for key, value := range input {
		// Check if the key is one of the keys that should be modified
		if key == "DP1" || key == "DP2" || key == "DP3" ||
			key == "UP1" || key == "UP2" || key == "UP3" ||
			key == "DP1_perc" || key == "DP2_perc" || key == "DP3_perc" ||
			key == "UP1_perc" || key == "UP2_perc" || key == "UP3_perc" || key == "new_go_indicators" {
			newKey := key + "_btc"
			result[newKey] = value
		}
	}

	return result
}

func BuildHourlyUpdateArr(coin string,DP_UP_of_Candle map[string]float64,DP_UP_of_Candleperc map[string]string,candel_trends []bson.M) bson.M{
	update := bson.M{
		"DP1":DP_UP_of_Candle["DP1"],
		"DP2":DP_UP_of_Candle["DP2"],
		"DP3":DP_UP_of_Candle["DP3"],
		"UP1":DP_UP_of_Candle["UP1"],
		"UP2":DP_UP_of_Candle["UP2"],
		"UP3":DP_UP_of_Candle["UP3"],
		"DP1_perc":DP_UP_of_Candleperc["DP1"],
		"DP2_perc":DP_UP_of_Candleperc["DP2"],
		"DP3_perc":DP_UP_of_Candleperc["DP3"],
		"UP1_perc":DP_UP_of_Candleperc["UP1"],
		"UP2_perc":DP_UP_of_Candleperc["UP2"],
		"UP3_perc":DP_UP_of_Candleperc["UP3"],
	}

	dp1 , _  := helpers.ToFloat64(DP_UP_of_Candleperc["DP1"])
	dp2 , _  := helpers.ToFloat64(DP_UP_of_Candleperc["DP2"])
	dp3 , _  := helpers.ToFloat64(DP_UP_of_Candleperc["DP3"])
	update["DP3_DP1_perc"] = dp3 - dp1
	update["DP2_DP1_perc"] = dp2 - dp1
	update["candel_trends"] = candel_trends[0]["candel_trends"]
	update["new_go_indicators"] = 1
 
	return update

}

func calculateHourDifference(startDateString, endDateString string) (int, time.Time, error) {
	// Define the date and time layout
	layout := "2006-01-02 15:04:05"

	// Parse start_date and end_date
	startDate, err := time.Parse(layout, startDateString)
	if err != nil {
		return 0, time.Time{}, err
	}

	endDate, err := time.Parse(layout, endDateString)
	if err != nil {
		return 0, time.Time{}, err
	}

	// Set both dates to the start and end of the respective hours
	startDate = startDate.Truncate(time.Hour)
	endDate = endDate.Truncate(time.Hour).Add(time.Hour - time.Second)

	// Calculate the hour difference
	hourDifference := int(endDate.Sub(startDate).Hours())

	return hourDifference, startDate, nil
}

func CalculateDPUPOfCandle(coin string, date time.Time, open, close float64) map[string]float64 {
	// 0.0000782 ,   0.0000778 ,
	lastCandle2, _ := helpers.GetLastCandle(coin, date, 2)
	lastCandle3, _ := helpers.GetLastCandle(coin, date, 3)
	lastCandle5, _ := helpers.GetLastCandle(coin, date, 5)
	lastCandle6, _ := helpers.GetLastCandle(coin, date, 6)
	lastCandle8, _ := helpers.GetLastCandle(coin, date, 8)
	lastCandle9, _ := helpers.GetLastCandle(coin, date, 9)

	candleMovePerc1 := 0.0
	candleMovePerc2 := 0.0
	candleMovePerc3 := 0.0
	diff1 := 0.0
	diff2 := 0.0
	diff3 := 0.0
	DP1 := 0.0
	DP2 := 0.0
	DP3 := 0.0
	UP1 := 0.0
	UP2 := 0.0
	UP3 := 0.0

	if len(lastCandle3) > 0 {
		candleMovePerc1 = ((close - lastCandle3[0]["close"].(float64)) / close) * 100
	}

	if len(lastCandle6) > 0 {
		candleMovePerc2 = ((close - lastCandle6[0]["close"].(float64)) / close) * 100
	}

	if len(lastCandle9) > 0 {
		candleMovePerc3 = ((close - lastCandle9[0]["close"].(float64)) / close) * 100
	}

	if len(lastCandle2) > 0 {
		diff1 = close - lastCandle2[0]["open"].(float64)
	}

	if len(lastCandle5) > 0 {
		diff2 = close - lastCandle5[0]["open"].(float64)
	}

	if len(lastCandle8) > 0 {
		diff3 = close - lastCandle8[0]["open"].(float64)
	}

	if diff1 < 0 {
		DP1 = candleMovePerc1
	}

	if diff2 < 0 {
		DP2 = candleMovePerc2
	}

	if diff3 < 0 {
		DP3 = candleMovePerc3
	}

	if diff1 > 0 {
		UP1 = candleMovePerc1
	}

	if diff2 > 0 {
		UP2 = candleMovePerc2
	}

	if diff3 > 0 {
		UP3 = candleMovePerc3
	}

	returnData := make(map[string]float64)
	returnData["DP1"] = abs(DP1)
	returnData["DP2"] = abs(DP2)
	returnData["DP3"] = abs(DP3)
	returnData["UP1"] = abs(UP1)
	returnData["UP2"] = abs(UP2)
	returnData["UP3"] = abs(UP3)
	//fmt.Println("returnData",returnData)
	return returnData
}

func getDPUPPercentiles(arr []float64, index string) map[string]float64 {
	percArray := []int{1, 2, 3, 4, 5, 10, 15, 20, 25, 50, 75, 100}
	objToReturn := make(map[string]float64)

	for i := 0; i < 12; i++ {
		val := percArray[i]
		sellVal := int((float64(len(arr)) / 100.0) * float64(val))
		valueToAssign := 0.0
		if sellVal < len(arr) {
			valueToAssign = arr[sellVal]
		}
		indexName := fmt.Sprintf("%s_%d", index, val)
		objToReturn[indexName] = valueToAssign
	}
	if index == "DP1" {
		fmt.Println("get_DP_UP_percentiles for index : "+index, objToReturn)
	}

	return objToReturn
}

// Helper function to check if a field name is DP field
func isDPField(fieldName string) bool {
	return len(fieldName) >= 2 && fieldName[:2] == "DP"
}

// Helper function to check if a field name is UP field
func isUPField(fieldName string) bool {
	return len(fieldName) >= 2 && fieldName[:2] == "UP"
}
func CalculateDPUPPercentiles(coin string, toArr map[string]interface{}, duration int) map[string]string {
	finalArr := make(map[string]string)
	DP1Now := toArr["DP1"].(float64)
	DP2Now := toArr["DP2"].(float64)
	DP3Now := toArr["DP3"].(float64)
	UP1Now := toArr["UP1"].(float64)
	UP2Now := toArr["UP2"].(float64)
	UP3Now := toArr["UP3"].(float64)

	response, err := helpers.DPUPPercentileData(coin, toArr, duration)

	if len(response) == 0 || err != nil {
		fmt.Println("calculateDPUPPercentiles: NO response")
		return nil
	}
	//fmt.Println("response",response)
	dpFields := make(map[string][]float64)
	upFields := make(map[string][]float64)
	firstEntry := response[0]

	//fmt.Println("dp1",dp1)
	// fmt.Println("dp2",dp2)
	// fmt.Println("dp3",dp3)
	// fmt.Println("up1",up1)
	// fmt.Println("up2",up2)
	// fmt.Println("up3",up3)
	// Iterate through the fields of the first map entry
	for fieldName, fieldValue := range firstEntry {
		var parsed []float64
		if dpSlice, ok := fieldValue.(primitive.A); ok {
			if isDPField(fieldName) {
				for _, element := range dpSlice {
					//fmt.Println("field",fieldName)
					//fmt.Println("element",reflect.TypeOf(element))
					value, ok := helpers.ToFloat64(element)
					if ok {
						parsed = append(parsed, value)
					}
				}
				dpFields[fieldName] = parsed
			}
			if isUPField(fieldName) {
				for _, element := range dpSlice {
					value, ok := helpers.ToFloat64(element)
					if ok {
						parsed = append(parsed, value)
					}

				}
				upFields[fieldName] = parsed
			}
		}
	}

	// Now you have maps dpFields and upFields containing values for DP and UP fields
	// Print the values
	// fmt.Println("DP Fields:")
	// fmt.Println(dpFields)

	// fmt.Println("UP Fields:")
	// fmt.Println(upFields)

	DP1 := dpFields["DP1"]
	DP2 := dpFields["DP2"]
	DP3 := dpFields["DP3"]
	UP1 := upFields["UP1"]
	UP2 := upFields["UP2"]
	UP3 := upFields["UP3"]
	// fmt.Println(".==================================================== Before Filters .====================================================")
	// fmt.Println("DP1",DP1)
	// // fmt.Println("DP2",DP2)
	// // fmt.Println("DP3",DP3)
	// // fmt.Println("UP1",UP1)
	// // fmt.Println("UP2",UP2)
	// // fmt.Println("UP3",UP3)
	// fmt.Println(".==================================================== Before Filters Ends .====================================================")

	DP1 = filterNonZero(DP1)
	DP2 = filterNonZero(DP2)
	DP3 = filterNonZero(DP3)
	UP1 = filterNonZero(UP1)
	UP2 = filterNonZero(UP2)
	UP3 = filterNonZero(UP3)

	// fmt.Println(".==================================================== After Filters .====================================================")
	// fmt.Println("DP1",DP1)
	// // fmt.Println("DP2",DP2)
	// // fmt.Println("DP3",DP3)
	// // fmt.Println("UP1",UP1)
	// // fmt.Println("UP2",UP2)
	// // fmt.Println("UP3",UP3)
	// fmt.Println(".==================================================== After Filters Ends .====================================================")

	DPSortSliceDescending(DP1)
	DPSortSliceDescending(DP2)
	DPSortSliceDescending(DP3)
	DPSortSliceDescending(UP1)
	DPSortSliceDescending(UP2)
	DPSortSliceDescending(UP3)

	//fmt.Println(".==================================================== After Sort .====================================================")
	//fmt.Println("DP1",DP1)
	// fmt.Println("DP2",DP2)
	// fmt.Println("DP3",DP3)
	// fmt.Println("UP1",UP1)
	// fmt.Println("UP2",UP2)
	// fmt.Println("UP3",UP3)
	//fmt.Println(".==================================================== After Sort Ends .====================================================")

	DP1Per := getDPUPPercentiles(DP1, "DP1")
	DP2Per := getDPUPPercentiles(DP2, "DP2")
	DP3Per := getDPUPPercentiles(DP3, "DP3")
	UP1Per := getDPUPPercentiles(UP1, "UP1")
	UP2Per := getDPUPPercentiles(UP2, "UP2")
	UP3Per := getDPUPPercentiles(UP3, "UP3")

	currDP1 := getVolumeInDPUPPercentiles(DP1Per, DP1Now, "DP1")
	currDP2 := getVolumeInDPUPPercentiles(DP2Per, DP2Now, "DP2")
	currDP3 := getVolumeInDPUPPercentiles(DP3Per, DP3Now, "DP3")
	currUP1 := getVolumeInDPUPPercentiles(UP1Per, UP1Now, "UP1")
	currUP2 := getVolumeInDPUPPercentiles(UP2Per, UP2Now, "UP2")
	currUP3 := getVolumeInDPUPPercentiles(UP3Per, UP3Now, "UP3")

	finalArr["DP1"] = currDP1
	finalArr["DP2"] = currDP2
	finalArr["DP3"] = currDP3
	finalArr["UP1"] = currUP1
	finalArr["UP2"] = currUP2
	finalArr["UP3"] = currUP3
	fmt.Println("calculate_DP_UP_percentiles final_arr", finalArr)
	return finalArr
}

func getVolumeInDPUPPercentiles(percentileArr map[string]float64, quantity float64, check string) string {
	percentile := make(map[string]float64)
	html := []string{"0", "1", "2", "3", "4", "5", "10", "15", "20", "25", "50", "75", "100"}

	for i := 1; i <= 12; i++ {
		val := html[i]
		getPercentile := percentileArr[check+"_"+val]
		percentile[val] = getPercentile
	}
	//fmt.Println("get_volume_in_DP_UP_percentiles for : "+check,percentile)
	LastQtyTimeAgo := "0"

	if quantity >= percentile["100"] && quantity <= percentile["75"] {
		LastQtyTimeAgo = "100"
	} else if quantity >= percentile["75"] && quantity <= percentile["50"] {
		LastQtyTimeAgo = "75"
	} else if quantity >= percentile["50"] && quantity <= percentile["25"] {
		LastQtyTimeAgo = "50"
	} else if quantity >= percentile["25"] && quantity <= percentile["20"] {
		LastQtyTimeAgo = "25"
	} else if quantity >= percentile["20"] && quantity <= percentile["15"] {
		LastQtyTimeAgo = "20"
	} else if quantity >= percentile["15"] && quantity <= percentile["10"] {
		LastQtyTimeAgo = "15"
	} else if quantity >= percentile["10"] && quantity <= percentile["5"] {
		LastQtyTimeAgo = "10"
	} else if quantity >= percentile["5"] && quantity <= percentile["4"] {
		LastQtyTimeAgo = "5"
	} else if quantity >= percentile["4"] && quantity <= percentile["3"] {
		LastQtyTimeAgo = "4"
	} else if quantity >= percentile["3"] && quantity <= percentile["2"] {
		LastQtyTimeAgo = "3"
	} else if quantity >= percentile["2"] && quantity <= percentile["1"] {
		LastQtyTimeAgo = "2"
	} else if quantity >= percentile["1"] {
		LastQtyTimeAgo = "1"
	}
	//fmt.Println("get_volume_in_DP_UP_percentiles LastQtyTimeAgo :"+check,LastQtyTimeAgo)
	return LastQtyTimeAgo
}

func abs(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}

func filterNonZero(arr []float64) []float64 {
	filtered := make([]float64, 0)
	for _, v := range arr {
		if v != 0 {
			filtered = append(filtered, v)
		}
	}
	return filtered
}

// DPs SortSliceDescending sorts a slice of integers in descending order.
func DPSortSliceDescending(slice []float64) {
	sort.Slice(slice, func(i, j int) bool {
		return slice[i] > slice[j]
	})
}

func processCandelTrends(coin string, dateToGet time.Time, open, close float64) []bson.M {
	lastTrend, err := helpers.LastCandelTrends(coin, dateToGet, "candel_trends")
	if err != nil {
		return []bson.M{}
	}

	var lastTrendValue string = "up"
	if len(lastTrend) > 0 {
		lastTrendValue = lastTrend[0]["candel_trends"].(string)
		//fmt.Println("lastTrend",lastTrend)
	}
	var newTrendValue string
	fmt.Println("lastTrendValue", lastTrendValue)
	switch lastTrendValue {
	case "up":
		trendValueGot := calculateNewTrend("up", coin, dateToGet, close)
		fmt.Println("trend_value_got", trendValueGot)
		if len(trendValueGot) > 0 {
			if trendValueGot["new_trend"].(string) != "" {
				newTrendValue = trendValueGot["new_trend"].(string)
			} else {
				newTrendValue = lastTrendValue
			}
		} else {
			newTrendValue = lastTrendValue
		}
	case "strong_up":
		trendValueGot := calculateNewTrend("strong_up", coin, dateToGet, close)
		if len(trendValueGot) > 0 {
			if trendValueGot["new_trend"].(string) != "" {
				newTrendValue = trendValueGot["new_trend"].(string)
			} else {
				newTrendValue = lastTrendValue
			}
		} else {
			newTrendValue = lastTrendValue
		}
	case "down":
		trendValueGot := calculateNewTrend("down", coin, dateToGet, close)
		if len(trendValueGot) > 0 {
			if trendValueGot["new_trend"].(string) != "" {
				newTrendValue = trendValueGot["new_trend"].(string)
			} else {
				newTrendValue = lastTrendValue
			}
		} else {
			newTrendValue = lastTrendValue
		}
	case "strong_down":
		trendValueGot := calculateNewTrend("strong_down", coin, dateToGet, close)
		if len(trendValueGot) > 0 {
			if trendValueGot["new_trend"].(string) != "" {
				newTrendValue = trendValueGot["new_trend"].(string)
			} else {
				newTrendValue = lastTrendValue
			}
		} else {
			newTrendValue = lastTrendValue
		}
	default:
		newTrendValue = lastTrendValue
	}
	data := []bson.M{
		{"candel_trends": newTrendValue, "last_candel_trends": lastTrendValue},
	}
	return data

}

func calculateNewTrend(lastTrendValue string, coin string, dateToGet time.Time, close float64) map[string]interface{} {
	finalTemp := map[string]interface{}{}
	finalTemp["currentClose"] = close
	newTrendValue := ""
	closestHH := 0

	if lastTrendValue == "up" || lastTrendValue == "strong_up" {

		if lastTrendValue == "up" {
			lastHH, lastHHerr := helpers.GetHHSwingStatusFromCandleChart(coin, dateToGet)
			if lastHHerr != nil {
				fmt.Println("ERROR ON lastHHerr", lastHHerr.Error())
			}
			lastLH, lastLHerr := helpers.GetLHSwingStatusFromCandleChart(coin, dateToGet)
			if lastLHerr != nil {
				fmt.Println("ERROR ON lastHHerr", lastLHerr.Error())
			}
			if len(lastHH) > 0 {
				if len(lastLH) > 0 {
					hhOpenTimeHumanReadable := lastHH[0]["openTime_human_readible"].(string)
					lhOpenTimeHumanReadable := lastLH[0]["openTime_human_readible"].(string)

					hhTime, err := time.Parse(customLayout, hhOpenTimeHumanReadable)
					if err != nil {
						// Handle parsing error
						fmt.Printf("Error parsing hhOpenTimeHumanReadable: %v\n", err)
					} else {
						lhTime, err := time.Parse(customLayout, lhOpenTimeHumanReadable)
						if err != nil {
							// Handle parsing error
							fmt.Printf("Error parsing lhOpenTimeHumanReadable: %v\n", err)
						} else {
							if hhTime.After(lhTime) {
								closestHH = 1
							}
						}
					}
				}

				finalTemp["highest_swing_point"] = lastHH[0]["highest_swing_point"].(float64)

				if close > lastHH[0]["highest_swing_point"].(float64) {
					newTrendValue = "strong_up"
				}
			}
		}

		if newTrendValue != "strong_up" {
			lastLL, lastLLerr := helpers.GetLLSwingStatusFromCandleChart(coin, dateToGet)
			if lastLLerr != nil {
				fmt.Println("ERROR ON lastHHerr", lastLLerr.Error())
			}
			lastHL, lastHLerr := helpers.GetHLSwingStatusFromCandleChart(coin, dateToGet)
			if lastHLerr != nil {
				fmt.Println("ERROR ON lastHHerr", lastHLerr.Error())
			}
			closestLL := 0

			if len(lastLL) > 0 {
				llOpenTimeHumanReadable := lastLL[0]["openTime_human_readible"].(string)

				if len(lastHL) > 0 {
					hlOpenTimeHumanReadable := lastHL[0]["openTime_human_readible"].(string)
					llTime, err := time.Parse(customLayout, llOpenTimeHumanReadable)
					if err != nil {
						// Handle parsing error
						fmt.Printf("Error parsing llOpenTimeHumanReadable: %v\n", err)
					} else {
						hlTime, err := time.Parse(customLayout, hlOpenTimeHumanReadable)
						if err != nil {
							// Handle parsing error
							fmt.Printf("Error parsing hlOpenTimeHumanReadable: %v\n", err)
						} else {
							if llTime.After(hlTime) {
								closestLL = 1
							}
						}
					}

					// if time.Date(llOpenTimeHumanReadable).After(time.Date(hlOpenTimeHumanReadable)) {
					// 	closestLL = 1
					// }
				}
				lowest_swing_point, _ := helpers.ToFloat64(lastLL[0]["lowest_swing_point"])
				finalTemp["lowest_swing_point"] = lowest_swing_point

				if close < lowest_swing_point {
					newTrendValue = "strong_down"
				}
			}

			finalTemp["closestLL"] = closestLL

			if len(lastHL) > 0 && newTrendValue != "strong_down" && closestLL == 0 {

				lowest_swing_point, _ := helpers.ToFloat64(lastHL[0]["lowest_swing_point"])
				finalTemp["lowest_swing_point_latest"] = lowest_swing_point

				if close < lowest_swing_point {
					newTrendValue = "down"
				}
			}
		}
	}

	if lastTrendValue == "down" || lastTrendValue == "strong_down" {
		if lastTrendValue == "down" {
			lastLL, lastLLerr := helpers.GetLLSwingStatusFromCandleChart(coin, dateToGet)

			if lastLLerr != nil {
				fmt.Println("ERROR ON lastHHerr", lastLLerr.Error())
			}
			//lastHL, _ := helpers.GetHLSwingStatusFromCandleChart(coin, dateToGet)

			if len(lastLL) > 0 {
				lowest_swing_point, _ := helpers.ToFloat64(lastLL[0]["lowest_swing_point"])
				finalTemp["lowest_swing_point"] = lowest_swing_point

				if close < lowest_swing_point {
					newTrendValue = "strong_down"
				}
			}
		}

		if newTrendValue != "strong_down" {

			lastHH, lastHHerr := helpers.GetHHSwingStatusFromCandleChart(coin, dateToGet)
			if lastHHerr != nil {
				fmt.Println("ERROR ON lastHHerr", lastHHerr.Error())
			}

			lastLH, lastLHerr := helpers.GetLHSwingStatusFromCandleChart(coin, dateToGet)

			if lastLHerr != nil {
				fmt.Println("ERROR ON lastHHerr", lastLHerr.Error())
			}
			if len(lastHH) > 0 {
				hhOpenTimeHumanReadable := lastHH[0]["openTime_human_readible"].(string)

				if len(lastLH) > 0 {
					lhOpenTimeHumanReadable := lastLH[0]["openTime_human_readible"].(string)
					hhTime, err := time.Parse(customLayout, hhOpenTimeHumanReadable)
					if err != nil {
						// Handle parsing error
						fmt.Printf("Error parsing hhOpenTimeHumanReadable: %v\n", err)
					} else {
						lhTime, err := time.Parse(customLayout, lhOpenTimeHumanReadable)
						if err != nil {
							// Handle parsing error
							fmt.Printf("Error parsing lhOpenTimeHumanReadable: %v\n", err)
						} else {
							if hhTime.After(lhTime) {
								closestHH = 1
							}
						}
					}

				}

				highest_swing_point, _ := helpers.ToFloat64(lastHH[0]["highest_swing_point"])

				finalTemp["highest_swing_point"] = highest_swing_point

				if close > highest_swing_point {
					newTrendValue = "strong_up"
				}
			}

			finalTemp["closestHH"] = closestHH

			if len(lastLH) > 0 && newTrendValue != "strong_up" && closestHH == 0 {
				highest_swing_point, _ := helpers.ToFloat64(lastLH[0]["highest_swing_point"])
				finalTemp["highest_swing_point_latest"] = highest_swing_point

				if close > highest_swing_point {
					newTrendValue = "up"
				}
			}
		}
	}

	finalTemp["new_trend"] = newTrendValue
	return finalTemp
}

func dpupTrendDirectionCalculations(coin string, startDate, endDate time.Time) (map[string]interface{}, error) {
	var getTrend map[string]interface{}
	getTrend = make(map[string]interface{})

	var limit int64 = 1
	candleData, err := helpers.GetDailyCandleData(coin, startDate, endDate, limit)
	if err != nil {
		fmt.Println("dpupTrendDirectionCalculations GetDailyCandleData Error", err.Error())
		return nil, err
	}

	if len(candleData) == 0 {
		fmt.Println("dpupTrendDirectionCalculations GetDailyCandleData is Empty")
		return getTrend, fmt.Errorf("dpupTrendDirectionCalculations GetDailyCandleData is Empty")
	}
	currentCandle := candleData[0]
	//fmt.Println("candleData",candleData)
	// Access individual values by key
	keysToCheck := []string{"DP1_perc", "DP2_perc", "DP3_perc", "UP1_perc", "UP2_perc", "UP3_perc"}
	noErrors := true
	values := make(map[string]float64)

	for _, key := range keysToCheck {
		if val, ok := currentCandle[key]; ok {
			if strVal, isString := val.(string); isString {
				//fmt.Println("strVal",strVal)
				//fmt.Println("key",key)

				if num, err := helpers.ConvertStrNumbersToFloat(strVal); err == nil {
					values[key] = num
				} else {
					//fmt.Println("errerr",err)
					noErrors = false
					break
				}
			}
		} else {
			noErrors = false
			break
		}
	}
	if !noErrors {
		fmt.Println("DP or UP perc values are not present or not in string format")
		return nil, fmt.Errorf("DP or UP perc values are not present or not in string format")
	}
	var dpup_trend_direction string = "up"     //default values
	var new_dpup_trend_direction string = "up" // default values
	value, exists := currentCandle["dpup_trend_direction"]
	if exists {
		_, ok := value.(string)
		if ok {
			dpup_trend_direction = value.(string)
		}
	}
	if dpup_trend_direction == "up" {
		if (values["DP1_perc"] > 0 && values["DP1_perc"] <= 90) && (values["DP2_perc"] > 0 && values["DP2_perc"] <= 90) && (values["DP3_perc"] > 0 && values["DP3_perc"] <= 90) {
			new_dpup_trend_direction = "down"
		} else if (values["DP1_perc"] > 0 && values["DP1_perc"] <= 90) && (values["DP3_perc"]) == 0 || values["DP3_perc"] > 80 {
			new_dpup_trend_direction = "down"
		}
	}
	if dpup_trend_direction == "down" {
		if (values["UP1_perc"] > 0 && values["UP1_perc"] <= 90) && (values["UP2_perc"] > 0 && values["UP2_perc"] <= 90) && (values["UP3_perc"] > 0 && values["UP3_perc"] <= 90) {
			new_dpup_trend_direction = "up"
		} else if (values["UP1_perc"] > 0 && values["UP1_perc"] <= 90) && (values["UP3_perc"]) == 0 || values["UP3_perc"] > 80 {
			new_dpup_trend_direction = "up"
		}
	}
	filters := bson.M{
		"_id": currentCandle["_id"].(primitive.ObjectID),
	}
	update := bson.M{
		"$set": bson.M{
			"dpup_trend_direction": new_dpup_trend_direction,
		},
	}
	err = helpers.UpdateDailyData(filters, update)
	getTrend["dpup_trend_direction"] = new_dpup_trend_direction

	// update method here
	return getTrend, nil
}

func calculateDailyTrend(coin string, startDate, endDate time.Time) (map[string]interface{}, error) {
	fmt.Println("inside Daily Trend")

	var getTrend map[string]interface{}
	var limit int64 = 2
	getTrend = make(map[string]interface{})

	// we need current and latest previous , i..e, two candles
	// adjust the start Date
	startDate = startDate.AddDate(0, 0, -1)
	candlesData, err := helpers.GetDailyCandleData(coin, startDate, endDate, limit)
	if err != nil {
		fmt.Println("calculateDailyTrend candlesData Error", err.Error())
		return nil, err
	}

	if len(candlesData) == 0 || len(candlesData) < 2 {
		fmt.Println("calculateDailyTrend candlesData is Empty or has Lenght less than two")
		return getTrend, fmt.Errorf("calculateDailyTrend candlesData is Empty or has Lenght less than two")
	}

	// process the candles data accordingly.... sort by created date desc applied.
	currentCandle := candlesData[0]  // first index will be current
	previousCandle := candlesData[1] // second will be the previous
	var (
		current_daily_trend string
		new_trend           string
		currentcandle_color string = "red"
		candle_color        string = "red"
	)
	trend_value, exists := currentCandle["daily_trend"]
	if exists {
		_, ok := trend_value.(string)
		if ok {
			current_daily_trend = trend_value.(string)
		}
	}
	currentClose := currentCandle["close"].(float64)
	currentOpen := currentCandle["open"].(float64)
	if currentClose > currentOpen {
		currentcandle_color = "green"
	}
	previousClose := previousCandle["close"].(float64)
	previousOpen := previousCandle["open"].(float64)
	previousLow := previousCandle["low"].(float64)
	previousHigh := previousCandle["high"].(float64)
	if previousClose > previousOpen {
		candle_color = "green"
	}

	if current_daily_trend == "" || current_daily_trend == "up" {
		if candle_color == "red" {
			if currentClose < previousClose {
				new_trend = "down"
			} else {
				new_trend = "up"
			}
		} else if candle_color == "green" {
			if currentcandle_color == "red" {
				if currentClose < previousLow {
					new_trend = "down"
				} else {
					new_trend = "up"
				}
			}
		}
	} else if current_daily_trend == "down" {
		if candle_color == "green" {
			if currentClose > previousClose {
				new_trend = "up"
			}
		}
		if candle_color == "red" {
			if currentClose > previousHigh {
				new_trend = "up"
			}
		}
	}

	if new_trend == "" {
		if current_daily_trend == "" {
			new_trend = "up"
		} else {
			new_trend = current_daily_trend
		}
	}

	filters := bson.M{
		"_id": currentCandle["_id"].(primitive.ObjectID),
	}
	update := bson.M{
		"$set": bson.M{
			"daily_trend": new_trend,
		},
	}
	err = helpers.UpdateDailyData(filters, update)
	//getTrend["_id"] = currentCandle["_id"].(primitive.ObjectID)
	getTrend["daily_trend"] = new_trend
	//getTrend["openTime_human_readible"] = 	currentCandle["openTime_human_readible"].(string)

	return getTrend, nil

}

func ProcessFinalData(data []interface{}, typeValue int) {

	//collectionName := "market_chart"
	fmt.Println("data", data)
	// for _, insertdata := range data {
	// 	postUpdate := bson.M{}
	// 	postSearchCriteria := bson.M{}
	// 	keysToUpdate := []string{"DP1", "DP2", "DP3", "UP1", "UP2", "UP3", "DP1_perc", "DP2_perc", "DP3_perc", "UP1_perc", "UP2_perc", "UP3_perc"}

	// 	if typeValue == 1 {

	// 		for _, key := range keysToUpdate {
	// 			keyTitle := key+"_btc"
	// 			if dataMap, ok := insertdata.(map[string]interface{}); ok {
	// 				if dpUp, ok := dataMap["DP_UP_of_Candle"].(map[string]interface{}); ok {
	// 					value := dpUp
	// 					postUpdate[keyTitle] = value
	// 				}
	// 			}
	// 		}

	// 		postSearchCriteria["openTime_human_readible"] = insertdata["openTime_human_readible"]
	// 		postSearchCriteria["coin"] = bson.M{"$nin": []string{"NCASHBTC", "BTCUSDT"}}
	// 	} else {
	// 		for _, key := range keysToUpdate {
	// 			postUpdate[key] = insertdata["DP_UP_of_Candle"].(map[string]interface{})[key]
	// 		}

	// 		postSearchCriteria["_id"] = insertdata["_id"]
	// 	}
	// 	_ = mongohelpers.MongoUpdateOne(collectionName,postSearchCriteria,postUpdate,false)
	// }
	return
}

func diffHours(start, end time.Time) int {
	diff := end.Sub(start)
	return int(diff.Hours())
}
