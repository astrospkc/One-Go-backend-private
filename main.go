package main

import (
	"fmt"
	"gobackend/connect"
	"gobackend/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "https://one-go-private.vercel.app", // you can also just allow this for production
		AllowMethods: "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders: "*",
		AllowCredentials: true,
		AllowOriginsFunc: func(origin string) bool {
		return origin == "http://localhost:3000" || origin == "https://one-go-private.vercel.app"
	},
	}))

	connect.Connect()

	fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")

	routes.RegisterNormalRoutes(app)
	routes.RegisterAPIKeyRoutes(app)

	

	app.Listen(":8000")
}
