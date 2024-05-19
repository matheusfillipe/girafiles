package api

import (
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/joho/godotenv"
)

type Settings struct {
	// Name to display in the HTML
	AppName string
	// Host the application Listen to
	Host string
	// Port the application Listen to
	Port string
	// Debug mode
	Debug bool
	// Path to store the files
	StorePath string
	// File persistance time in hours. 0 to keep files forever
	FilePersistanceTime int
	// File size limit in MB
	FileSizeLimit int
	// Size limit for STORE_PATH in MB. If exceeded, the oldest files will be deleted first. Set to 0 to disable
	StorePathSizeLimit int
	// Users and Passwords. Leave it empty to disable authentication. Format: user1:password1,user2:password2
	Users map[string]string
	// IP Rate Limit per minute. 0 to disable
	IPMinRateLimit int
	// IP Rate Limit per hour. 0 to disable
	IPHourRateLimit int
	// IP Rate Limit per day. 0 to disable
	IPDayRateLimit int
}

var singleInstance *Settings

func getDefaultSettings() *Settings {
	return &Settings{
		AppName:             "GiraFiles",
		Host:                "0.0.0.0",
		Port:                "8000",
		Debug:               false,
		StorePath:           "/tmp/girafiles/data",
		FilePersistanceTime: 0,
		FileSizeLimit:       100,
		StorePathSizeLimit:  2048,
		Users:               map[string]string{},
		IPMinRateLimit:      0,
		IPHourRateLimit:     0,
		IPDayRateLimit:      0,
	}
}

func (s *Settings) IsAuthEnabled() bool {
	return len(s.Users) > 0
}

func (s *Settings) IsIPRateLimitEnabled() bool {
	return s.IPMinRateLimit > 0 || s.IPHourRateLimit > 0 || s.IPDayRateLimit > 0
}

func (s *Settings) IsStorePathSizeLimitEnabled() bool {
	return s.StorePathSizeLimit > 0
}

func parseAuthUsers(users string) map[string]string {
	var usersMap = map[string]string{}
	if users != "" {
		for _, user := range strings.Split(users, ",") {
			var userArray = strings.Split(user, ":")
			if len(userArray) != 2 {
				log.Fatalf("Error parsing 'USERS'. Expected a list of users and passwords separated by comma and colon but got '%s'", users)
			}
			usersMap[userArray[0]] = userArray[1]
		}
	}
	return usersMap
}

func getEnv(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		var intValue, err = strconv.Atoi(value)
		if err != nil {
			log.Fatalf("Error parsing '%s'. Expected a integer but got '%s'", key, value)
		}
		return intValue
	}
	return defaultValue
}

func newSettings() *Settings {
	var settings = getDefaultSettings()
	err := godotenv.Load()
	if err != nil {
		if err.Error() == "open .env: no such file or directory" {
			log.Println("No .env file found. Using default settings and environment variables.")
		} else {
			log.Fatal("Error loading .env file")
		}
	}

	settings = &Settings{
		AppName:             getEnv("APP_NAME", settings.AppName),
		Host:                getEnv("HOST", settings.Host),
		Port:                getEnv("PORT", settings.Port),
		Debug:               getIntEnv("DEBUG", 0) == 1,
		StorePath:           getEnv("STORE_PATH", settings.StorePath),
		FilePersistanceTime: getIntEnv("FILE_PERSISTANCE_TIME", settings.FilePersistanceTime),
		FileSizeLimit:       getIntEnv("FILE_SIZE_LIMIT", settings.FileSizeLimit),
		StorePathSizeLimit:  getIntEnv("STORE_PATH_SIZE_LIMIT", settings.StorePathSizeLimit),
		Users:               parseAuthUsers(getEnv("USERS", "")),
		IPMinRateLimit:      getIntEnv("IP_MIN_RATE_LIMIT", settings.IPMinRateLimit),
		IPHourRateLimit:     getIntEnv("IP_HOUR_RATE_LIMIT", settings.IPHourRateLimit),
		IPDayRateLimit:      getIntEnv("IP_DAY_RATE_LIMIT", settings.IPDayRateLimit),
	}

	// mkdir -p STORE_PATH
	if _, err := os.Stat(settings.StorePath); os.IsNotExist(err) {
		if os.MkdirAll(settings.StorePath, os.ModePerm) != nil {
			log.Fatalf("Error creating directory '%s'", settings.StorePath)
		}
	}

	// Check if STORE_PATH is a directory
	if fileInfo, err := os.Stat(settings.StorePath); err != nil || !fileInfo.IsDir() {
		log.Fatalf("'%s' is not a directory", settings.StorePath)
	}

	return settings
}

var lock = &sync.Mutex{}

func GetSettings() *Settings {
	if singleInstance == nil {
		lock.Lock()
		defer lock.Unlock()
		if singleInstance == nil {
			singleInstance = newSettings()
		}
	}
	return singleInstance
}
