package utils

import (
	
	"fmt"
	"errors"
	"strings"
	"os"
	"log"
	"net"
	"net/http"
	"time"
	"github.com/joho/godotenv"
	
	"go.mongodb.org/mongo-driver/bson/primitive"
	
)
// initialize
func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

// Read ENV Var by the Key.
func GetEnv(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Fatalf("Environment variable %s not found", key)
	}
	return value
}
func GetTimeStamp() int64 {
	timestamp := time.Now().UTC().Unix()
	return timestamp
}

// Validate the mongo Object ID, that was sent as string.
func IsValidObjectID(id string) bool {
	_, err := primitive.ObjectIDFromHex(id)
	return err == nil
}
// Value in array
func FieldExistsInArray(arr []string, target string) bool {
	return strings.Contains(strings.Join(arr, ","), target)
}


// GetClientIP extracts the client IP address from the request
func GetClientIP(req *http.Request) string {
	// Check X-Real-IP and X-Forwarded-For headers in case of reverse proxy or load balancer
	ip := req.Header.Get("X-Real-IP")
	if ip == "" {
		ip = req.Header.Get("X-Forwarded-For")
	}
	
	// If headers are not available, extract the IP address from req.RemoteAddr
	if ip == "" {
		host, _, err := net.SplitHostPort(req.RemoteAddr)
	
		if err != nil {
			return ""
		}
		if host == "::1" {
			host = "127.0.0.1" // Replace IPv6 loopback address with IPv4 loopback address
		}
		ip = host
		
	}

	return ip
}
// converting a string number to Integer
func CustomStringToNumber(s string) (int, error) {
	var value int
	var sign int = 1
	var started bool

	for _, ch := range s {
		if ch >= '0' && ch <= '9' {
			digit := int(ch - '0')
			value = value*10 + digit
			started = true
		} else if ch == '-' && !started {
			sign = -1
		} else {
			return 0, errors.New("invalid input: unable to convert string to number")
		}
	}

	return value * sign, nil
}

// ToFloat
func ToFloat(value interface{}) (float64, error) {
	switch v := value.(type) {
	case string:
		num, err := parseStringToFloat(v)
		if err != nil {
			return 0, err
		}
		return num, nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	default:
		return 0, fmt.Errorf("unable to convert to float: unsupported type %T", value)
	}
}

func parseStringToFloat(str string) (float64, error) {
	var num float64
	_, err := fmt.Sscanf(str, "%f", &num)
	if err != nil {
		return 0, err
	}
	return num, nil
}

// Time Ago.

func FormatTimeAgo(from time.Time, to time.Time) string {
	// Calculate the duration between the two time values
	duration := from.Sub(to)

	// Convert the duration to a positive value
	if duration < 0 {
		duration = -duration
	}

	// Determine the appropriate time unit and value
	var unit string
	var value int

	switch {
	case duration.Seconds() < 60:
		unit = "second"
		value = int(duration.Seconds())
	case duration.Minutes() < 60:
		unit = "minute"
		value = int(duration.Minutes())
	case duration.Hours() < 24:
		unit = "hour"
		value = int(duration.Hours())
	default:
		unit = "day"
		value = int(duration.Hours() / 24)
	}

	// Handle pluralization
	if value > 1 {
		unit += "s"
	}

	// Return the formatted result
	return fmt.Sprintf("%d %s ago", value, unit)
}