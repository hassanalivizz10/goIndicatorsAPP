package cron

import (
    //"fmt"
    "github.com/robfig/cron/v3"
    "indicatorsAPP/controller"
)

func StartCronJobs() {
    c := cron.New()

    // Schedule your cron jobs
    _, _ = c.AddFunc("@every 20s", controller.RunBigRaiseAndBigDrop)  // Run every hour at minute 0
    // Add more cron jobs here

    c.Start()
}


