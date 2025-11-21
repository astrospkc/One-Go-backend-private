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
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Collection struct{
	Id       primitive.ObjectID `bson:"id,omitempty" json:"id"`
	UserId		 primitive.ObjectID 	`bson:"user_id" json:"user_id"`
	Title 		string `bson:"title" json:"title"`
	Description	string `bson:"description" json:"description"`
	TotalProject int64 `bson:"total_project" json:"total_project"`
	CreatedAt	time.Time	`bson:"created_at" json:"created_at"`
	UpdatedAt	time.Time	`bson:"updated_at" json:"updated_at"`
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
			CreatedAt: time.Now().UTC(),
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
type GetCollectionResponse struct{
	Collection models.Collection `json:"collection"`
	Code int `json:"code"`
}

func GetCollection() fiber.Handler{
	return func(c* fiber.Ctx) error{
		col_id:=c.Params("id")
		var col models.Collection
		err := connect.ColCollection.FindOne(context.Background(), bson.M{"id":col_id}).Decode(&col)
		if err!=nil{
			return c.Status(fiber.StatusInternalServerError).JSON(GetCollectionResponse{
				Collection: col,
				Code: fiber.StatusInternalServerError,
			})
		}
		return c.JSON(GetCollectionResponse{
			Collection: col,
			Code: fiber.StatusOK,
		})
	}
}

type DeleteCollectionResponse struct{
	Message string `json:"message"`
	Code int `json:"code"`
}
// get the id of the collection
func DeleteCollection() fiber.Handler{
	return func(c* fiber.Ctx)error{
		col_id:=c.Params("id")
		_,err := connect.ColCollection.DeleteOne(context.Background(), bson.M{"id":col_id})
		if err!=nil{
			return c.Status(fiber.StatusInternalServerError).JSON(DeleteCollectionResponse{
				Message: "Failed to delete collection",
				Code: fiber.StatusInternalServerError,
			})
		}
		return c.JSON(DeleteCollectionResponse{
			Message: "Collection deleted successfully",
			Code: fiber.StatusOK,
		})
	}
}


type UpdateCollectionResponse struct {
    Collection models.Collection `json:"collection"`
    Code       int               `json:"code"`
}

func UpdateCollection() fiber.Handler {
    return func(c *fiber.Ctx) error {
        colID := c.Params("id")

        // Parse incoming body
        var payload models.Collection
        if err := c.BodyParser(&payload); err != nil {
            return c.Status(fiber.StatusBadRequest).JSON(UpdateCollectionResponse{
                Code: fiber.StatusBadRequest,
            })
        }

        // Build dynamic update document
        updateFields := bson.M{}

        if payload.Title != "" {
            updateFields["title"] = payload.Title
        }
        if payload.Description != "" {
            updateFields["description"] = payload.Description
        }
        

        // Always update time
        updateFields["updated_at"] = time.Now().UTC()
	

        // If nothing to update
        if len(updateFields) == 1 { // only updatedAt present
            return c.Status(fiber.StatusBadRequest).JSON(UpdateCollectionResponse{
                Code: fiber.StatusBadRequest,
            })
        }

        filter := bson.M{"id": colID}

        update := bson.M{
            "$set": updateFields,
        }

        opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
        var updated models.Collection

        err := connect.ColCollection.FindOneAndUpdate(
            context.Background(),
            filter,
            update,
            opts,
        ).Decode(&updated)
        if err != nil {
            return c.Status(fiber.StatusInternalServerError).JSON(UpdateCollectionResponse{
                Code: fiber.StatusInternalServerError,
            })
        }

        return c.JSON(UpdateCollectionResponse{
            Collection: updated,
            Code:       fiber.StatusOK,
        })
    }
}
