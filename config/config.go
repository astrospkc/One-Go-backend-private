package config

import "gobackend/env"

type Config struct {
	RazorpayKeyId     string
	RazorpayKeySecret string
}

func LoadConfig() *Config {
	envs := env.NewEnv()
	return &Config{
		RazorpayKeyId:     envs.RAZORPAY_KEY_ID,
		RazorpayKeySecret: envs.RAZORPAY_KEY_SECRET,
	}
}