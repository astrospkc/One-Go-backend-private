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
		AllowOrigins:"http://localhost:3000 ,https://onego.xastrosbuild.site",
		AllowMethods: "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
	}))

	connect.Connect()

	fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")

	routes.RegisterNormalRoutes(app)
	routes.RegisterAPIKeyRoutes(app)

	

	app.Listen(":8000")
}
