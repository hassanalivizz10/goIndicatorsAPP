package main

import (
	"fmt"
   "indicatorsAPP/apiroutes"
    "indicatorsAPP/cron"
	"indicatorsAPP/config"
    "log"
	"runtime"
	"os"
    "github.com/joho/godotenv"
)

func main() {
	//fmt.Println("HEELO THERE")	
    defer func() {
		if r := recover(); r != nil {
			// Get the stack trace information
			buf := make([]byte, 4096)
			n := runtime.Stack(buf, false)

			// Print the error details, including file, line number, and function
			fmt.Println("Panic occurred:")
			fmt.Println(string(buf[:n]))

			// Optionally, you can log the error or perform other cleanup tasks

			// Exit the program with a non-zero exit code to indicate failure
			os.Exit(1)
		}
	}()

	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Connect to MongoDB
	err = config.ConnectToMongoDB()
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
    // Start HTTP server with API routes
    go apiroutes.StartServer()

    // Start cron jobs
    cron.StartCronJobs()

    // Block indefinitely
    select {}
}
