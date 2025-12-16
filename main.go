package main

import (
	"fmt"
	"gobackend/config"
	"gobackend/connect"
	"gobackend/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

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

	app.Listen(":8080")
}
