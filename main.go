package main

import (
	"context"
	"encoding/json"
	"fmt"
	"gobackend/config"
	"gobackend/connect"
	"gobackend/routes"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"google.golang.org/genai"
)

// https://one-go-private.vercel.app
var Cfg *config.Config

func main() {
	app := fiber.New()
	Cfg = config.LoadConfig()

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000, https://one-go-private.vercel.app",
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "*",
		AllowCredentials: true,
	}))
	app.Server().MaxRequestBodySize = 50 * 1024 * 1024
	connect.Connect()

	fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")
	connect.InitGemini()
	connect.InitRedis()

	// Example: Set & Get
	connect.RedisClient.Do(connect.Rctx,
		connect.RedisClient.B().Set().Key("key").Value("hello").Build(),
	)

	val, _ := connect.RedisClient.Do(connect.Rctx,
		connect.RedisClient.B().Get().Key("key").Build(),
	).ToString()

	println("Value:", val)

	routes.RegisterNormalRoutes(app)
	routes.RegisterAPIKeyRoutes(app)

	// ---------------------------------------------
	// Gemini testing
	chat, err := connect.GeminiClient.Chats.Create(context.Background(), "gemini-2.5-flash", nil, nil)
	if err != nil {
		log.Fatal(err)
	}

	result, err := chat.SendMessage(context.Background(), genai.Part{Text: "What's the weather in New York?"})
	if err != nil {
		log.Fatal(err)
	}
	debugPrint(result)

	result, err = chat.SendMessage(context.Background(), genai.Part{Text: "How about San Francisco?"})
	if err != nil {
		log.Fatal(err)
	}
	debugPrint(result)
	// ------------------------------------------

	app.Listen(":8080")
}

func debugPrint[T any](r *T) {

	response, err := json.MarshalIndent(*r, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(response))
}
