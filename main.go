package main

import (
	"fmt"
	"gobackend/connect"
	"gobackend/routes"
	"gobackend/services"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:3000",
		AllowMethods: "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
	}))

	connect.Connect()

	fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")

	routes.RegisterNormalRoutes(app)
	routes.RegisterAPIKeyRoutes(app)

	_, err := services.CreatePresignedUrlAndUploadObject("cms-one-go", "img.jpg")
	if err != nil {
		log.Fatalf("Failed to generate URL: %v", err)
	}
	app.Listen(":8000")
}
