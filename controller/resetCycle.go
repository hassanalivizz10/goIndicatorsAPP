package controller
import (
	"fmt"
	"indicatorsAPP/mongohelpers"
	"go.mongodb.org/mongo-driver/bson"
	
)

func RecycleTheTrack(){
// Update Trending Collection and reset the drop raise indicators	
 trendingFilters := bson.M{}
 trendingCollection:= "market_trending"
 updateTrending:=bson.M{
	"$set":bson.M{
		"big_raise_pull_back":"no",
		"big_drop_pull_back":"no",
	},
 }
 err := mongohelpers.MongoUpdateMany(trendingCollection,trendingFilters,updateTrending,false)
 if err!=nil{
	fmt.Println("RecycleTheTrack Error on UpdateMany",err.Error())
 }
 // Delete the track for raise and pull
 trackCollectionName := "big_drop_and_pull_back_track"
 filters := bson.M{}
 err = mongohelpers.MongoDeleteMany(trackCollectionName,filters)
 if err!=nil{
	fmt.Println("RecycleTheTrack Error on DeleteMany",err.Error())
 }
 //
}