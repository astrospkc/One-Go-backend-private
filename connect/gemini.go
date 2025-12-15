package connect

import (
	"context"
	"fmt"
	"gobackend/env"
	"log"

	"google.golang.org/genai"
)

var GeminiClient *genai.Client

func InitGemini() {
	envs := env.NewEnv()
	gemini_api_key := envs.GEMINI_API_KEY

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  gemini_api_key,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("GeminiClient initialized")
	GeminiClient = client
}
