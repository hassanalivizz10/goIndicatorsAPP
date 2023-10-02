package cron

import (
    //"fmt"
    "github.com/robfig/cron/v3"
    "indicatorsAPP/controller"
)

func StartCronJobs() {
    c := cron.New()

    // Schedule your cron jobs
    //_, _ = c.AddFunc("@every 20s", controller.RunBigRaiseAndBigDrop)  // every 20seconds
   // _, _ = c.AddFunc("@every 20s", controller.BigRaisePullBack) // every 20seconds
    //_, _ = c.AddFunc("@every 20s", controller.BigDropPullBack)  // every 20seconds
    _, _ = c.AddFunc("*/1 * * * *", controller.StrategyCron)  // every hour 8 minute
   // _ , _ = c.AddFunc("0 * * * *", controller.RecycleTheTrack)   // every hour start
    // Add more cron jobs here
    c.Start()
}


