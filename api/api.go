package api

import (
	"fmt"
	"time"
	"sort"
	"io/ioutil"
	"net/http"
	"encoding/json"
	"indicatorsAPP/helpers"
	//"indicatorsAPP/mongohelpers"
	 "github.com/gin-gonic/gin"
	 "go.mongodb.org/mongo-driver/bson"
	 "go.mongodb.org/mongo-driver/bson/primitive"
)
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
	fmt.Println("jsonData",jsonData)
	fmt.Println(len(jsonData))
	if  len(jsonData) != 0 {
		if coin := jsonData["coin"].(string); coin == "" {
			errors = append(errors, "Coin is Required Field")
		}
		customLayout := "2006-01-02 15:04:05"
		startDateStr := jsonData["start_date"].(string)  //c.Query("start_date")
		startDate, err := time.Parse(customLayout, startDateStr)
		if err != nil {
			errors = append(errors, "Start Date has Invalid Format")
		}

		endDateStr :=  jsonData["end_date"].(string) //c.Query("end_date")
		var diffInDays int = 0
		if endDateStr != "" {
			endDate, err := time.Parse(customLayout, endDateStr)
			if err != nil {
				errors = append(errors, "End Date has Invalid Format")
			}

			diffInDays = int(endDate.Sub(startDate).Hours() / 24)
		} 
		fmt.Println("errors",errors)
		if len(errors) > 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  400,
				"data":    nil, // Replace with your data
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
			dpupDirection , _ := dpupTrendDirectionCalculations(coin,fromDate,toDate)
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
			"error":   nil, // Handle errors if necessary
			"message": "Success",
		})
		return
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  400,
			"data":    nil, // Replace with your data
			"error":   "Params Missing", // Handle errors if necessary
			"message": "Error On Your Request",
		})
		return
		
	}
}
// POST METHOD to Set Hourly Indicators
func setHourlyIndicators(c *gin.Context) {
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
		//var diffInDays int = 0
		if coin, ok := requestBody["coin"].(string); !ok || coin == "" {
			errors = append(errors, "Coin is a required field")
		}
		if startDateStr, ok := requestBody["start_date"].(string); !ok || startDateStr == "" {
			errors = append(errors, "Start Date is a required field")
		} else {
			_, err := time.Parse(time.RFC3339, startDateStr)
			if err != nil {
				errors = append(errors, "Start Date has an invalid format")
			}
			if endDateStr, ok := requestBody["end_date"].(string); ok && endDateStr != "" {
				_, err = time.Parse(time.RFC3339, endDateStr)
				if err != nil {
					errors = append(errors, "End Date has an invalid format")
				}
			}
		}

		if len(errors) == 0 {
			where := bson.M{}
			coin, _ := requestBody["coin"].(string)
			startDateStr, _ := requestBody["start_date"].(string)
			endDateStr, _ := requestBody["end_date"].(string)

			startDate, _ := time.Parse(time.RFC3339, startDateStr)
			where["coin"] = coin

			if startDateStr != "" && endDateStr != "" {
				endDate, _ := time.Parse(time.RFC3339, endDateStr)
				where["created_date"] = bson.M{"$gte": startDate, "$lte": endDate}
			} else {
				where["created_date"] = bson.M{"$gte": startDate}
			}

			var resArr []interface{}
			chartDataFetched, _ :=  helpers.MarketChartDataForCoin(where)
			dataToParse := []interface{}{chartDataFetched}

			if len(dataToParse) == 0 {
				resArr = append(resArr, "NO Data Found")
			} else {
				settings := dataToParse[0].([]interface{})

				for _, current := range settings {
					temp := map[string]interface{}{}
					coinSymbol := coin
					open := current.(map[string]interface{})["open"].(float64)
					_id := current.(map[string]interface{})["_id"].(string)
					//high := current.(map[string]interface{})["high"].(float64)
					//low := current.(map[string]interface{})["low"].(float64)
					close := current.(map[string]interface{})["close"].(float64)
					openTimeHumanReadable := current.(map[string]interface{})["openTime_human_readible"].(string)
					letDate, _ := time.Parse(time.RFC3339, openTimeHumanReadable)
					startDateCandle := time.Date(letDate.Year(), letDate.Month(), letDate.Day(), letDate.Hour(), 0, 0, 0, time.UTC)
					endDateCandle := time.Date(letDate.Year(), letDate.Month(), letDate.Day(), letDate.Hour(), 59, 59, 0, time.UTC)

					DP_UP_of_Candle := CalculateDPUPOfCandle(coinSymbol, startDateCandle, open, close)

					toarr := map[string]interface{}{
						"DP1":             DP_UP_of_Candle["DP1"],
						"DP2":             DP_UP_of_Candle["DP2"],
						"DP3":             DP_UP_of_Candle["DP3"],
						"UP1":             DP_UP_of_Candle["UP1"],
						"UP2":             DP_UP_of_Candle["UP2"],
						"UP3":             DP_UP_of_Candle["UP3"],
						"startCandleDate": startDateCandle,
						"endCandleDate":   endDateCandle,
					}

					DP_UP_of_CandlePerc := CalculateDPUPPercentiles(coinSymbol, toarr, 30)

					temp["_id"] = _id
					temp["openTime_human_readible"] = openTimeHumanReadable
					temp["coinsymbol"] = coinSymbol
					temp["DP_UP_of_Candleperc"] = DP_UP_of_CandlePerc
					temp["DP_UP_of_Candle"] = DP_UP_of_Candle

					
					//candleTrends := CalculateCandleTrends(coinSymbol, close, startDateCandle, previousDateCandle)
					//temp["candel_trends"] = candleTrends

					resArr = append(resArr, temp)
				}
				var processType int = 0
				ProcessFinalData(resArr, processType)
				var candelTrends interface{}
				if coin == "BTCUSDT" {
					processType = 1
					ProcessFinalData(resArr, processType)
				}

				//candelTrends := processCandelTrends(coin, startDateStr, totalHours, 0)

				c.JSON(http.StatusOK, gin.H{
					"status": 200,
					"data":   requestBody,
					"response": map[string]interface{}{
						"dpup":        resArr,
						"candel_trends": candelTrends,
					},
					"error":   nil,
					"message": "Successfully Done",
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
		}
	}

	c.JSON(http.StatusBadRequest, gin.H{
		"status":  400,
		"data":    nil,
		"error":   "Params Missing",
		"message": "Error On Your Request",
	})
}



func CalculateDPUPOfCandle(coin string, date time.Time, open, close float64) map[string]float64 {
	lastCandle2 , _ := helpers.GetLastCandle(coin, date, 2)
	lastCandle3 , _ := helpers.GetLastCandle(coin, date, 3)
	lastCandle5 , _ := helpers.GetLastCandle(coin, date, 5)
	lastCandle6 , _  := helpers.GetLastCandle(coin, date, 6)
	lastCandle8 , _ := helpers.GetLastCandle(coin, date, 8)
	lastCandle9 , _ := helpers.GetLastCandle(coin, date, 9)

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

	return objToReturn
}

func CalculateDPUPPercentiles(coin string, toArr map[string]interface{}, duration int) map[string]string {
	finalArr := make(map[string]string)
	DP1Now := toArr["DP1"].(float64)
	DP2Now := toArr["DP2"].(float64)
	DP3Now := toArr["DP3"].(float64)
	UP1Now := toArr["UP1"].(float64)
	UP2Now := toArr["UP2"].(float64)
	UP3Now := toArr["UP3"].(float64)

	response , err := helpers.DPUPPercentileData(coin, toArr, duration)

	if len(response) == 0 || err!=nil {
		fmt.Println("calculateDPUPPercentiles: NO response")
		return nil
	}

	DP1 := response[0]["DP1"].([]float64)
	DP2 := response[0]["DP2"].([]float64)
	DP3 := response[0]["DP3"].([]float64)
	UP1 := response[0]["UP1"].([]float64)
	UP2 := response[0]["UP2"].([]float64)
	UP3 := response[0]["UP3"].([]float64)

	DP1 = filterNonZero(DP1)
	DP2 = filterNonZero(DP2)
	DP3 = filterNonZero(DP3)
	UP1 = filterNonZero(UP1)
	UP2 = filterNonZero(UP2)
	UP3 = filterNonZero(UP3)


	sort.Float64s(DP1)
	sort.Float64s(DP2)
	sort.Float64s(DP3)
	sort.Float64s(UP1)
	sort.Float64s(DP1)
	sort.Float64s(UP2)
	sort.Float64s(UP3)

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

func processCandelTrends(coin string, date1 time.Time, totalHours int, startOver int) []bson.M{
	var final []bson.M
	var	res []bson.M
	//lastTrend := map[string]string{}

	for i := 0; i < totalHours; i++ {
		//trendValueGot := map[string]string{}
		//newTrendValue := ""
		//lastTrendValue := ""
		makeDate := date1.Add(time.Duration(i) * time.Hour)
		date := makeDate.Truncate(time.Hour)
		currentData, err := helpers.GetLastCandle(coin, date, 0) // last candle by coin, date, and third param subtraction of given num
		//lastDateToGet := date.Add(-1 * time.Hour)

		if len(currentData) == 0 || err!=nil {
			continue
		}

		
		for _, data := range currentData {
			final = append(final, data)
		}
	}

	if len(final) == 0 {
		return []bson.M{}
	}

	
	counter := 0

	for index := 0; index < len(final); index++ {
		counter++
		dataToUpDate := final[index]
		var current map[string]interface{}
		for _, value := range dataToUpDate {
			if v, ok := value.(map[string]interface{}); ok {
				current = v
				break // Exit the loop after finding the first map
			}
		}
		newTrendValue := ""
		temp333 := map[string]interface{}{}
		currentOpen := current["openTime_human_readible"].(string)
		fmt.Println("currentOpen", currentOpen)
		temp333["currentOpen"] = currentOpen

		if counter == 1 && startOver == 1 {
			temp333["lastdatetime"] = nil
			temp333["last_candel_trends"] = nil
			temp333["date"] = nil
			newTrendValue = "up"
		} else {

			makeDate, err := time.Parse(time.RFC3339, currentOpen)
			if err != nil {
				continue
			}
			date := time.Date(makeDate.Year(), makeDate.Month(), makeDate.Day(), makeDate.Hour(), makeDate.Minute(), makeDate.Second(), 0, time.UTC)
			dateToGet := time.Date(makeDate.Year(), makeDate.Month(), makeDate.Day(), makeDate.Hour()-1, makeDate.Minute(), makeDate.Second(), 0, time.UTC)
			temp333["date"] = date
			temp333["makeDate"] = makeDate
			lastData, _ := helpers.LastCandelTrends(coin, current["created_date"].(time.Time), "candel_trends")
			fmt.Println("lastData", lastData, date)
			lastTrendValue := lastData[0]["candel_trends"].(string)
			temp333["last_candel_trends"] = lastTrendValue
			temp333["dsds"] = lastData[0]["_id"].(string)

			switch lastTrendValue {
			case "up":
				trendValueGot := calculateNewTrend("up", coin, dateToGet, current)
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
				trendValueGot := calculateNewTrend("strong_up", coin, dateToGet, current)
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
				trendValueGot := calculateNewTrend("down", coin, dateToGet, current)
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
				trendValueGot := calculateNewTrend("strong_down", coin, dateToGet, current)
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
		}

		//idToUpdate := current["_id"].(string)
		//candelTrends := newTrendValue
		temp333["candel_trends"] = newTrendValue

		//updateCandelTrends(candelTrends, idToUpdate)
		res = append(res, temp333)
	}

	return res
}

func calculateNewTrend(lastTrendValue string, coin string, dateToGet time.Time, currentData map[string]interface{}) map[string]interface{} {
	finalTemp := map[string]interface{}{}
	finalTemp["currentClose"] = currentData["close"].(float64)
	newTrendValue := ""
	closestHH := 0
	if lastTrendValue == "up" || lastTrendValue == "strong_up" {
		
		if lastTrendValue == "up" {
			lastHH, _ := helpers.GetHHSwingStatusFromCandleChart(coin, dateToGet)
			lastLH, _ := helpers.GetLHSwingStatusFromCandleChart(coin, dateToGet)

			if len(lastHH) > 0 {
				if len(lastLH) > 0 {
					hhOpenTimeHumanReadable := lastHH[0]["openTime_human_readible"].(string)
					lhOpenTimeHumanReadable := lastLH[0]["openTime_human_readible"].(string)
			
					hhTime, err := time.Parse(time.RFC3339, hhOpenTimeHumanReadable)
					if err != nil {
						// Handle parsing error
						fmt.Printf("Error parsing hhOpenTimeHumanReadable: %v\n", err)
					} else {
						lhTime, err := time.Parse(time.RFC3339, lhOpenTimeHumanReadable)
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

				if currentData["close"].(float64) > lastHH[0]["highest_swing_point"].(float64) {
					newTrendValue = "strong_up"
				}
			}
		}

		if newTrendValue != "strong_up" {
			lastLL, _ := helpers.GetLLSwingStatusFromCandleChart(coin, dateToGet)
			lastHL, _ := helpers.GetHLSwingStatusFromCandleChart(coin, dateToGet)
			closestLL := 0

			if len(lastLL) > 0 {
				llOpenTimeHumanReadable := lastLL[0]["openTime_human_readible"].(string)

				if len(lastHL) > 0 {
					hlOpenTimeHumanReadable := lastHL[0]["openTime_human_readible"].(string)
					llTime, err := time.Parse(time.RFC3339, llOpenTimeHumanReadable)
					if err != nil {
						// Handle parsing error
						fmt.Printf("Error parsing llOpenTimeHumanReadable: %v\n", err)
					} else {
						hlTime, err := time.Parse(time.RFC3339, hlOpenTimeHumanReadable)
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

				finalTemp["lowest_swing_point"] = lastLL[0]["lowest_swing_point"].(float64)

				if currentData["close"].(float64) < lastLL[0]["lowest_swing_point"].(float64) {
					newTrendValue = "strong_down"
				}
			}

			finalTemp["closestLL"] = closestLL

			if len(lastHL) > 0 && newTrendValue != "strong_down" && closestLL == 0 {
				finalTemp["lowest_swing_point_latest"] = lastHL[0]["lowest_swing_point"].(float64)

				if currentData["close"].(float64) < lastHL[0]["lowest_swing_point"].(float64) {
					newTrendValue = "down"
				}
			}
		}
	}

	if lastTrendValue == "down" || lastTrendValue == "strong_down" {
		if lastTrendValue == "down" {
			lastLL, _ := helpers.GetLLSwingStatusFromCandleChart(coin, dateToGet)
			//lastHL, _ := helpers.GetHLSwingStatusFromCandleChart(coin, dateToGet)

			if len(lastLL) > 0 {
				finalTemp["lowest_swing_point"] = lastLL[0]["lowest_swing_point"].(float64)

				if currentData["close"].(float64) < lastLL[0]["lowest_swing_point"].(float64) {
					newTrendValue = "strong_down"
				}
			}
		}

		if newTrendValue != "strong_down" {
			
			lastHH, _ := helpers.GetHHSwingStatusFromCandleChart(coin, dateToGet)
			lastLH, _ := helpers.GetLHSwingStatusFromCandleChart(coin, dateToGet)

			if len(lastHH) > 0 {
				hhOpenTimeHumanReadable := lastHH[0]["openTime_human_readible"].(string)

				if len(lastLH) > 0 {
					lhOpenTimeHumanReadable := lastLH[0]["openTime_human_readible"].(string)
					hhTime, err := time.Parse(time.RFC3339, hhOpenTimeHumanReadable)
					if err != nil {
						// Handle parsing error
						fmt.Printf("Error parsing hhOpenTimeHumanReadable: %v\n", err)
					} else {
						lhTime, err := time.Parse(time.RFC3339, lhOpenTimeHumanReadable)
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

				if currentData["close"].(float64) > lastHH[0]["highest_swing_point"].(float64) {
					newTrendValue = "strong_up"
				}
			}

			finalTemp["closestHH"] = closestHH

			if len(lastLH) > 0 && newTrendValue != "strong_up" && closestHH == 0 {
				finalTemp["highest_swing_point_latest"] = lastLH[0]["highest_swing_point"].(float64)

				if currentData["close"].(float64) > lastLH[0]["highest_swing_point"].(float64) {
					newTrendValue = "up"
				}
			}
		}
	}

	finalTemp["new_trend"] = newTrendValue
	return finalTemp
}


func dpupTrendDirectionCalculations(coin string, startDate , endDate  time.Time) (map[string]interface{}, error) {
	var getTrend map[string]interface{}
	getTrend = make(map[string]interface{})

	var limit int64 = 1
	candleData ,err := helpers.GetDailyCandleData(coin,startDate,endDate,limit)
	if err != nil {
		fmt.Println("dpupTrendDirectionCalculations GetDailyCandleData Error",err.Error())
		return nil , err
	}

	if len(candleData) == 0 {
		fmt.Println("dpupTrendDirectionCalculations GetDailyCandleData is Empty")
		return getTrend,  fmt.Errorf("dpupTrendDirectionCalculations GetDailyCandleData is Empty")
	}
	currentCandle := candleData[0]
	//fmt.Println("candleData",candleData)
	// Access individual values by key
	keysToCheck := []string{"DP1_perc", "DP2_perc", "DP3_perc","UP1_perc", "UP2_perc", "UP3_perc"}
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
	if !noErrors{
		fmt.Println("DP or UP perc values are not present or not in string format")
		return nil , fmt.Errorf("DP or UP perc values are not present or not in string format")
	}
	var dpup_trend_direction string = "up"  //default values
	var new_dpup_trend_direction string = "up" // default values
	value, exists := currentCandle["dpup_trend_direction"]
	if exists{
		_ , ok := value.(string)
		if ok{
			dpup_trend_direction = value.(string)
		}
	}
	if dpup_trend_direction == "up"{
		if ((values["DP1_perc"] > 0 && values["DP1_perc"] <= 90) && (values["DP2_perc"] > 0 && values["DP2_perc"] <= 90) && (values["DP3_perc"] > 0 && values["DP3_perc"] <= 90)){
			new_dpup_trend_direction = "down"
		}  else if ((values["DP1_perc"] > 0 && values["DP1_perc"] <= 90) && (values["DP3_perc"]) == 0 || values["DP3_perc"] > 80) {
			new_dpup_trend_direction = "down"
		}
	}
	if dpup_trend_direction == "down"{
		if ((values["UP1_perc"] > 0 && values["UP1_perc"] <= 90) && (values["UP2_perc"] > 0 && values["UP2_perc"] <= 90) && (values["UP3_perc"] > 0 && values["UP3_perc"] <= 90)){
			new_dpup_trend_direction = "up"
		}  else if ((values["UP1_perc"] > 0 && values["UP1_perc"] <= 90) && (values["UP3_perc"]) == 0 || values["UP3_perc"] > 80) {
			new_dpup_trend_direction = "up"
		}
	}
	filters := bson.M{
		"_id":currentCandle["_id"].(primitive.ObjectID),
	}
	update := bson.M{
		"$set":bson.M{
			"dpup_trend_direction":new_dpup_trend_direction,
		},
	}
	err = helpers.UpdateDailyData(filters, update) 
	getTrend["dpup_trend_direction"] = 	new_dpup_trend_direction
	
	// update method here
	return getTrend , nil
}


func calculateDailyTrend(coin string, startDate , endDate time.Time) (map[string]interface{}, error) {
	fmt.Println("inside Daily Trend")

	var getTrend map[string]interface{}
	var limit int64 = 2
	getTrend = make(map[string]interface{})

	// we need current and latest previous , i..e, two candles
	// adjust the start Date
	startDate = startDate.AddDate(0, 0, -1)
	candlesData ,err := helpers.GetDailyCandleData(coin,startDate,endDate,limit)
	if err != nil {
		fmt.Println("calculateDailyTrend candlesData Error",err.Error())
		return nil , err
	}

	if len(candlesData) == 0 || len(candlesData) < 2 {
		fmt.Println("calculateDailyTrend candlesData is Empty or has Lenght less than two")
		return getTrend, fmt.Errorf("calculateDailyTrend candlesData is Empty or has Lenght less than two")
	}

	// process the candles data accordingly.... sort by created date desc applied. 
	currentCandle  := candlesData[0]  // first index will be current
	previousCandle := candlesData[1] // second will be the previous 
	var (current_daily_trend string  
		new_trend string 		
		currentcandle_color string = "red" 
		candle_color string = "red")
	trend_value , exists := currentCandle["daily_trend"]
	if exists{
		_ , ok := trend_value.(string)
		if ok {
			current_daily_trend = trend_value.(string)
		}
	}
	currentClose := currentCandle["close"].(float64)
	currentOpen  := currentCandle["open"].(float64)
	if currentClose > currentOpen {
		currentcandle_color = "green"
	}
	previousClose := previousCandle["close"].(float64)
	previousOpen  := previousCandle["open"].(float64)
	previousLow := previousCandle["low"].(float64)
	previousHigh := previousCandle["high"].(float64)
	if previousClose > previousOpen {
		candle_color = "green"
	}

	if (current_daily_trend == "" || current_daily_trend == "up"){
		if candle_color == "red"{
			if currentClose < previousClose {
				new_trend = "down"
			} else {
				new_trend = "up"
			}	
		} else if candle_color == "green"{
			if currentcandle_color == "red"{
				if currentClose < previousLow {
					new_trend = "down"
				} else {
					new_trend = "up"
				}
			}
		}
	} else if current_daily_trend == "down" {
		if candle_color == "green"{
			if currentClose > previousClose {
				new_trend = "up"
			}
		}
		if candle_color == "red"{
			if currentClose > previousHigh {
				new_trend = "up"
			}
		}
	}
	
	if new_trend == ""{
		if current_daily_trend == ""{
			new_trend = "up"
		} else {
			new_trend = current_daily_trend
		}
	}

	filters := bson.M{
		"_id":currentCandle["_id"].(primitive.ObjectID),
	}
	update := bson.M{
		"$set":bson.M{
			"daily_trend":new_trend,
		},
	}
	err = helpers.UpdateDailyData(filters, update) 
	//getTrend["_id"] = currentCandle["_id"].(primitive.ObjectID)
	getTrend["daily_trend"] = 	new_trend
	//getTrend["openTime_human_readible"] = 	currentCandle["openTime_human_readible"].(string)
	
	
	return getTrend , nil
	
}




func ProcessFinalData(data []interface{}, typeValue int) {
	
	//collectionName := "market_chart"
	fmt.Println("data",data)
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