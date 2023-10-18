package cron

import (
    //"fmt"
    "github.com/robfig/cron/v3"
    "indicatorsAPP/controller"
)

func StartCronJobs() {
    c := cron.New()
  //   c.AddFunc("@every 20s", func() {
	// 	fmt.Println("Hello, World! This runs every 20 seconds.")
	// })
  
    // Schedule your cron jobs


    _, _ = c.AddFunc("@every 5s", controller.RunBigRaiseAndBigDrop)  // every 20seconds
    _, _ = c.AddFunc("@every 10s", controller.BigRaisePullBack) // every 20seconds
   _, _ = c.AddFunc("@every 10s", controller.BigDropPullBack)  // every 20seconds
    _, _ = c.AddFunc("12 * * * *", controller.StrategyCron)  // every hour 12th minute
    _ , _ = c.AddFunc("0 * * * *", controller.RecycleTheTrack)   // every hour start
 
 
   //   Add more cron jobs above
  c.Start()
}


