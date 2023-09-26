package helpers

import(
	"go.mongodb.org/mongo-driver/bson"
	"indicatorsAPP/mongohelpers"
	"time"
	"fmt"
)

func MarketChartDataForCoin(filters bson.M) ([]bson.M, error){
	collectionName := "market_chart"
	projection := bson.M{}
	var limit int64 = 0
	var sortOrder int  = 0
	sortBy := ""
	docs ,err := mongohelpers.MongoFind(collectionName, filters,projection, limit, sortOrder, sortBy)
	if err!=nil{
		return []bson.M{} , err
	}
	return docs , nil
}

func GetLastCandle(coin string,timestamp time.Time,durationLimit int) ([]bson.M, error){
	// filters processing
	duration := time.Duration(durationLimit) * time.Hour
	newDate := time.Now().UTC()
	startDate := newDate.Add(-duration)
	endDate := newDate
	

	// filters
	filters := bson.M{
		"coin":         coin,
		"created_date": bson.M{"$gte": startDate, "$lte": endDate},
	}

	collectionName := "market_chart"
	projection := bson.M{}
	var limit int64 = 1
	var sortOrder int  = 0
	sortBy := ""
	docs ,err := mongohelpers.MongoFind(collectionName, filters,projection, limit, sortOrder, sortBy)
	if err!=nil{
		return []bson.M{} , err
	}
	return docs , nil
}


func DPUPPercentileData(coin string, toArr map[string]interface{}, duration int) ([]bson.M, error) {
	// Convert duration to a negative duration for subtraction
	startCandleDate := toArr["startCandleDate"].(time.Time)
	endCandleDate := toArr["endCandleDate"].(time.Time)
	
	
	duration *= -1
	startDate := startCandleDate.Add(time.Duration(duration) * time.Hour)
	
	match := bson.M{
		"coin":        coin,
		"created_date": bson.M{"$gte": startDate, "$lte": endCandleDate},
	}

	pipeline := []bson.M{
		{"$match": match},
		{"$sort": bson.M{"coin": 1}},
		{
			"$group": bson.M{
				"_id":  bson.M{"coin": "$coin"},
				"DP1": bson.M{"$push": "$DP1"},
				"DP2": bson.M{"$push": "$DP2"},
				"DP3": bson.M{"$push": "$DP3"},
				"UP1": bson.M{"$push": "$UP1"},
				"UP2": bson.M{"$push": "$UP2"},
				"UP3": bson.M{"$push": "$UP3"},
			},
		},
	}
	collectionName := "market_chart"

	docs ,err := mongohelpers.MongoAggregate(collectionName, pipeline)
	if err!=nil{
		return []bson.M{} , err
	}
	return docs , nil
	
}


func LastCandelTrends(coin string,timestamp time.Time,filtertype string) ([]bson.M, error){

	// filters
	filters := bson.M{
		"coin":         coin,
		"created_date": bson.M{"$lt": timestamp},
	}

	if filtertype == "candel_trends"{
		filters["$and"] = []bson.M{
				{
					"candel_trends": bson.M{
						"$exists": true,
					},
				},
				{
					"candel_trends": bson.M{
						"$ne": "",
					},
				},
			}
		}

	collectionName := "market_chart"
	projection := bson.M{
		"close":                   1,
        "candel_trends":           1,
        "openTime_human_readible": 1,
        "closeTime_human_readible": 1,
	}
	var limit int64 = 1
	var sortOrder int  = -1
	sortBy := "created_date"
	docs ,err := mongohelpers.MongoFind(collectionName, filters,projection, limit, sortOrder, sortBy)
	if err!=nil{
		return []bson.M{} , err
	}
	return docs , nil
}


func GetDailyData(coin string,date time.Time,durationLimit int) ([]bson.M, error){
	// filters processing
	//duration := time.Duration(durationLimit)
	newDate := date.AddDate(0, 0, -durationLimit)
	startDate := time.Date(newDate.Year(), newDate.Month(), newDate.Day(), 0, 0, 0, 0, time.UTC)
	endDate := time.Date(newDate.Year(), newDate.Month(), newDate.Day(), 23, 59, 59, 999999999, time.UTC)
	

	// filters
	filters := bson.M{
		"coin":         coin,
		"created_date": bson.M{"$gte": startDate, "$lte": endDate},
	}

	collectionName := "market_chart_daily"
	projection := bson.M{
		"created_date":            1,
		"dpup_trend_direction":    1,
		"UP1_perc":                1,
		"UP2_perc":                1,
		"UP3_perc":                1,
		"DP1_perc":                1,
		"DP2_perc":                1,
		"DP3_perc":                1,
		"open":                    1,
		"daily_trend":             1,
		"coin":                    1,
		"low":                     1,
		"high":                    1,
		"candle_color":            1,
		"close":                   1,
		"openTime_human_readible": 1,
		"closeTime_human_readible": 1,
	}
	var limit int64 = 0
	var sortOrder int  = -1
	sortBy := "created_date"
	docs ,err := mongohelpers.MongoFind(collectionName, filters,projection, limit, sortOrder, sortBy)
	if err!=nil{
		return []bson.M{} , err
	}
	return docs , nil
}

func GetHHSwingStatusFromCandleChart(coin string,date time.Time) ([]bson.M, error){
	filters := bson.M{
		"coin":         coin,
		"global_swing_status":"HH",
		"created_date": bson.M{"$lt": date},
	}

	collectionName := "market_chart"
	projection := bson.M{
		"created_date"               :    1,
		"highest_swing_point"        :    1,
		"openTime_human_readible"    :    1,
		"closeTime_human_readible"   :    1,
	}
	var limit int64    = 0
	var sortOrder int  = -1
	sortBy             := "created_date"
	docs ,err := mongohelpers.MongoFind(collectionName, filters,projection, limit, sortOrder, sortBy)
	if err!=nil{
		return []bson.M{} , err
	}
	return docs , nil
}

func GetLHSwingStatusFromCandleChart(coin string,date time.Time) ([]bson.M, error){
	filters := bson.M{
		"coin":         coin,
		"global_swing_status":"LH",
		"created_date": bson.M{"$lt": date},
	}

	collectionName := "market_chart"
	projection := bson.M{
		"created_date"               :    1,
		"highest_swing_point"        :    1,
		"openTime_human_readible"    :    1,
		"closeTime_human_readible"   :    1,
	}
	var limit int64    = 0
	var sortOrder int  = -1
	sortBy             := "created_date"
	docs ,err := mongohelpers.MongoFind(collectionName, filters,projection, limit, sortOrder, sortBy)
	if err!=nil{
		return []bson.M{} , err
	}
	return docs , nil
}

func GetLLSwingStatusFromCandleChart(coin string,date time.Time) ([]bson.M, error){
	filters := bson.M{
		"coin":         coin,
		"global_swing_status":"LL",
		"created_date": bson.M{"$lt": date},
	}

	collectionName := "market_chart"
	projection := bson.M{
		"created_date"               :    1,
		"highest_swing_point"        :    1,
		"openTime_human_readible"    :    1,
		"closeTime_human_readible"   :    1,
	}
	var limit int64    = 0
	var sortOrder int  = -1
	sortBy             := "created_date"
	docs ,err := mongohelpers.MongoFind(collectionName, filters,projection, limit, sortOrder, sortBy)
	if err!=nil{
		return []bson.M{} , err
	}
	return docs , nil
}

func GetHLSwingStatusFromCandleChart(coin string,date time.Time) ([]bson.M, error){
	filters := bson.M{
		"coin":         coin,
		"global_swing_status":"HL",
		"created_date": bson.M{"$lt": date},
	}

	collectionName := "market_chart"
	projection := bson.M{
		"created_date"               :    1,
		"highest_swing_point"        :    1,
		"openTime_human_readible"    :    1,
		"closeTime_human_readible"   :    1,
	}
	var limit int64    = 0
	var sortOrder int  = -1
	sortBy             := "created_date"
	docs ,err := mongohelpers.MongoFind(collectionName, filters,projection, limit, sortOrder, sortBy)
	if err!=nil{
		return []bson.M{} , err
	}
	return docs , nil
}

func LastIndicatorValue(coin string, date time.Time, key string) ([]bson.M, error) {
	collectionName := "market_chart"
	filters := bson.M{
		"coin": coin,
		"created_date": bson.M{
			"$lt": date,
		},
	}

	if key == "daily_trend" {
		filters["$and"] = []bson.M{
			{"daily_trend": bson.M{"$exists": true}},
			{"daily_trend": bson.M{"$ne": ""}},
		}
	} else if key == "dpup_trend_direction" {
		filters["$and"] = []bson.M{
			{"dpup_trend_direction": bson.M{"$exists": true}},
			{"dpup_trend_direction": bson.M{"$ne": ""}},
		}
	}

	projection := bson.M{
		"created_date":            1,
		"dpup_trend_direction":    1,
		"UP1_perc":                1,
		"UP2_perc":                1,
		"UP3_perc":                1,
		"DP1_perc":                1,
		"DP2_perc":                1,
		"DP3_perc":                1,
		"open":                    1,
		"daily_trend":             1,
		"coin":                    1,
		"low":                     1,
		"high":                    1,
		"candle_color":            1,
		"close":                   1,
		"openTime_human_readible": 1,
		"closeTime_human_readible": 1,
	}

	var limit int64    = 1
	var sortOrder int  = -1
	sortBy             := "created_date"
	docs ,err := mongohelpers.MongoFind(collectionName, filters,projection, limit, sortOrder, sortBy)
	if err!=nil{
		return []bson.M{} , err
	}
	return docs , nil
}


func ListCoins() ([]bson.M,error){
	collectionName := "coins"
	var limit int64 = 0
	var sortOrder int = 1
	var sortBy string = "symbol"
	excludedSymbols := []string{"ZENBTC", "POEBTC", "NCASHBTC", "XEMBTC"}
	projection := bson.M{"symbol":1}
	// Create a filter to find documents where "symbol" is not in the excluded list
	filters := bson.M{"symbol": bson.M{"$nin": excludedSymbols}}
	docs ,err := mongohelpers.MongoFind(collectionName, filters,projection, limit, sortOrder, sortBy)
	if err!=nil{
		return []bson.M{} , err
	}
	return docs , nil
}


// getStartOfCurrentHour returns the start time of the current hour in ISO8601 format
func getStartOfCurrentHour() string {
	currentTime := time.Now().UTC()
	startOfCurrentHour := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), currentTime.Hour(), 0, 0, 0,currentTime.Location())
	startOfCurrentHour = startOfCurrentHour.Add(-time.Hour) // Decrement the hour by 1 to get the previous hour
	strDate := startOfCurrentHour.Format("2006-01-02 15:04:05")
	fmt.Println("strDate",strDate)
	return strDate
}

func GetBodyMoveAverage(coin string) ([]bson.M,error) {
	collectionName := "market_chart"
	var limit int64 = 1
	var sortOrder int = -1
	var sortBy string = "created_date"
	filters  := bson.M{
		"coin":                   coin,
		"openTime_human_readible": getStartOfCurrentHour(),
	}
	projection := bson.M{
		"close":                 1,
		"open":                  1,
		"openTime_human_readible": 1,
		"created_date":          1,
		"coin":                  1,
		"body_move_average":     1,
		"closeTime_human_readible": 1,
	}
	docs ,err := mongohelpers.MongoFind(collectionName, filters,projection, limit, sortOrder, sortBy)
	if err!=nil{
		return []bson.M{} , err
	}
	return docs , nil
}


func FetchMarketPrices(coin string,timestamp time.Time) ([]bson.M,error) {
	collectionName := "market_prices"
	var limit int64 = 1
	var sortOrder int = -1
	var sortBy string = "created_date"
	filters  := bson.M{
		"coin":                   coin,
		"created_date":  bson.M{"$gte":timestamp},
	}
	projection := bson.M{
		"created_date":1,
		"price":1,
	}
	docs ,err := mongohelpers.MongoFind(collectionName, filters,projection, limit, sortOrder, sortBy)
	if err!=nil{
		return []bson.M{} , err
	}
	return docs , nil
}

func AddRaiseDropEntry(filters bson.M, toSet bson.M , upsert bool) error{
	collectionName := "big_drop_and_pull_back_track_testing"
	err := mongohelpers.MongoUpdateOne(collectionName,filters,toSet,upsert)
	if err !=nil{
		return err
	}
	return nil
}