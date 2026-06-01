package config

import(
	"log"
	"os"
	"strconv"

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
	CloudinaryCloudName    string
	CloudinaryAPIKey       string
	CloudinaryAPISecret    string
	CloudinaryUploadFolder string
	GNewsAPIKey  		   string
	GNewsEnabled 		   string
	AlertSyncEnabled             string
	AlertSyncInitialDelaySeconds int
	AlertSyncIntervalMinutes     int
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
		CloudinaryCloudName:    getEnv("CLOUDINARY_CLOUD_NAME", ""),
		CloudinaryAPIKey:       getEnv("CLOUDINARY_API_KEY", ""),
		CloudinaryAPISecret:    getEnv("CLOUDINARY_API_SECRET", ""),
		CloudinaryUploadFolder: getEnv("CLOUDINARY_UPLOAD_FOLDER", "disaster_alert/reports"),
		GNewsAPIKey: 			getEnv("GNEWS_API_KEY", ""),
		GNewsEnabled: 			getEnv("GNEWS_ENABLED", ""),
		AlertSyncEnabled:             getEnv("ALERT_SYNC_ENABLED", "true"),
		AlertSyncInitialDelaySeconds: getEnvAsInt("ALERT_SYNC_INITIAL_DELAY_SECONDS", 120),
		AlertSyncIntervalMinutes:     getEnvAsInt("ALERT_SYNC_INTERVAL_MINUTES", 30),
		
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)

	if value == "" {
		return fallback
	}

	return value
}

func getEnvAsInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return parsed
}
