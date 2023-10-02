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
		startOver := 0
		coin := jsonData["coin"].(string)
		if val, ok := jsonData["forcerun"]; ok {
			// "forcerun" key exists, try to assert it as a string
			if forceRunStr, ok := val.(string); ok {
				// Successfully extracted the value as a string
				if forceRunStr!=""{
					startOver = 1
				}
			} 
		}
		
		for i := 0; i <= diffInDays; i++ {
			setStartDate := startDate.AddDate(0, 0, diffInDays) // years, months, days
			// Set the time components for the start and end dates
			fromDate = time.Date(setStartDate.Year(), setStartDate.Month(), setStartDate.Day(), 0, 0, 0, 0, time.UTC)
			toDate = time.Date(setStartDate.Year(), setStartDate.Month(), setStartDate.Day(), 23, 59, 59, 999999999, time.UTC)

		}
		// dpup indicator
		dpupDirection, _ := dpupTrendDirectionCalculations(coin, startDate, diffInDays, startOver)

		// daily trends
		getTrend, _ := calculateDailyTrend(coin, startDate, diffInDays, startOver)

		response := map[string]interface{}{
			"dpup_direction": dpupDirection,
			"getTrend":       getTrend,
		}
		c.JSON(http.StatusOK, gin.H{
			"status":  200,
			"data":    response, // Replace with your data
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


func dpupTrendDirectionCalculations(coin string, startDate time.Time, diffInDays int, startOver int) ([]map[string]interface{}, error) {
	var getTrend []map[string]interface{}
	var final []interface{}

	startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)

	for i := 0; i <= diffInDays; i++ {
		//var newDpupTrendDirection string
		dateToGet := startDate.AddDate(0, 0, i)
		currentData, err := helpers.GetDailyData(coin, dateToGet, 0)
		if err != nil {
			// Candle is missing for this hour.
			continue
		}

		dataToParse := []interface{}{currentData}
		final = append(final, dataToParse)
	}

	if len(final) == 0 {
		// No Candles found.
		return getTrend, nil
	}

	var counter int
	var historyArr []map[string]interface{}
	for _, data := range final {
		temp33 := make(map[string]interface{})
		var newDpupTrendDirection string
		var lastDpupTrendDirection string
		datatoupdate := data.([]interface{})[0]
		current := datatoupdate.([]interface{})[0].(map[string]interface{})
		currentDay, err := time.Parse(time.RFC3339, current["openTime_human_readible"].(string))
		if err != nil {
			return nil, err
		}
		temp33["currentDay"] = currentDay

		if startOver == 1 && counter == 0 {
			newDpupTrendDirection = "up"
		} else {
			lastDayData, err := helpers.LastIndicatorValue(coin, currentDay, "dpup_trend_direction")
			if err != nil {
				return nil, err
			}

			if len(lastDayData) == 0 {
				temp33["lastDpupTrendDirection"] = nil
				continue
			} else {
				lastDpupTrendDirection = lastDayData[0]["dpup_trend_direction"].(string)
				temp33["lastDpupTrendDirection"] = lastDpupTrendDirection
			}

			dp1Perc := current["DP1_perc"].(float64)
			dp2Perc := current["DP2_perc"].(float64)
			dp3Perc := current["DP3_perc"].(float64)
			up1Perc := current["UP1_perc"].(float64)
			up2Perc := current["UP2_perc"].(float64)
			up3Perc := current["UP3_perc"].(float64)

			if lastDpupTrendDirection == "up" {
				mes := fmt.Sprintf("%.2f,%.2f,%.2f", dp1Perc, dp2Perc, dp3Perc)
				fmt.Println("up", mes)
				if (dp1Perc > 0 && dp1Perc <= 90) && (dp2Perc > 0 && dp2Perc <= 90) && (dp3Perc > 0 && dp3Perc <= 90) {
					newDpupTrendDirection = "down"
				} else {
					if (dp1Perc > 0 && dp1Perc <= 10) && ((dp3Perc == 0 || dp3Perc > 80)) {
						newDpupTrendDirection = "down"
					} else {
						newDpupTrendDirection = lastDpupTrendDirection
					}
				}
			}

			if lastDpupTrendDirection == "down" {
				mes := fmt.Sprintf("%.2f,%.2f,%.2f", up1Perc, up2Perc, up3Perc)
				fmt.Println("down", mes)
				if (up1Perc > 0 && up1Perc <= 90) && (up2Perc > 0 && up2Perc <= 90) && (up3Perc > 0 && up3Perc <= 90) {
					newDpupTrendDirection = "up"
				} else {
					if (up1Perc > 0 && up1Perc <= 10) && ((up3Perc == 0 || up3Perc > 80)) {
						newDpupTrendDirection = "up"
					} else {
						newDpupTrendDirection = lastDpupTrendDirection
					}
				}
			}
		}

		temp33["_id"] = current["_id"]
		temp33["openTime_human_readible"] = current["openTime_human_readible"].(string)
		temp33["newDpupTrendDirection"] = newDpupTrendDirection
		historyArr = append(historyArr, temp33)
		counter++
		//updateDailyCandleTrends("dpup_trend_direction", newDpupTrendDirection, current["_id"].(string))
	}

	return historyArr, nil
}


func calculateDailyTrend(coin string, startDate time.Time, diffInDays int, startOver int) ([]map[string]interface{}, error) {
	fmt.Println("inside Daily Trend")

	var getTrend []map[string]interface{}
	var final []interface{}

	startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)
	//today := time.Now().UTC().UTC().Truncate(24 * time.Hour)

	for i := 0; i <= diffInDays; i++ {
		dateToGet := startDate.AddDate(0, 0, i)
		currentData, err := helpers.GetDailyData(coin, dateToGet, 0)
		if err != nil {
			// Candle is missing for this hour.
			continue
		}

		dataToParse := []interface{}{currentData}
		final = append(final, dataToParse)
	}

	if len(final) == 0 {
		return getTrend, nil
	}

	var finalUpdate []map[string]interface{}
	counter := 0
	for _, data := range final {
		counter++
		tempObject := make(map[string]interface{})
		var newTrend string
		datatoupdate := data.([]interface{})[0]
		current := datatoupdate.([]interface{})[0].(map[string]interface{})

		if len(current) == 0 {
			continue
		}

		currentDay, err := time.Parse(time.RFC3339, current["openTime_human_readible"].(string))
		if err != nil {
			return nil, err
		}

		datetoget := currentDay.AddDate(0, 0, -1)
		previous, err := helpers.GetDailyData(coin, datetoget, 0)
		if err != nil {
			fmt.Println("No Previous Data", currentDay)
			lastDayData, err := helpers.LastIndicatorValue(coin, currentDay, "daily_trend")
			if err != nil {
				return nil, err
			}
			tempObject["lastopenTime_human_readible"] = nil
			newTrend = lastDayData[0]["daily_trend"].(string)
		} else {
			if startOver == 1 && counter == 1 {
				tempObject["lastopenTime_human_readible"] = nil
				newTrend = "up"
			} else {
				previousTrend := previous[0]["daily_trend"].(string)
				tempObject["lastopenTime_human_readible"] = previous[0]["openTime_human_readible"].(string)

				if previousTrend != "" && previousTrend != "null" {
					newTrend = previousTrend
				} else {
					lastDayData, err := helpers.LastIndicatorValue(coin, currentDay, "daily_trend")
					if err != nil {
						return nil, err
					}
					trend := lastDayData[0]["daily_trend"].(string)
					fmt.Println("previousTrend", previousTrend, trend, currentDay)

					currentClose := current["close"].(float64)
					currentOpen := current["open"].(float64)
					currentCandleColor := ""

					if currentClose > currentOpen {
						currentCandleColor = "green"
					} else {
						currentCandleColor = "red"
					}

					previousClose := previous[0]["close"].(float64)
					previousOpen := previous[0]["open"].(float64)
					previousCandleColor := ""

					if previousClose > previousOpen {
						previousCandleColor = "green"
					} else {
						previousCandleColor = "red"
					}

					if trend == "up" {
						if previousCandleColor == "red" {
							if currentClose < previousClose {
								newTrend = "down"
							} else {
								newTrend = trend
							}
						}
						if previousCandleColor == "green" {
							if currentCandleColor == "red" && currentClose < previous[0]["low"].(float64) {
								newTrend = "down"
							} else {
								newTrend = trend
							}
						}
					}

					if trend == "down" {
						if previousCandleColor == "green" {
							if currentClose > previousClose {
								newTrend = "up"
							} else {
								newTrend = trend
							}
						}
						if previousCandleColor == "red" {
							if currentClose > previous[0]["high"].(float64) {
								newTrend = "up"
							} else {
								newTrend = trend
							}
						}
					}
				}
			}
		}

		tempObject["daily_trend"] = newTrend
		tempObject["_id"] = current["_id"]
		tempObject["openTime_human_readible"] = current["openTime_human_readible"].(string)
		tempObject["trend"] = current["daily_trend"].(string)
		finalUpdate = append(finalUpdate, tempObject)
	}

	return finalUpdate, nil
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