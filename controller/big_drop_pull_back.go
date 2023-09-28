package controller
import (
	"fmt"
	"sync"
	"time"
	"indicatorsAPP/helpers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)
var bigDropPullBackMutex sync.Mutex
func BigDropPullBack(){
 dropData , err := helpers.FetchRaiseORDropEmptyPullBackTime("big_drop_pull_back")
 if err!=nil{
	fmt.Println("ERROR ON BIG DROP PULL BACK",err.Error())
	return
 }
 if len(dropData) ==0 {
	fmt.Println("No BigDrop Pull Back")
	return
 }
 for _, current := range dropData {
	bigDropPullBackMutex.Lock()
	id                     := current["_id"].(primitive.ObjectID);
	coinSymbol             := current["symbol"].(string)
    pull_back_price        := current["pull_back_price"].(float64)
	pull_back_created_date := current["created_date"].(time.Time)
	open_price             := current["open_price"].(float64)
	drop_price             := current["drop_price"].(float64)
	trailing_price         := current["trailing_price"].(float64)
	move                   := current["move"].(float64)
	created_date           := time.Now()
	
	pulled_back , err := helpers.GetPullBackPrice(coinSymbol,pull_back_price,pull_back_created_date,false)
	if err!=nil{
		fmt.Println("pulled_back error for coin"+coinSymbol,err)
		bigDropPullBackMutex.Unlock()
		continue
	}
	
	if len(pulled_back) == 0{
		fmt.Println("pulled_back Not Found for coin")
		bigDropPullBackMutex.Unlock()
		continue
	}

	triggered_price , ok := helpers.ToFloat64(pulled_back[0]["price"])
	if !ok{
		fmt.Println("pulled_back currentPrice Unsupported numeric type errored")
		bigDropPullBackMutex.Unlock()
		continue
	}


	err = helpers.UpdatePullBackTime(id)
	if err!=nil{
		fmt.Println("UpdatePullBackTime error for coin"+coinSymbol,err)
		bigDropPullBackMutex.Unlock()
		continue
	}
	err = helpers.UpdateMarketTrendingEntry(coinSymbol,false)
	if err!=nil{
		fmt.Println("UpdateMarketTrending error for coin"+coinSymbol,err)
		bigDropPullBackMutex.Unlock()
		continue
	}
	insertData := bson.M{
		"coin":coinSymbol,
		"type":"big_drop_pull_back",
		"trailing_price":trailing_price,
		"drop_price":drop_price,
		"move":move,
		"open_price":open_price,
		"pull_back_price":pull_back_price,
		"triggered_price":triggered_price,
		"pull_back_time":pull_back_created_date,
		"created_date":created_date,
	}
	err = helpers.AddCandleTrack(insertData)
	if err!=nil{
		fmt.Println("UpdateMarketTrending error for coin"+coinSymbol,err)
		bigDropPullBackMutex.Unlock()
		continue
	}
	bigDropPullBackMutex.Unlock()	
 } // ends dropData for loop

}