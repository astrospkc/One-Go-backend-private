package env

import (
	"log"

	"github.com/spf13/viper"
)

type ENV struct {
	MONGODB_URI           string `mapstructure:"MONGODB_URI"`
	JWT_SECRET            string `mapstructure:"JWT_SECRET"`
	S3_BUCKET_NAME        string `mapstructure:"S3_BUCKET_NAME"`
	AWS_SECRET_ACCESS_KEY string `mapstructure:"AWS_SECRET_ACCESS_KEY"`
	AWS_ACCESS_KEY_ID     string `mapstructure:"AWS_ACCESS_KEY_ID"`
	RESEND_API_KEY        string `mapstructure:"RESEND_API_KEY"`
	AIVEN_KEY             string `mapstructure:"AIVEN_KEY"`
	AIVEN_USERNAME        string `mapstructure:"AIVEN_USERNAME"`
	AIVEN_PASSWORD        string `mapstructure:"AIVEN_PASSWORD"`
	AIVEN_HOST            string `mapstructure:"AIVEN_HOST"`
	AIVEN_PORT            int    `mapstructure:"AIVEN_PORT"`
	RAZORPAY_KEY_ID       string `mapstructure:"RAZORPAY_KEY_ID"`
	RAZORPAY_KEY_SECRET   string `mapstructure:"RAZORPAY_KEY_SECRET"`
	RAZORPAY_WEBHOOK_SECRET string `mapstructure:"RAZORPAY_WEBHOOK_SECRET"`
	GEMINI_API_KEY        string `mapstructure:"GEMINI_API_KEY"`
}

func NewEnv() *ENV {
	env := ENV{}
	viper.SetConfigFile(".env")

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal("Can't find the file .env : ", err)
	}

	err = viper.Unmarshal(&env)
	if err != nil {
		log.Fatal("Environment can't be loaded: ", err)
	}

	return &env
}
