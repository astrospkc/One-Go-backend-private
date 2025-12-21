package middleware

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"gobackend/models"
	"gobackend/connect"
	
	
)

func IsSubscribed() fiber.Handler{
	return func(c *fiber.Ctx) error{
		user_id := c.Locals("user_id").(string)
		
		if user_id==""{
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message":"Unauthorized",
			})
		}
		filter := bson.M{
			"user_id": user_id,
			"status": bson.M{
				"$in": []string{"active", "pending"},
			},
		}
		var subscriptionList []models.Subscription
		cursor, err := connect.SubscriptionCollection.Find(context.TODO(), filter)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Unauthorized",
			})
		}
		if err = cursor.All(context.TODO(), &subscriptionList); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Unauthorized",
			})
		}
		if len(subscriptionList) == 0{
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message":"Unauthorized",
			})
		}

		for _,subscription := range subscriptionList{
			if subscription.Plan != "starter" && subscription.Status == "active" && subscription.EndAt.After(time.Now().UTC()){
				return c.Next()
			}
		}
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message":"Unauthorized",
		})
		
	}
}