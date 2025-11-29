package controller

import (
	"context"
	"fmt"
	"gobackend/connect"
	"gobackend/env"
	"gobackend/models"
	"gobackend/services"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
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

type CollectionWithProjectsData struct{
	models.Collection `bson:",inline" json:"collection"`
	Projects []models.Project `json:"projects"`
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
		return c.JSON(collection)

	}
}

type GetAllCollectionResponse struct{
	Collections []models.Collection `json:"collections"`
	Code int `json:"code"`
	Page int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
	TotalPages int `json:"total_pages"`
} 
func GetAllCollection() fiber.Handler{
	return func(c *fiber.Ctx) error{

		user_id, err:= FetchUserId(c)
		if err!=nil{
			return c.Status(fiber.StatusBadRequest).JSON(GetAllCollectionResponse{
				Collections: nil,
				Code: fiber.StatusBadRequest,
			})
		}

		page,_:=strconv.Atoi(c.Query("page"))
		limit,_:=strconv.Atoi(c.Query("limit"))
		
		opts:=options.Find().
		SetSkip(int64((page-1)*limit)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{"created_at",-1}})

		cursor, err := connect.ColCollection.Find(context.TODO(), bson.M{"user_id":user_id},opts)
		if err!=nil{
			return c.Status(fiber.StatusBadRequest).JSON(GetAllCollectionResponse{
				Collections: nil,
				Code: fiber.StatusBadRequest,
				Page: page,
				Limit: limit,
				Total: 0,
				TotalPages: 0,
			})
		}
		
	
		total, err := connect.ColCollection.CountDocuments(context.TODO(), bson.M{"user_id":user_id})
		if err!=nil{
			return c.Status(fiber.StatusInternalServerError).JSON(GetAllCollectionResponse{
				Collections: nil,
				Code: fiber.StatusInternalServerError,
				Page: page,
				Limit: limit,
				Total: 0,
				TotalPages: 0,
			})
		}

		var tot_pages int
		if limit>0{
		    tot_pages = (int(total) + limit - 1) / limit
		}else{
			tot_pages = 0
		}
		var collections []models.Collection
		if err := cursor.All(context.TODO(), &collections); err!=nil{
			return c.Status(fiber.StatusInternalServerError).JSON(GetAllCollectionResponse{
				Collections: nil,
				Code: fiber.StatusInternalServerError,
				Page: page,
				Limit: limit,
				Total: int(total),
				TotalPages: tot_pages,
			})
		}
		// fmt.Println(collections)
		return c.JSON(GetAllCollectionResponse{
			Collections: collections,
			Code: fiber.StatusOK,
			Page: page,
			Limit: limit,
			Total: int(total),
			TotalPages: tot_pages,
		})
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
		envs:=env.NewEnv()
		col_id:=c.Params("id")
		var project models.Project
		_,err:=connect.ProjectCollection.Find(context.Background(), bson.M{"collection_id":col_id})
		if err!=nil{
			return c.Status(fiber.StatusInternalServerError).JSON("failed to extract project info")
		}
		var files []string
		files=project.FileUpload
		bucketName:=envs.S3_BUCKET_NAME
		if len(files)>0{
			for _,file:=range files{
				err:=services.DeleteFromS3(bucketName,file)
				if err!=nil{
					return c.Status(fiber.StatusInternalServerError).JSON("failed to delete file")
				}
			}
		}

		projectFilter:=bson.M{"collection_id":col_id}
		_,err=connect.ProjectCollection.DeleteMany(context.Background(),projectFilter)
		if err!=nil{
			return c.Status(fiber.StatusInternalServerError).JSON("failed to delete project")
		}

		collectionFilter:= bson.M{"id":col_id}
		_,err=connect.ColCollection.DeleteOne(context.Background(),collectionFilter)
		if err!=nil{
			return c.Status(fiber.StatusInternalServerError).JSON("failed to delete collection")
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

func DeleteAllCollection() fiber.Handler{
	return func(c* fiber.Ctx)error{
		envs:=env.NewEnv()
		user_id,err:= FetchUserId(c)
		if err!=nil{
			return c.Status(fiber.StatusBadRequest).JSON(DeleteCollectionResponse{
				Message: "failed to fetch user_id",
				Code: fiber.StatusBadRequest,
			})
		}
		// fetch all the projects of the user
		var project []models.Project
		cursor,err:=connect.ProjectCollection.Find(context.Background(), bson.M{"user_id":user_id})
		if err!=nil{
			return c.Status(fiber.StatusBadRequest).JSON(DeleteCollectionResponse{
				Message: "failed to fetch projects",
				Code: fiber.StatusBadRequest,
			})
		}
		cursor.All(context.Background(), &project)
		// files to delete from s3
		var files []string
		for _,p:=range project{
			files = append(files,p.FileUpload...)
		}
		bucketName:=envs.S3_BUCKET_NAME
		if len(files)>0{
			for _,file:=range files{
				err:=services.DeleteFromS3(bucketName,file)
				if err!=nil{
					return c.Status(fiber.StatusInternalServerError).JSON(DeleteCollectionResponse{
						Message: "failed to delete file",
						Code: fiber.StatusInternalServerError,
					})
				}
			}
		}

		// now delete all the project of the user 
		projectFilter:= bson.M{"user_id":user_id}
		_,err=connect.ProjectCollection.DeleteMany(context.Background(),projectFilter)
		if err!=nil{
			return c.Status(fiber.StatusInternalServerError).JSON(DeleteCollectionResponse{
				Message: "failed to delelte projects",
				Code: fiber.StatusInternalServerError,
			})
		}

		// delete all the collections of the user
		filter := bson.M{"user_id":user_id}
		_,err=connect.ColCollection.DeleteMany(context.TODO(), filter)
		if err!=nil{
			return c.Status(fiber.StatusBadRequest).JSON(DeleteCollectionResponse{
				Message: "failed to delete collections",
				Code: fiber.StatusBadRequest,
			})
		}
		return c.JSON(DeleteCollectionResponse{
			Message: "collections deleted successfully",
			Code: fiber.StatusOK,
		})
	}
}

func FetchAllCollectionWithProjects() fiber.Handler{
	return func(c* fiber.Ctx) error{
		user_id, err:= FetchUserId(c)
		if err!=nil{
			return c.Status(fiber.StatusBadRequest).JSON("failed to fetch user_id")
		}
		// join collection and project
		var colData []models.Collection


		cursor, err := connect.ColCollection.Find(
			context.Background(),
			bson.M{"user_id": user_id},
		)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Failed to find collections"})
		}

		if err := cursor.All(context.Background(), &colData); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Decode error"})
		}

		var result []CollectionWithProjectsData
		for _, col := range colData {
			pipeline := mongo.Pipeline{
				bson.D{
					{Key: "$match", Value: bson.D{
						{Key: "id", Value: col.Id},
					}},
				},
				bson.D{
					{Key: "$lookup", Value: bson.D{
						{Key: "from", Value: "projects"},
						{Key: "localField", Value: "id"},
						{Key: "foreignField", Value: "collection_id"},
						{Key: "as", Value: "projects"},
					}},
				},
			}
			cur, err := connect.ColCollection.Aggregate(context.Background(), pipeline)
			if err != nil {
				continue
			}

			var tmp []CollectionWithProjectsData
			if err := cur.All(context.Background(), &tmp); err == nil && len(tmp) > 0 {
				result = append(result, tmp[0])
			}
		}

	return c.JSON(result)


	}
}
