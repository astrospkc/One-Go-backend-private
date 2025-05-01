package controller

import (
	"context"
	"gobackend/connect"
	"gobackend/models"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)


type LinkResponse struct{
	Id 			primitive.ObjectID	`bson:"id,omitempty" json:"id"`
	UserId		string 	`bson:"user_id,omitempty" json:"user_id"`
	Source		string	`bson:"source,omitempty" json:"source"`
	Title		string	`bson:"title,omitempty" json:"title"`
	Url			string	`bson:"url,omitempty" json:"url"` 
	Description	string	`bson:"description,omitempty" json:"description"`
	Category	string	`bson:"category,omitempty" json:"category"`
}
func CreateLink() fiber.Handler{
	return func(c *fiber.Ctx) error{
		user := c.Locals("user")
		claims,ok:=user.(jwt.MapClaims)
		if !ok{
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":"Invalid JWT claims format",
			})
		}

		user_id,ok:=claims["aud"].(string)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":"Invalid or missing and field",
			})
		}

		var p models.Link
		if err := c.BodyParser(&p); err!=nil{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":"Invalid request body",
			})
		}

		
		link := models.Link{
			Id:primitive.NewObjectID(),
			UserId:user_id,
			Source:p.Source,
			Title:p.Title,
			Url:p.Url,
			Description: p.Description,
			Category: p.Category,
		}
		_,err := connect.LinksCollection.InsertOne(context.Background(),link)
		if err!=nil{
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":"looks like some informaton are missing , try again",
			})
		}
		return c.JSON(fiber.Map{"success":"created"})

	}
}

func ReadLink() fiber.Handler{
	return func(c *fiber.Ctx) error {
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
				"error": "Invalid or missing  aud field",
			})
		}
		
		cursor, err := connect.LinksCollection.Find(context.TODO(),bson.M{"user_id":user_id})
		if err!=nil{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":"No Blogs could be found",
			})
		}

		var blogs []models.Link
		if err:=cursor.All(context.TODO(), &blogs); err!=nil{
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":"Failed to parse blogs data",
			})
		}

		return c.JSON(blogs)
	}
}

func ReadLinkWithLinkId() fiber.Handler{
	return func(c *fiber.Ctx) error {
		link_id:= c.Params("linkid")
		if link_id==""{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":"Please provide project id",
			})
		}
		objID, err := primitive.ObjectIDFromHex(link_id)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid link ID format",})
		}

		var link LinkResponse
		err = connect.LinksCollection.FindOne(context.TODO(), bson.M{"id":objID}).Decode(&link)
		if err!=nil{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":"Failed to find the link with given id",
			})
		}
		return c.JSON(link)
	}
}

func setLink(upd *LinkResponse) (bson.M, error){
	data, err := bson.Marshal(upd)
	if err!=nil{
		return nil, err
	}
	var m bson.M
	if err:= bson.Unmarshal(data, &m); err!=nil{
		return nil, err
	}
	return m, nil
}
func UpdateLinkWithLinkId() fiber.Handler{
	return func(c *fiber.Ctx) error {
		l_id := c.Params("linkid")
		if l_id==""{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":"Please provide link id",
			})
		}

		var upd LinkResponse
		if err := c.BodyParser(&upd); err!=nil{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":"invalid JSON",
			})
		}

		setLink, err := setLink(&upd)
		if err != nil {
            return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to prepare update"})
        }
        if len(setLink) == 0 {
            return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "No fields provided to update"})
        }

		objId, err := primitive.ObjectIDFromHex(l_id)
		if err!=nil{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":"Failed to convert in primitive type",
			})
		}

		filter := bson.M{"id":objId} 
		update:= bson.M{"$set":setLink}
		opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
		
		var updatedLink models.Link
		err = connect.LinksCollection.FindOneAndUpdate(context.TODO(), filter, update, opts).Decode(&updatedLink)
		if err !=nil{
			if err ==mongo.ErrNoDocuments{
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error":"Link not found"})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error":"Update failed"})
		}
		return c.JSON(updatedLink)
	}
}

func DeleteLinkWithLinkId() fiber.Handler{
	return func(c *fiber.Ctx) error {

		l_id:= c.Params("linkid")
		if l_id==""{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":"Please provide link id",
			})
		}

		objId, err := primitive.ObjectIDFromHex(l_id)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid Link ID format",})
		}

		filter := bson.M{"id":objId}
		result, err := connect.LinksCollection.DeleteOne(context.TODO(),filter)
		if err!=nil{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"eror":"Link was not deleted successful"})
		}
		return c.JSON(result)

	}
}

func DeleteAllLinks() fiber.Handler{
	return func(c *fiber.Ctx) error{
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
				"error": "Invalid or missing  aud field",
			})
		}
		result,err := connect.LinksCollection.DeleteMany(context.Background(),bson.M{"user_id":user_id})
		if err!=nil{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":"failed to delete all ",
			})
		}
		return c.JSON(result)
	}
}