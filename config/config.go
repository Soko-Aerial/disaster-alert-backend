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
		MongoURI:      			 getEnv("MONGO_URI", "mongodb+srv://arhinfulemmanuel294_db_user:Bill2704arhin@disasteralerts.fd5kx3l.mongodb.net/?appName=disasteralerts"),
		MongoDatabase: 			 getEnv("MONGO_DATABASE", "disaster_alert"),
		JWTSecret:     			 getEnv("JWT_SECRET", "WM9lFwk8rCdXNSacG5Z40s1IhOfvymV2iPzLJUEjT3xtqRAunHpgKoBb6Ye7DQ"),
		JWTExpiresIn:  			 getEnv("JWT_EXPIRES_IN", "24h"),
		FirebaseCredentialsFile: getEnv("FIREBASE_CREDENTIALS_FILE", "./firebase/service-account.json"),
		FirebaseCredentialsJSON: getEnv("FIREBASE_CREDENTIALS_JSON", ""),
		NASAFIRMSMapKey:   		 getEnv("NASA_FIRMS_MAP_KEY", "e49b6917e0acbceadd7bdafa36e0394f"),
		NASAFIRMSSource:   		 getEnv("NASA_FIRMS_SOURCE", "VIIRS_SNPP_NRT"),
		NASAFIRMSArea:     		 getEnv("NASA_FIRMS_AREA", "-20,-5,25,25"),
		NASAFIRMSDayRange: 		 getEnv("NASA_FIRMS_DAY_RANGE", "1"),
		NASAFIRMSLimit: 		 getEnv("NASA_FIRMS_LIMIT", "50"),
		OpenWeatherAPIKey: 		 getEnv("OPENWEATHER_API_KEY", "ac26e42b1a0c434dbfa11a5b355f5b11"),
		ReliefWebAppName:  	     getEnv("RELIEFWEB_APP_NAME", "disaster-alert-backend"),
		ReliefWebLimit:    		 getEnv("RELIEFWEB_LIMIT", "20"),
		GDELTLimit:    		     getEnv("GDELT_LIMIT", "20"),
		GDELTTimespan:    		 getEnv("GDELT_TIMESPAN", "24h"),
		ReliefWebEnabled: 		 getEnv("RELIEFWEB_ENABLED", "false"),
		GDACSEnabled:       	 getEnv("GDACS_ENABLED", "true"),
		NASAFIRMSEnabled:   	 getEnv("NASA_FIRMS_ENABLED", "true"),
		GDELTEnabled:       	 getEnv("GDELT_ENABLED", "true"),
		OpenWeatherEnabled: 	 getEnv("OPENWEATHER_ENABLED", "true"),
		
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)

	if value == "" {
		return fallback
	}

	return value
}
