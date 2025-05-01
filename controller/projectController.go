package controller

import (
	"context"
	"fmt"

	"gobackend/connect"
	"gobackend/models"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type ProjectUpdate struct {
    Title        *string `json:"title,omitempty" bson:"title,omitempty"`
    Description  *string `json:"description,omitempty" bson:"description,omitempty"`
    Tags         *string `json:"tags,omitempty" bson:"tags,omitempty"`
    Thumbnail    *string `json:"thumbnail,omitempty" bson:"thumbnail,omitempty"`
    GithubLink   *string `json:"githublink,omitempty" bson:"githublink,omitempty"`
    LiveDemoLink *string `json:"livedemolink,omitempty" bson:"livedemolink,omitempty"`
    // (no CreatedAt here)
}

// import "github.com/gofiber/fiber/v2"

func CreateProject() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// first get the user email , for inserting to that userid
		id:= c.Params("col_id")
		user := c.Locals("user")
		claims,ok := user.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid JWT claims format",
			})
		}
		user_id, ok := claims["aud"].(string)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "user id not okay",
			})
		}

		var p models.Project
		if err := c.BodyParser(&p); err!=nil{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":"Invalid request body",
			})
		}
		
	
		u_id,err := primitive.ObjectIDFromHex(user_id)
		if err!=nil{
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "id format invalid",
			})
		}
		col_id ,err:= primitive.ObjectIDFromHex(id)
		if err!=nil{
			 return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid ID format",
        })
		}
		fmt.Printf("%T\n", col_id)
		project := models.Project{
			Id:primitive.NewObjectID(),
			UserId:u_id,
			CollectionId:col_id,
			Title: p.Title,
			Description: p.Description,
			Tags:p.Tags,
			Thumbnail: p.Tags,
			GithubLink: p.GithubLink,
			LiveDemoLink: p.LiveDemoLink,
		}

		_,err = connect.ProjectCollection.InsertOne(context.TODO(), project)
		if err!=nil{
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "It can be duplicacy error or you might have missed some information.",
		})

		
	}
	return c.JSON(fiber.Map{"success": "created",})
	}
}

func ReadProject() fiber.Handler{
	return func(c *fiber.Ctx) error {
		col_id := c.Params("col_id")
		
		id ,err:= primitive.ObjectIDFromHex(col_id)
		if err!=nil{
			 return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid ID format",
        })
		}
		
		
		// var project_info models.Project
		cursor,err := connect.ProjectCollection.Find(context.TODO(), bson.M{"collection_id":id})

		// cursor,err := connect.ProjectCollection.Find(context.TODO(), bson.M{"email":email})
		if err!=nil{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "No project could be found",
			})
		}
		
		var projects []models.Project
		if err := cursor.All(context.TODO(), &projects); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to parse project data",
			})
		}
		// fmt.Println(projects)
		return c.JSON(projects)
	}
}

func setDoc(upd *ProjectUpdate) (bson.M, error){
	data, err:= bson.Marshal(upd)
	if err!=nil{
		return nil,err
	}

	var m bson.M
	if err:= bson.Unmarshal(data, &m); err!=nil{
		return nil, err
	}
	return m, nil
}

func UpdateProject() fiber.Handler {
	return func(c *fiber.Ctx) error{
		p_id := c.Params("projectid")
		if p_id == ""{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":"Please provide project id",
			})
		}
		
		var upd ProjectUpdate
		if err := c.BodyParser(&upd); err!=nil{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":"Invalid JSON",
			})
		}

		// build the $set doc
		setDoc, err:=  setDoc(&upd)
		 if err != nil {
            return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to prepare update"})
        }
        if len(setDoc) == 0 {
            return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "No fields provided to update"})
        }

		fmt.Println("setDoc: ", setDoc)
		objId, err := primitive.ObjectIDFromHex(p_id)
		if err !=nil{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":"Failed to convert in primitive type",
			})
		}


		// fmt.Println("object Id: ", objId)
		// // lets find first to check it is actually working or not
		// var foundProj models.Project
		// err = connect.ProjectCollection.FindOne(context.TODO(),bson.M{"id":objId}).Decode(&foundProj)
		// if err!=nil{
		// 	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		// 		"error":"Could not find the project",
		// 	})
		// }
		// fmt.Println("foundProj: ", foundProj)


		fmt.Printf("pid: %T \n", objId)
		filter:=bson.M{"id":objId}
		update:=bson.M{"$set":setDoc}
		opts:=options.FindOneAndUpdate().SetReturnDocument(options.After)

		var updatedDoc models.Project
		err = connect.ProjectCollection.FindOneAndUpdate(context.TODO(),filter, update, opts).Decode(&updatedDoc)
		if err != nil {
            if err == mongo.ErrNoDocuments {
                return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Project not found"})
            }
            return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Update failed"})
        }

        return c.JSON(updatedDoc)
	}
}

func FindOneViaPID() fiber.Handler{
	return func(c *fiber.Ctx) error{
		p_id:= c.Params("projectid")
		if p_id==""{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":"Please provide project id",
			})
		}
		objID, err := primitive.ObjectIDFromHex(p_id)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid project ID format",})
		}
	
		var project ProjectUpdate
		err = connect.ProjectCollection.FindOne(context.TODO(), bson.M{"id": objID} ).Decode(&project)
		if err!=nil{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":"Failed to find the project with this project id, try with valid project id",
			})
		}
		return c.JSON(project)


	}
}

func DeleteProject() fiber.Handler{
	return func(c *fiber.Ctx) error{
		// get the project id
		p_id:= c.Params("projectid")
		if p_id==""{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":"Please provide project id",
			})
		}
		objId, err := primitive.ObjectIDFromHex(p_id)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid project ID format",})
		}

		filter := bson.M{"id":objId}
		
		result, err := connect.ProjectCollection.DeleteOne(context.TODO(),filter)
		if err!=nil{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"eror":"Project was not deleted successful"})
		}

		return c.JSON(result)
	}
}

func DeleteAllProject() fiber.Handler{
	return func(c *fiber.Ctx) error{
		u_id := c.Params("u_id")
		if u_id==""{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":"User id needed",
			})
		}
		uid, err := primitive.ObjectIDFromHex(u_id)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid project ID format",})
		}
		filter := bson.M{"user_id":uid}
		result,err:=connect.ProjectCollection.DeleteMany(context.TODO(), filter)
		if err!=nil{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"eror":"Project was not deleted successful"})
		}
		return c.JSON(result)




	}
}