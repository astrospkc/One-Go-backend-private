package controller

import (
	"bytes"
	"context"
	"fmt"
	"gobackend/connect"
	"gobackend/models"
	"gobackend/services"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// type Media struct{
// 	Id 			primitive.ObjectID	`bson:"id,omitempty" json:"id"`
// 	UserId		string 	`bson:"user_id" json:"user_id"`
// 	CollectionId primitive.ObjectID		`bson:"collection_id" json:"collection_id"`
// 	File        string 	`bson:"file" json:"file"`
// 	CreatedAt	time.Time	`bson:"time" json:"time"`
// }

func PostMedia() fiber.Handler{
	return func(c *fiber.Ctx) error {
		// get the url , get the body
		col_id := c.Params("col_id")
		id,err := primitive.ObjectIDFromHex(col_id)
		if err!=nil{
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":"id format not valid",
			})
		}
		user:= c.Locals("user")
		claims,ok:=user.(jwt.MapClaims)
		if !ok{
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":"Invalid Jwt claims format",
			})
		}

		user_id, ok:= claims["aud"].(string)
		if !ok{
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":"Invalid or missing aud field",
			})
		}
		me := new(models.Media)
		fileHeader, err := c.FormFile("file")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("File is required")
		}

		// open the file
		file , err:=fileHeader.Open()
		if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to read file")
		}
		defer file.Close()

		// Read content into buffer
		var buf bytes.Buffer
		_, err = io.Copy(&buf, file)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to buffer file")
		}

		filename := fileHeader.Filename
		ext := strings.ToLower(filepath.Ext(filename))
		ext = strings.TrimPrefix(ext, ".") 
		// ext := strings.ToLower(filepath.Ext(filename))
		fmt.Println("extension: ", filename)
		mimeType:=fileHeader.Header.Get("Content-Type")
		fmt.Println("mimetype: ", mimeType)
		
		if err := c.BodyParser(me); err!=nil{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":"Invalid request body",
			})
		}
		
		bucketName:=os.Getenv("S3_BUCKET_NAME")
		objectKey:=fmt.Sprintf("uploads/pic_%s.%s", time.Now().Format("20060102_150405"), ext)
		_, err = services.CreatePresignedUrlAndUploadObject(bucketName, objectKey,buf.Bytes(),mimeType)
		if err != nil {
			log.Fatalf("Failed to generate URL: %v", err)
		}

		
		media := models.Media{
			Id:primitive.NewObjectID(),
			UserId: user_id,
			CollectionId: id,
			Key:objectKey,
			Title:me.Title,
			Content: me.Content,
			CreatedAt: time.Now(),
		}
		_,err = connect.MediaCollection.InsertOne(context.Background(), media)
		if err!=nil{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":"looks like there is an error while inserting data",
			})
		}

		// fmt.Println("url - : ", url)
		return c.JSON(fiber.Map{
			"success":"created",
			"media":media,
		})

	}

	
}

func GetAllMediaFiles() fiber.Handler{
		return func(c *fiber.Ctx) error {
		col_id := c.Params("col_id")
		id,err := primitive.ObjectIDFromHex(col_id)
		if err!=nil{
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":"id format not valid",
			})
		}
	
		cursor, err := connect.MediaCollection.Find(context.TODO(), bson.M{"collection_id":id})
		if err!=nil{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":"No media could be found",
			})
		}

		var medias []models.Media
		if err := cursor.All(context.TODO(), &medias); err!=nil{
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":"Failed to parse media data",
			})
		}
		return c.JSON(medias)

		}

	}