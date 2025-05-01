// Connects to MongoDB and sets a Stable API version
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
	// allowedOrigins :=  "http://localhost:3000, http://192.168.120.145:3000"
	app.Use(cors.New(cors.Config{
        // AllowHeaders:     "Origin,Content-Type,Accept,Content-Length,Accept-Language,Accept-Encoding,Connection,Access-Control-Allow-Origin",
        AllowOrigins:    "http://localhost:3000",
    
        AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
    }))
	connect.Connect()
	
	fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")

	routes.RegisterNormalRoutes(app)
	routes.RegisterAPIKeyRoutes(app)

	url, err := services.CreatePresignedUrlAndUploadObject("cms-one-go", "img.jpg")

	if err != nil {
		log.Fatalf("Failed to generate URL: %v", err)
	}
	fmt.Println("Presigned URL:", url)

	// geturl, err := services.GetPresignedGetUrl("cms-one-go", "image.jpg")

	// if err != nil {
	// 	log.Fatalf("Failed to generate URL: %v", err)
	// }
	// fmt.Println("Presigned URL:", geturl)





	// app.Post("/api/createUser", controller.CreateUser)
	app.Listen(":8080")

	
}
