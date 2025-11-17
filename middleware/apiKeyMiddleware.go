package middleware

import (
	"context"
	"gobackend/connect"
	"gobackend/models"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// TODO: APi key can be more secured if signature is added

func ValidateAPIKey() fiber.Handler {
	return func(c *fiber.Ctx) error{

	apikey := c.Get("X-API-Key")

	filter := bson.M{
		"key":apikey,
	}
	var u models.APIkey
	err := connect.APIKeyCollection.FindOne(context.TODO(), filter).Decode(&u)
	if err!=nil{
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":"no user found with this apikey",
		})
	}

	c.Locals("user_id", u.UserId)
	return c.Next()
}

}