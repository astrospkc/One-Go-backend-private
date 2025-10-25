package controller

import (
	"context"
	"fmt"
	"gobackend/connect"
	"gobackend/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Collection struct{
	Id       primitive.ObjectID `bson:"id,omitempty" json:"id"`
	UserId		 primitive.ObjectID 	`bson:"user_id" json:"user_id"`
	Title 		string `bson:"title" json:"title"`
	Description	string `bson:"description" json:"description"`
	TotalProject int64 `bson:"total_project" json:"total_project"`
	CreatedAt	time.Time	`bson:"time" json:"time"`
}

func CreateCollection() fiber.Handler{
	return func(c *fiber.Ctx) error {
		fmt.Println("createing collection : ")
		user_id, err:= FetchUserId(c)
		fmt.Println("crate colle: ", user_id)
		if err!=nil{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":"failed to fetch user_id",
		})
		}
		
		var col models.Collection
		if err := c.BodyParser(&col); err!=nil{
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":"Invalid request body",
			})
		}

		
		collection := models.Collection{
			Id:primitive.NewObjectID().Hex(),
			UserId: user_id,
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

		user_id, err:= FetchUserId(c)
		if err!=nil{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":"failed to fetch user_id",
		})
		}
		cursor, err := connect.ColCollection.Find(context.TODO(), bson.M{"user_id":user_id})
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