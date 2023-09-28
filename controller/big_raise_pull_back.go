package controller
import (
	"fmt"
	"sync"
	"time"
	"indicatorsAPP/helpers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)
var bigRaisePullBackMutex sync.Mutex
func BigRaisePullBack(){
 dropData , err := helpers.FetchRaiseORDropEmptyPullBackTime("big_raise_pull_back")
 if err!=nil{
	fmt.Println("ERROR ON BIG Raise PULL BACK",err.Error())
	return
 }
 if len(dropData) ==0 {
	fmt.Println("No BigRaise Pull Back")
	return
 }
 for _, current := range dropData {
	bigRaisePullBackMutex.Lock()
	var pull_back_created_date time.Time
	date_created , ok := current["created_date"].(time.Time)
	if ok {
		pull_back_created_date =  date_created
	}
	createdDatePrimitive, ok := current["created_date"].(primitive.DateTime)
	if ok {
		pull_back_created_date = time.Unix(0, int64(createdDatePrimitive)*int64(time.Millisecond)).UTC()
	} 
	
	fmt.Println("pull_back_created_date",pull_back_created_date)
	id                     := current["_id"].(primitive.ObjectID);
	coinSymbol             := current["coin"].(string)
    pull_back_price        := current["pull_back_price"].(float64)
	
	open_price             := current["open_price"].(float64)
	raise_price            := current["raise_price"].(float64)
	trailing_price         := current["trailing_price"].(float64)
	move                   := current["move"].(float64)
	created_date           := time.Now()
	
	raised_back , err := helpers.GetPullBackPrice(coinSymbol,pull_back_price,pull_back_created_date,true)
	if err!=nil{
		fmt.Println("raised_back error for coin"+coinSymbol,err)
		bigRaisePullBackMutex.Unlock()
		continue
	}
	fmt.Println("raised_back",raised_back)
	if len(raised_back) == 0{
		fmt.Println("raised_back Not Found for coin")
		bigRaisePullBackMutex.Unlock()
		continue
	}

	triggered_price , ok := helpers.ToFloat64(raised_back[0]["price"])
	if !ok{
		fmt.Println("raised_back currentPrice Unsupported numeric type errored")
		bigRaisePullBackMutex.Unlock()
		continue
	}


	err = helpers.UpdatePullBackTime(id)
	if err!=nil{
		fmt.Println("UpdatePullBackTime error for coin"+coinSymbol,err)
		bigRaisePullBackMutex.Unlock()
		continue
	}
	err = helpers.UpdateMarketTrendingEntry(coinSymbol,true)
	if err!=nil{
		fmt.Println("UpdateMarketTrending error for coin"+coinSymbol,err)
		bigRaisePullBackMutex.Unlock()
		continue
	}
	insertData := bson.M{
		"coin":coinSymbol,
		"type":"big_raise_pull_back",
		"trailing_price":trailing_price,
		"raise_price":raise_price,
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
		bigRaisePullBackMutex.Unlock()
		continue
	}
	bigRaisePullBackMutex.Unlock()	
 } // ends dropData for loop

}