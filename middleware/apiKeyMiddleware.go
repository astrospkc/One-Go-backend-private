package middleware

import (
	"context"
	"gobackend/connect"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func ValidateAPIKey(c *fiber.Ctx) bool {
	apikey := c.Get("API-Key")
	user := c.Locals("user")
	claims,ok := user.(jwt.MapClaims)
	if !ok {
		return false
	}
	user_id, ok := claims["aud"].(string)
	if !ok {
		return false
	}
	
	
	filter := bson.M{
		"user_id":user_id,
		"key":apikey,
	}
	count,err := connect.APIKeyCollection.CountDocuments(context.TODO(), filter)
	if err != nil {
		log.Println("Error while checking API key:", err)
		return false // or handle error
	}
	if count == 0 {
	// No match found: API key doesn't belong to this user
		return false
	}
	return true
	
}