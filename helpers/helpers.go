package helpers

import (
	"fmt"
	"indicatorsAPP/mongohelpers"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func FetchCandlesData(coin, queryType string, start time.Time, end time.Time) ([]bson.M, error) {
	collectionName := "market_chart"
	if queryType == "daily" {
		collectionName = "market_chart_daily"
	}
	filters := bson.M{
		"coin":         coin,
		"created_date": bson.M{"$gte": start, "$lte": end},
	}
	projection := bson.M{
		"_id":      0,
		"openTime": 1,
		"open":     1,
		"coin":     1,
		"low":      1,
		"high":     1,
		"volume":   1,
		"close":    1,
	}
	var limit int64 = 0
	var sortOrder int = -1
	sortBy := "created_date"
	docs, err := mongohelpers.MongoFind(collectionName, filters, projection, limit, sortOrder, sortBy)
	if err != nil {
		return []bson.M{}, err
	}
	return docs, nil
}

func MarketChartDataForCoin(filters bson.M) ([]bson.M, error) {
	collectionName := "market_chart"
	projection := bson.M{
		"created_date": 1,
		//"dpup_trend_direction":    1,
		//"UP1_perc":                1,
		//"UP2_perc":                1,
		//"UP3_perc":                1,
		//"DP1_perc":                1,
		//"DP2_perc":                1,
		//"DP3_perc":                1,
		"open": 1,
		//"daily_trend":             1,
		"coin":                     1,
		"low":                      1,
		"high":                     1,
		"candle_color":             1,
		"close":                    1,
		"openTime_human_readible":  1,
		"closeTime_human_readible": 1,
	}
	var limit int64 = 0
	var sortOrder int = 0
	sortBy := ""
	docs, err := mongohelpers.MongoFind(collectionName, filters, projection, limit, sortOrder, sortBy)
	if err != nil {
		return []bson.M{}, err
	}
	return docs, nil
}

func GetLastCandle(coin string, timestamp time.Time, durationLimit int) ([]bson.M, error) {
	// filters processing
	//duration := time.Duration(durationLimit) * time.Hour
	newDate := timestamp

	startDate := time.Date(newDate.Year(), newDate.Month(), newDate.Day(), newDate.Hour()-durationLimit, 0, 0, 0, time.UTC)
	endDate := time.Date(newDate.Year(), newDate.Month(), newDate.Day(), newDate.Hour()-durationLimit, 59, 59, 0, time.UTC)

	// filters
	filters := bson.M{
		"coin":         coin,
		"created_date": bson.M{"$gte": startDate, "$lte": endDate},
	}

	collectionName := "market_chart"
	projection := bson.M{}
	var limit int64 = 1
	var sortOrder int = 0
	sortBy := ""
	docs, err := mongohelpers.MongoFind(collectionName, filters, projection, limit, sortOrder, sortBy)
	if err != nil {
		return []bson.M{}, err
	}
	return docs, nil
}

func DPUPPercentileData(coin string, toArr map[string]interface{}, duration int) ([]bson.M, error) {
	// Convert duration to a negative duration for subtraction
	startCandleDate := toArr["startCandleDate"].(time.Time)
	endCandleDate := toArr["endCandleDate"].(time.Time)

	// 30 Days back so multiple by 24 , duration is 30 ,
	duration *= -24
	startDate := startCandleDate.Add(time.Duration(duration) * time.Hour)
	endDate := endCandleDate
	match := bson.M{
		"coin":         coin,
		"created_date": bson.M{"$gte": startDate, "$lte": endDate},
	}
	//fmt.Println("match",match)
	pipeline := []bson.M{
		{"$match": match},
		{"$sort": bson.M{"coin": 1}},
		{
			"$group": bson.M{
				"_id": bson.M{"coin": "$coin"},
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

	docs, err := mongohelpers.MongoAggregate(collectionName, pipeline)
	if err != nil {
		return []bson.M{}, err
	}
	return docs, nil

}

func LastCandelTrends(coin string, timestamp time.Time, filtertype string) ([]bson.M, error) {

	// filters
	filters := bson.M{
		"coin":         coin,
		"created_date": bson.M{"$lt": timestamp},
	}

	if filtertype == "candel_trends" {
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
		"close":                    1,
		"candel_trends":            1,
		"openTime_human_readible":  1,
		"closeTime_human_readible": 1,
	}
	var limit int64 = 1
	var sortOrder int = -1
	sortBy := "created_date"
	docs, err := mongohelpers.MongoFind(collectionName, filters, projection, limit, sortOrder, sortBy)
	if err != nil {
		return []bson.M{}, err
	}
	return docs, nil
}

func GetDailyCandleData(coin string, startDate, endDate time.Time, limit int64) ([]bson.M, error) {
	// filters
	filters := bson.M{
		"coin":         coin,
		"created_date": bson.M{"$gte": startDate, "$lte": endDate},
	}

	collectionName := "market_chart_daily"
	projection := bson.M{
		"created_date":             1,
		"dpup_trend_direction":     1,
		"UP1_perc":                 1,
		"UP2_perc":                 1,
		"UP3_perc":                 1,
		"DP1_perc":                 1,
		"DP2_perc":                 1,
		"DP3_perc":                 1,
		"open":                     1,
		"daily_trend":              1,
		"coin":                     1,
		"low":                      1,
		"high":                     1,
		"candle_color":             1,
		"close":                    1,
		"openTime_human_readible":  1,
		"closeTime_human_readible": 1,
	}

	var sortOrder int = -1
	sortBy := "created_date"
	docs, err := mongohelpers.MongoFind(collectionName, filters, projection, limit, sortOrder, sortBy)
	if err != nil {
		return []bson.M{}, err
	}
	return docs, nil
}
func GetDailyData(coin string, date time.Time, durationLimit int) ([]bson.M, error) {
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
		"created_date":             1,
		"dpup_trend_direction":     1,
		"UP1_perc":                 1,
		"UP2_perc":                 1,
		"UP3_perc":                 1,
		"DP1_perc":                 1,
		"DP2_perc":                 1,
		"DP3_perc":                 1,
		"open":                     1,
		"daily_trend":              1,
		"coin":                     1,
		"low":                      1,
		"high":                     1,
		"candle_color":             1,
		"close":                    1,
		"openTime_human_readible":  1,
		"closeTime_human_readible": 1,
	}
	var limit int64 = 0
	var sortOrder int = -1
	sortBy := "created_date"
	docs, err := mongohelpers.MongoFind(collectionName, filters, projection, limit, sortOrder, sortBy)
	if err != nil {
		return []bson.M{}, err
	}
	return docs, nil
}

func GetHHSwingStatusFromCandleChart(coin string, date time.Time) ([]bson.M, error) {
	filters := bson.M{
		"coin":                coin,
		"global_swing_status": "HH",
		"created_date":        bson.M{"$lt": date},
	}

	collectionName := "market_chart"
	projection := bson.M{
		"created_date":             1,
		"highest_swing_point":      1,
		"openTime_human_readible":  1,
		"closeTime_human_readible": 1,
	}
	var limit int64 = 1
	var sortOrder int = -1
	sortBy := "created_date"
	docs, err := mongohelpers.MongoFind(collectionName, filters, projection, limit, sortOrder, sortBy)
	if err != nil {
		fmt.Println("err", err)
		return []bson.M{}, err
	}
	return docs, nil
}

func GetLHSwingStatusFromCandleChart(coin string, date time.Time) ([]bson.M, error) {
	filters := bson.M{
		"coin":                coin,
		"global_swing_status": "LH",
		"created_date":        bson.M{"$lt": date},
	}

	collectionName := "market_chart"
	projection := bson.M{
		"created_date":             1,
		"highest_swing_point":      1,
		"openTime_human_readible":  1,
		"closeTime_human_readible": 1,
	}
	var limit int64 = 1
	var sortOrder int = -1
	sortBy := "created_date"
	docs, err := mongohelpers.MongoFind(collectionName, filters, projection, limit, sortOrder, sortBy)
	if err != nil {
		return []bson.M{}, err
	}
	return docs, nil
}

func GetLLSwingStatusFromCandleChart(coin string, date time.Time) ([]bson.M, error) {
	filters := bson.M{
		"coin":                coin,
		"global_swing_status": "LL",
		"created_date":        bson.M{"$lt": date},
	}

	collectionName := "market_chart"
	projection := bson.M{
		"created_date":             1,
		"lowest_swing_point":       1,
		"openTime_human_readible":  1,
		"closeTime_human_readible": 1,
	}
	var limit int64 = 1
	var sortOrder int = -1
	sortBy := "created_date"
	docs, err := mongohelpers.MongoFind(collectionName, filters, projection, limit, sortOrder, sortBy)
	if err != nil {
		return []bson.M{}, err
	}
	return docs, nil
}

func GetHLSwingStatusFromCandleChart(coin string, date time.Time) ([]bson.M, error) {
	filters := bson.M{
		"coin":                coin,
		"global_swing_status": "HL",
		"created_date":        bson.M{"$lt": date},
	}

	collectionName := "market_chart"
	projection := bson.M{
		"created_date":             1,
		"lowest_swing_point":       1,
		"openTime_human_readible":  1,
		"closeTime_human_readible": 1,
	}
	var limit int64 = 1
	var sortOrder int = -1
	sortBy := "created_date"
	docs, err := mongohelpers.MongoFind(collectionName, filters, projection, limit, sortOrder, sortBy)
	if err != nil {
		return []bson.M{}, err
	}
	return docs, nil
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
		"created_date":             1,
		"dpup_trend_direction":     1,
		"UP1_perc":                 1,
		"UP2_perc":                 1,
		"UP3_perc":                 1,
		"DP1_perc":                 1,
		"DP2_perc":                 1,
		"DP3_perc":                 1,
		"open":                     1,
		"daily_trend":              1,
		"coin":                     1,
		"low":                      1,
		"high":                     1,
		"candle_color":             1,
		"close":                    1,
		"openTime_human_readible":  1,
		"closeTime_human_readible": 1,
	}

	var limit int64 = 1
	var sortOrder int = -1
	sortBy := "created_date"
	docs, err := mongohelpers.MongoFind(collectionName, filters, projection, limit, sortOrder, sortBy)
	if err != nil {
		return []bson.M{}, err
	}
	return docs, nil
}

func ListCoins() ([]bson.M, error) {
	collectionName := "coins"
	var limit int64 = 0
	var sortOrder int = 1
	var sortBy string = "symbol"
	excludedSymbols := []string{"ZENBTC", "POEBTC", "NCASHBTC", "XEMBTC"}
	projection := bson.M{"symbol": 1}
	// Create a filter to find documents where "symbol" is not in the excluded list
	filters := bson.M{"symbol": bson.M{"$nin": excludedSymbols}}
	docs, err := mongohelpers.MongoFind(collectionName, filters, projection, limit, sortOrder, sortBy)
	if err != nil {
		return []bson.M{}, err
	}
	return docs, nil
}

// getStartOfCurrentHour returns the start time of the current hour in ISO8601 format
func getStartOfCurrentHour() string {
	currentTime := time.Now().UTC()
	startOfCurrentHour := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), currentTime.Hour(), 0, 0, 0, currentTime.Location())
	startOfCurrentHour = startOfCurrentHour.Add(-time.Hour) // Decrement the hour by 1 to get the previous hour
	strDate := startOfCurrentHour.Format("2006-01-02 15:04:05")
	//fmt.Println("strDate",strDate)
	return strDate
}

func CurrentHourCandleID(coin string) ([]bson.M, error) {
	collectionName := "market_chart"
	var limit int64 = 1
	var sortOrder int = -1
	var sortBy string = "created_date"
	filters := bson.M{
		"coin": coin,
		// "openTime_human_readible": getStartOfCurrentHour(),
	}
	projection := bson.M{
		"close": 1,
	}
	docs, err := mongohelpers.MongoFind(collectionName, filters, projection, limit, sortOrder, sortBy)
	if err != nil {
		return []bson.M{}, err
	}
	return docs, nil
}

func GetBodyMoveAverage(coin string) ([]bson.M, error) {
	collectionName := "market_chart"
	var limit int64 = 1
	var sortOrder int = -1
	var sortBy string = "created_date"
	filters := bson.M{
		"coin": coin,
		// "openTime_human_readible": getStartOfCurrentHour(),
	}
	projection := bson.M{
		"close":                    1,
		"open":                     1,
		"openTime_human_readible":  1,
		"created_date":             1,
		"coin":                     1,
		"body_move_average":        1,
		"closeTime_human_readible": 1,
	}
	docs, err := mongohelpers.MongoFind(collectionName, filters, projection, limit, sortOrder, sortBy)
	if err != nil {
		return []bson.M{}, err
	}
	return docs, nil
}

func GetCoinCurrentWickMove(coin, closeTime_human_readible string) ([]bson.M, error) {
	collectionName := "market_chart"
	var limit int64 = 1
	var sortOrder int = -1
	var sortBy string = "created_date"
	filters := bson.M{
		"coin":                     coin,
		"closeTime_human_readible": closeTime_human_readible,
	}
	projection := bson.M{
		"lower_wick_per_move": 1,
	}
	docs, err := mongohelpers.MongoFind(collectionName, filters, projection, limit, sortOrder, sortBy)
	if err != nil {
		return []bson.M{}, err
	}
	return docs, nil
}

func FetchMarketPrices(coin string, timestamp time.Time) ([]bson.M, error) {
	collectionName := "market_prices"
	var limit int64 = 1
	var sortOrder int = -1
	var sortBy string = "created_date"
	filters := bson.M{
		"coin":         coin,
		"created_date": bson.M{"$gte": timestamp},
	}
	projection := bson.M{
		"created_date": 1,
		"price":        1,
		"coin":         1,
	}
	docs, err := mongohelpers.MongoFind(collectionName, filters, projection, limit, sortOrder, sortBy)
	if err != nil {
		return []bson.M{}, err
	}
	return docs, nil
}

func AddRaiseDropEntry(filters bson.M, toSet bson.M, upsert bool) error {
	collectionName := "big_drop_and_pull_back_track"
	err := mongohelpers.MongoUpdateOne(collectionName, filters, toSet, upsert)
	if err != nil {
		return err
	}
	return nil
}

func FetchRaiseORDropEmptyPullBackTime(filterType string) ([]bson.M, error) {
	collectionName := "big_drop_and_pull_back_track"
	var limit int64 = 0
	var sortOrder int = 1
	var sortBy string = "coin"
	filters := bson.M{
		"type":           filterType,
		"pull_back_time": bson.M{"$exists": false},
	}
	projection := bson.M{}
	docs, err := mongohelpers.MongoFind(collectionName, filters, projection, limit, sortOrder, sortBy)
	if err != nil {
		return []bson.M{}, err
	}
	return docs, nil
}

func GetPullBackPrice(coin string, price float64, startTime time.Time, raise bool) ([]bson.M, error) {
	collectionName := "market_prices"
	var limit int64 = 1
	var sortOrder int = -1
	var sortBy string = "created_date"
	filters := bson.M{
		"coin":         coin,
		"created_date": bson.M{"$gte": startTime},
	}
	if raise {
		filters["price"] = bson.M{"$lte": price}
	} else {
		filters["price"] = bson.M{"$gte": price}
	}
	projection := bson.M{
		"created_date": 1,
		"price":        1,
	}
	docs, err := mongohelpers.MongoFind(collectionName, filters, projection, limit, sortOrder, sortBy)
	if err != nil {
		return []bson.M{}, err
	}
	return docs, nil
}

func UpdatePullBackTime(id primitive.ObjectID) error {
	collectionName := "big_drop_and_pull_back_track"
	filters := bson.M{
		"_id": id,
	}
	upsert := false
	toSet := bson.M{
		"$set": bson.M{
			"pull_back_time": time.Now(),
			"completed":      1,
		},
	}
	err := mongohelpers.MongoUpdateOne(collectionName, filters, toSet, upsert)
	if err != nil {
		return err
	}
	return nil
}

func UpdateMarketTrendingEntry(coin string, raise bool) error {
	collectionName := "market_trending"
	filters := bson.M{
		"coin": coin,
	}
	upsert := false
	update := bson.M{}
	if raise {
		update["big_raise_pull_back"] = "yes"
	} else {
		update["big_drop_pull_back"] = "yes"
	}
	toSet := bson.M{
		"$set": update,
	}
	err := mongohelpers.MongoUpdateOne(collectionName, filters, toSet, upsert)
	if err != nil {
		return err
	}
	return nil
}

func AddCandleTrack(data bson.M) error {
	collectionName := "big_drop_and_pull_back_track_for_candle"
	_, err := mongohelpers.MongoInsertOne(collectionName, data)
	if err != nil {
		return err
	}
	return nil
}

func InsertTestData(data bson.M) error {
	collectionName := "logs_testing_big_drop_raise"
	_, err := mongohelpers.MongoInsertOne(collectionName, data)
	if err != nil {
		return err
	}
	return nil
}

func ToFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case int:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case float32:
		return float64(v), true
	case float64:
		return v, true
	case uint:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	default:
		return 0, false
	}
}

func UpdateStrategyValue(id primitive.ObjectID, strategy, coin string) error {
	collectionName := "market_chart"
	filters := bson.M{
		"_id":  id,
		"coin": coin,
	}
	upsert := false
	toSet := bson.M{
		"$set": bson.M{
			"strategy": strategy,
		},
	}
	err := mongohelpers.MongoUpdateOne(collectionName, filters, toSet, upsert)
	if err != nil {
		return err
	}
	return nil
}

func ConvertStrNumbersToFloat(number string) (float64, error) {
	f, err := strconv.ParseFloat(number, 64)
	return f, err
}

func UpdateDailyData(filters, update bson.M) error {
	collectionName := "market_chart_daily"
	err := mongohelpers.MongoUpdateOne(collectionName, filters, update, false)
	return err
}

// ConvertToType attempts to convert an interface to the specified target type.
func ConvertToType(value interface{}, targetType interface{}) (interface{}, error) {
	switch targetType.(type) {
	case int:
		if v, ok := value.(int); ok {
			return v, nil
		}
	case int16:
		if v, ok := value.(int16); ok {
			return v, nil
		}
	case int32:
		if v, ok := value.(int32); ok {
			return v, nil
		}
	case int64:
		if v, ok := value.(int64); ok {
			return v, nil
		}
	case float32:
		if v, ok := value.(float32); ok {
			return v, nil
		}
	case float64:
		if v, ok := value.(float64); ok {
			return v, nil
		}
	case string:
		if v, ok := value.(string); ok {
			return v, nil
		}
	}

	return nil, fmt.Errorf("Conversion not supported")
}

func UpdateHourlyData(filters, toSet bson.M) error {
	collectionName := "market_chart"
	err := mongohelpers.MongoUpdateOne(collectionName, filters, toSet, false)
	if err != nil {
		return err
	}
	return nil
}

func GetCurrentDownBarrier(coin string) ([]bson.M, error) {
	collectionName := "barrier_values_collection"

	filters := bson.M{
		"coin":           coin,
		"barrier_type":   "down",
		"barrier_status": "very_strong_barrier",
	}
	projection := bson.M{
		"_id":                        0,
		"barier_value":               1,
		"coin":                       1,
		"created_date":               1,
		"global_swing_parent_status": 1,
	}
	// find first record sorted by date.
	var limit int64 = 1
	var sortOrder int = -1
	sortBy := "created_date"
	docs, err := mongohelpers.MongoFind(collectionName, filters, projection, limit, sortOrder, sortBy)
	if err != nil {
		return []bson.M{}, err
	}
	return docs, nil
}

func GetLastDownBarrier(coin string, currentBarrierTime time.Time) ([]bson.M, error) {

	collectionName := "barrier_values_collection"

	filters := bson.M{
		"coin":           coin,
		"barrier_type":   "down",
		"barrier_status": "very_strong_barrier",
		"created_date":   bson.M{"$lt": currentBarrierTime},
	}
	projection := bson.M{
		"_id":                        0,
		"barier_value":               1,
		"coin":                       1,
		"created_date":               1,
		"global_swing_parent_status": 1,
	}
	// find first record sorted by date.
	var limit int64 = 1
	var sortOrder int = -1
	sortBy := "created_date"
	docs, err := mongohelpers.MongoFind(collectionName, filters, projection, limit, sortOrder, sortBy)
	if err != nil {
		return []bson.M{}, err
	}
	return docs, nil
}

func GetNextSwingPoint(coin string, currentBarrierTime time.Time) ([]bson.M, error) {

	collectionName := "barrier_values_collection"

	filters := bson.M{
		"coin":                       coin,
		"barrier_type":               "up",
		"global_swing_parent_status": "HH",
		"barrier_status":             "very_strong_barrier",
		//"created_date": bson.M{"$lte": currentBarrierTime},
	}
	projection := bson.M{
		"_id":                        0,
		"barier_value":               1,
		"coin":                       1,
		"created_date":               1,
		"global_swing_parent_status": 1,
	}
	// find first record sorted by date.
	var limit int64 = 1
	var sortOrder int = -1
	sortBy := "created_date"
	docs, err := mongohelpers.MongoFind(collectionName, filters, projection, limit, sortOrder, sortBy)
	if err != nil {
		return []bson.M{}, err
	}
	return docs, nil
}

func UpdateDownBarrierRejectionData(id primitive.ObjectID, data bson.M) error {
	collectionName := "market_chart"
	filters := bson.M{
		"_id": id,
	}
	upsert := false
	toSet := bson.M{
		"$set": data,
	}
	err := mongohelpers.MongoUpdateOne(collectionName, filters, toSet, upsert)
	if err != nil {
		return err
	}
	return nil
}

func CurrentCandelTrend(coin string) (bson.M, error) {
	// Filters
	filters := bson.M{
		"coin": coin,
	}

	collectionName := "market_chart"
	projection := bson.M{
		"candel_trends": 1,
	}

	// Retrieve only the latest document
	var limit int64 = 1
	var sortOrder int = -1
	sortBy := "created_date"
	docs, err := mongohelpers.MongoFind(collectionName, filters, projection, limit, sortOrder, sortBy)
	if err != nil {
		return bson.M{}, err
	}
	if len(docs) > 0 {
		return docs[0], nil
	}
	return bson.M{}, nil
}
