package controller

import (
	"context"
	"gobackend/connect"
	"gobackend/models"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func CreateCollection() fiber.Handler{
	return func(c *fiber.Ctx) error {
		user := c.Locals("user")
		claims,ok:=user.(jwt.MapClaims)
		if !ok{
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid JWT claims format",
			})
		}
		user_id, ok:=claims["aud"].(string)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "user id not okay",
			})
		}
		
		var col models.Collection
		if err := c.BodyParser(&col); err!=nil{
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":"Invalid request body",
			})
		}

		id,err:=primitive.ObjectIDFromHex(user_id)
		if err!=nil{
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":"Invalid request body",
			})
		}
		collection := models.Collection{
			Id:primitive.NewObjectID(),
			UserId: id,
			Title: col.Title,
			Description: col.Description,
		}

		_,err = connect.ColCollection.InsertOne(context.Background(),collection)
		if err!=nil{
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":"Try with different name or check for missing details",
			})
		}
		return c.JSON(fiber.Map{
			"id":collection.Id,
			"user_id":collection.UserId,
			"title":collection.Title,
			"description":collection.Description,
			"time":collection.CreatedAt,

		})

	}
}

func GetAllCollection() fiber.Handler{
	return func(c *fiber.Ctx) error{
		user := c.Locals("user")
		claims, ok:=user.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid JWT claims format",
			})
		}
		user_id, ok:=claims["aud"].(string)
	
		id, err:= primitive.ObjectIDFromHex(user_id)
		if err!=nil{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "id format is not valid",
			})
		}
		cursor, err := connect.ColCollection.Find(context.TODO(), bson.M{"user_id":id})
		if err!=nil{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "No collection could be found",
			})
		}
		var collections []models.Collection
		if err := cursor.All(context.TODO(), &collections); err!=nil{
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to parse project data",
			})
		}
		// fmt.Println(collections)
		return c.JSON(collections)
	}
}