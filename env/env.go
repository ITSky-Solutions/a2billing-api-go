package env

import (
	"log"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
)

func LoadDotEnv[T interface{}](config *T) {
	if err := godotenv.Load(); err != nil {
		log.Fatalln("Failed to load environment variables:", err)
	}

	if err := env.Parse(config); err != nil {
		log.Fatalln(err)
	}
}

type _Env struct {
	DbUser     string `env:"API_DB_USER,required"`
	DbPassword string `env:"API_DB_PASS"`
	DbName     string `env:"API_DB_NAME,required"`
	DbHost     string `env:"API_DB_HOST,required"`
	DbPort     string `env:"API_DB_PORT"`

	ApiPort string `env:"API_PORT"`
}

var Env _Env

func init() {
	LoadDotEnv(&Env)

	if len(Env.DbPort) == 0 {
		Env.DbPort = "3306"
	}

	if len(Env.ApiPort) == 0 {
		Env.ApiPort = "8080"
	}
}
