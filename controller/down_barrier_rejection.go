package controller
import (
	"fmt"
	"sync"
	"time"
	//"reflect"
	"indicatorsAPP/helpers"
	"indicatorsAPP/utils"
	//"indicatorsAPP/mongohelpers"
	"go.mongodb.org/mongo-driver/bson/primitive"
	
	"go.mongodb.org/mongo-driver/bson"
)


var hourChangeReset time.Time
var barriersData []BarriersDataStruct
var coinListCacheForBarrier []bson.M
var rejectionMutex sync.Mutex
// Defaults ....
var wickMoveFactorValue  float64  = 2
var debug = true

type BarriersDataStruct struct {
	Symbol                 		string 
	CurrentDownBarrier          float64
	LastDownBarrier				float64
	NextHighSwingPoint          float64
	CurrentDownBarrierTime      time.Time 
	PreviousDownBarrierTime     time.Time 
	HighSwingTime     		    time.Time
	Date 						time.Time
	HighPointDiff			    float64
	LowPointDiff			    float64
	Wickvalue					float64
	Triggered					bool
	BarrierDropped				bool
	AddedForCurrentHour			bool
	CurrentCandleID             primitive.ObjectID 
}



/*
rejectionMutex.Lock() // Acquire the mutex
// Access the shared resource
rejectionMutex.Unlock() // Release the mutex when done
*/
//ruleType := "buy"
func RunDownBarrierRejection(){
	currentDateTime := time.Now().UTC()
	currentHourDate := time.Date(currentDateTime.Year(), currentDateTime.Month(), currentDateTime.Day(), currentDateTime.Hour(), 0, 0, 0, currentDateTime.Location())

	if hourChangeReset.IsZero() || currentHourDate.After(hourChangeReset) {
		fmt.Println("RunDownBarrierRejection ",currentHourDate)
		barriersData = nil
		hourChangeReset = currentHourDate
	}

	if len(coinListCacheForBarrier) ==0 {
		coinList , err := helpers.ListCoins()
		if err!=nil{
			debugLogs("RunDownBarrierRejection Error",err)
			return
		}
		coinListCacheForBarrier = coinList
	}
	//fmt.Println("coinListCacheForBarrier",coinListCacheForBarrier)
	//coinList a bson.M document
	for _, currentCoin := range coinListCacheForBarrier {
		coinSymbol := currentCoin["symbol"].(string)
		if coinSymbol != "LINKUSDT" && coinSymbol != "DOTBTC" {
			continue;
		}	
		rejectionMutex.Lock()
		foundPriceObject := findBarriersObject(coinSymbol, currentHourDate)
		if foundPriceObject != nil {
			debugLogs("Big Drop Found in foundPriceObject for coin:", coinSymbol, *foundPriceObject)
			rejectionMutex.Unlock()
			continue
		}
		// only to verify if current candles data is set or not.
		bodyMoveAverage, err := helpers.GetBodyMoveAverage(coinSymbol)
		//fmt.Println("bodyMoveAverage",bodyMoveAverage)
		if err!=nil{
			debugLogs("bodyMoveAverage Error "+coinSymbol,err)
			rejectionMutex.Unlock()
			continue
		}
		if len(bodyMoveAverage) == 0 || len(bodyMoveAverage[0]) == 0 {
			debugLogs("bodyMoveAverage is empty for current Hour Down Barrier Rejection "+coinSymbol)
			rejectionMutex.Unlock()
			continue
		}
		
		var CurrentCandleID primitive.ObjectID 
		if id , ok := bodyMoveAverage[0]["_id"].(primitive.ObjectID); ok{
			CurrentCandleID =  id
		}  else {
			utils.CheckType(bodyMoveAverage[0]["_id"]);
			debugLogs("CurrentCandleID is Invalid "+coinSymbol)
			rejectionMutex.Unlock()
			continue
		}
	
		// Fetch Active Down Barrier and process the values started  .....
		currentDownBarrier, err := helpers.GetCurrentDownBarrier(coinSymbol)
		debugLogs("currentDownBarrier",currentDownBarrier)
		if err!=nil{
			debugLogs("GetCurrentDownBarrier down barrier rejection Error "+coinSymbol,err)
			rejectionMutex.Unlock()
			continue
		}
		if len(currentDownBarrier) == 0 || len(currentDownBarrier[0]) == 0 {
			debugLogs("currentDownBarrier is empty "+coinSymbol)
			rejectionMutex.Unlock()
			continue
		}
		
		var current_down_barrier_value float64
		if val , ok := currentDownBarrier[0]["barier_value"].(float64); ok{
			current_down_barrier_value = val
		} else {
			debugLogs("current_down_barrier_value is missing ",coinSymbol)
			rejectionMutex.Unlock()
			continue;
		}
		var current_barrier_time time.Time
		if val , err := utils.ConvertToTime(currentDownBarrier[0]["created_date"]); err==nil{
			current_barrier_time = val
		} else {
			debugLogs("current_barrier_time is missing or has error ",coinSymbol,err)
			rejectionMutex.Unlock()
			continue;
		}
		// Fetch Active Down Barrier and process the values ended  .....

		// Fetch previous From Current Active Down Barrier and process the values started  .....
		previousActiveBarrierFromCurrent ,err := helpers.GetLastDownBarrier(coinSymbol,current_barrier_time)
		debugLogs("previousActiveBarrierFromCurrent",previousActiveBarrierFromCurrent)	
		if err!=nil{
			debugLogs("GetCurrentDownBarrier down barrier rejection Error "+coinSymbol,err)
			rejectionMutex.Unlock()
			continue
		}
		if len(previousActiveBarrierFromCurrent) == 0 || len(previousActiveBarrierFromCurrent[0]) == 0 {
			debugLogs("previousActiveBarrierFromCurrent is empty "+coinSymbol,previousActiveBarrierFromCurrent)
			rejectionMutex.Unlock()
			continue
		}
		
		var last_down_barrier_value float64
		if val , ok := previousActiveBarrierFromCurrent[0]["barier_value"].(float64); ok{
			last_down_barrier_value = val
		} else {
			debugLogs("last_down_barrier_value is missing ",coinSymbol)
			rejectionMutex.Unlock()
			continue;
		}
		var last_barrier_time time.Time
		if val , err := utils.ConvertToTime(previousActiveBarrierFromCurrent[0]["created_date"]); err==nil{
			last_barrier_time = val
		} else {
			debugLogs("last_barrier_time is missing ",coinSymbol,err)
			rejectionMutex.Unlock()
			continue;
		}	
		// Fetch previous From Current Active Down Barrier and process the values ended  .....	


		// Get the Current High To have difference between
		nextHighPoint , err := helpers.GetNextSwingPoint(coinSymbol,last_barrier_time)
		debugLogs("nextHighPoint",nextHighPoint)	
		if err!=nil{
			debugLogs("GetNextSwingPoint down barrier rejection Error "+coinSymbol,err)
			rejectionMutex.Unlock()
			continue
		}
		if len(nextHighPoint) == 0 || len(nextHighPoint[0]) == 0 {
			debugLogs("bodyMoveAverage is empty "+coinSymbol)
			rejectionMutex.Unlock()
			continue
		}
		
		var next_high_point float64
		if val , ok := nextHighPoint[0]["barier_value"].(float64); ok{
			next_high_point = val
		} else {
			debugLogs("next_high_point is missing ",coinSymbol)
			rejectionMutex.Unlock()
			continue;
		}
		var high_point_time time.Time
		if val , err := utils.ConvertToTime(nextHighPoint[0]["created_date"]); err==nil{
			high_point_time = val
		} else {
			debugLogs("high_point_time is missing ",coinSymbol)
			rejectionMutex.Unlock()
			continue;
		}
		point1_val := last_down_barrier_value
		point2_val := current_down_barrier_value
		if last_down_barrier_value > last_down_barrier_value {
			point1_val = current_down_barrier_value
			point2_val = last_down_barrier_value
		}	
		barrierObject := BarriersDataStruct{
			Symbol                 		: coinSymbol ,
			CurrentDownBarrier          : current_down_barrier_value,
			LastDownBarrier				: last_down_barrier_value,
			NextHighSwingPoint          : next_high_point,
			CurrentDownBarrierTime      : current_barrier_time,
			PreviousDownBarrierTime     : last_barrier_time,
			HighSwingTime     		    : high_point_time,
			HighPointDiff               : calculatePercentageDifference(last_down_barrier_value,next_high_point),
			LowPointDiff                : calculatePercentageDifference(point1_val,point2_val),
			Triggered					: false,
			Date						: currentHourDate,
			CurrentCandleID             : CurrentCandleID,	
			
		}

		barriersData = append(barriersData, barrierObject)
		rejectionMutex.Unlock()
	} // ends coinListCacheForBarrier forLoop
	//fmt.Println("Time Now Data",time.Now().UTC())
	//debugLogs("barriersData",barriersData)	
	for i, barrierData := range barriersData {
		//utils.PrintStructValues(barrierData);
		if hasRejectionTimeChanged(hourChangeReset) {
			debugLogs("Breaking down barrier rejection after hour change", hourChangeReset, time.Now().UTC())
			break
		}
		dataToParse := barrierData
		// params to use
		id := dataToParse.CurrentCandleID
		coin := dataToParse.Symbol
		lastDownTime := dataToParse.PreviousDownBarrierTime
		wickMovetime := endOfHourString(lastDownTime)
		isSet := dataToParse.Triggered
		if isSet{
			// already triggered for the hour..
			continue
		}
	

		
		debugLogs("currentEntry",barriersData[i])
		

		wickMove , err := helpers.GetCoinCurrentWickMove(coin,wickMovetime)
		if err!=nil{
			debugLogs("wickMove error for coin",coin)
			continue;	
		}
		if len(wickMove) == 0  || len(wickMove[0]) == 0 {
			debugLogs("wickMove Not Found For Coin",coin)
			continue;	
		}
		debugLogs("wickMove",wickMove)
		var currentWickMove float64
		currentWickMove , ok := helpers.ToFloat64(wickMove[0]["lower_wick_per_move"])
		if !ok{
			debugLogs("currentWickMove Unsupported numeric type errored")
			continue;
		} else {
			barriersData[i].Wickvalue = currentWickMove
		}
		activeDownBarrier:=  dataToParse.LastDownBarrier
		if  dataToParse.LowPointDiff >= 1 {
			activeDownBarrier =  dataToParse.CurrentDownBarrier
		}
		
		
		barriersData[i].Triggered  = true
		toUpdate := bson.M{
			"cdb_pulled"   : "yes",   // current Down Barrier Pulled.
			"cdb_value"   : dataToParse.CurrentDownBarrier,
			"cdb_last_value" : dataToParse.LastDownBarrier,
			"cdb_diff_from_high" : dataToParse.HighPointDiff,
			"cdb_diff_from_low" : dataToParse.LowPointDiff,
			"cdb_diff_from_low_value" : activeDownBarrier,
			"cdb_swing_time" : dataToParse.HighSwingTime,
			"cdb_low_time" : dataToParse.CurrentDownBarrierTime,
			"cdb_last_low_time" : dataToParse.PreviousDownBarrierTime,
			"cdb_pulled_time": time.Now().UTC(),
			"cdb_wick_value" : currentWickMove,
			"cdb_high_value" : dataToParse.NextHighSwingPoint,
			
		}
		debugLogs("data",toUpdate)
		if(debug){
			//return
		}

		err = helpers.UpdateDownBarrierRejectionData(id,toUpdate)
		if err!=nil{
			continue;
		}
	
		
	} // ends barriersData loop
}

func calculatePercentageDifference(lastDownBarrier, nextHighSwingPoint float64) float64 {
	percentageDifference := (nextHighSwingPoint - lastDownBarrier) / lastDownBarrier * 100
	return percentageDifference
}


func hasRejectionTimeChanged(hourChangeReset time.Time) bool {
	return time.Now().UTC().Hour() != hourChangeReset.Hour()
}



func findBarriersObject(coinSymbol string, currentHourDate time.Time) *BarriersDataStruct {
	for i := range barriersData {
		barrierData := &barriersData[i]
		objectDate := barrierData.Date
		if barrierData.Symbol == coinSymbol && objectDate.Equal(currentHourDate) {
			return barrierData
		}
	}
	return nil
}


func GetHourStartTime(dateNow time.Time) time.Time {
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


func debugLogs(msg string, params ...interface{}) {
	if !debug{
		return
	}
	fmt.Print(msg)

	if len(params) > 0 {
		fmt.Print(" Params:")
		for i, param := range params {
			fmt.Printf(" %d:%v", i+1, param)
		}
	}

	fmt.Println()
}

func endOfHourString(inputTime time.Time) string {
	// Set the time to the beginning of the next hour and subtract one second
	endOfHourTime := inputTime.Add(time.Hour - time.Second)
	
	// Format the time as a string
	return endOfHourTime.Format("2006-01-02 15:04:05")
}