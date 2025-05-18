package env

import (
	"log"
	"fmt"
	"github.com/spf13/viper"
)

type ENV struct{
	MONGODB_URI                 string `mapstructure:"MONGODB_URI"`
	JWT_SECRET					string `mapstructure:"JWT_SECRET"`
	S3_BUCKET_NAME				string `mapstructure:"S3_BUCKET_NAME"`
	AWS_SECRET_ACCESS_KEY		string `mapstructure:"AWS_SECRET_ACCESS_KEY"` 
	AWS_ACCESS_KEY_ID			string `mapstructure:"AWS_ACCESS_KEY_ID"`
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
<<<<<<< HEAD
=======
	// fmt.Println(env)
>>>>>>> 778119e743aa12f39c6e30e01059e9b69b5fb929
   
	return &env
   }
