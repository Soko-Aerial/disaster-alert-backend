package config

import(
	"log"
	"os"

	"github.com/joho/godotenv"
)


type Config struct{
	AppName string
	AppEnv string
	Port string
	MongoURI string
	MongoDatabase string
	JWTSecret string
	JWTExpiresIn string
	FirebaseCredentialsFile string
	FirebaseCredentialsJSON string
	NASAFIRMSMapKey   string
	NASAFIRMSSource   string
	NASAFIRMSArea     string
	NASAFIRMSDayRange string
	NASAFIRMSLimit string
	OpenWeatherAPIKey string
	ReliefWebAppName string
	ReliefWebLimit   string
	GDELTLimit    string
	GDELTTimespan string
	ReliefWebEnabled string
	GDACSEnabled       string
	NASAFIRMSEnabled   string
	GDELTEnabled       string
	OpenWeatherEnabled string
}	

func LoadConfig() *Config {
	err := godotenv.Load()

	if err != nil {
		log.Println("No .env file found. Using system environment variables.")
	}

	return &Config{
		AppName:       			 getEnv("APP_NAME", "Disaster Alert"),
		AppEnv:        			 getEnv("APP_ENV", "development"),
		Port:          			 getEnv("PORT", "8080"),
		MongoURI:      			 getEnv("MONGO_URI", ""),
		MongoDatabase: 			 getEnv("MONGO_DATABASE", "disaster_alert"),
		JWTSecret:     			 getEnv("JWT_SECRET", ""),
		JWTExpiresIn:  			 getEnv("JWT_EXPIRES_IN", ""),
		FirebaseCredentialsFile: getEnv("FIREBASE_CREDENTIALS_FILE", ""),
		FirebaseCredentialsJSON: getEnv("FIREBASE_CREDENTIALS_JSON", ""),
		NASAFIRMSMapKey:   		 getEnv("NASA_FIRMS_MAP_KEY", ""),
		NASAFIRMSSource:   		 getEnv("NASA_FIRMS_SOURCE", ""),
		NASAFIRMSArea:     		 getEnv("NASA_FIRMS_AREA", ""),
		NASAFIRMSDayRange: 		 getEnv("NASA_FIRMS_DAY_RANGE", ""),
		NASAFIRMSLimit: 		 getEnv("NASA_FIRMS_LIMIT", ""),
		OpenWeatherAPIKey: 		 getEnv("OPENWEATHER_API_KEY", ""),
		ReliefWebAppName:  	     getEnv("RELIEFWEB_APP_NAME", ""),
		ReliefWebLimit:    		 getEnv("RELIEFWEB_LIMIT", ""),
		GDELTLimit:    		     getEnv("GDELT_LIMIT", ""),
		GDELTTimespan:    		 getEnv("GDELT_TIMESPAN", ""),
		ReliefWebEnabled: 		 getEnv("RELIEFWEB_ENABLED", ""),
		GDACSEnabled:       	 getEnv("GDACS_ENABLED", ""),
		NASAFIRMSEnabled:   	 getEnv("NASA_FIRMS_ENABLED", ""),
		GDELTEnabled:       	 getEnv("GDELT_ENABLED", ""),
		OpenWeatherEnabled: 	 getEnv("OPENWEATHER_ENABLED", ""),
		
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)

	if value == "" {
		return fallback
	}

	return value
}
