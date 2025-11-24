package controller

import (
	"context"
	"fmt"
	"time"

	"gobackend/connect"
	"gobackend/env"
	"gobackend/models"

	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type ProjectUpdate struct {
    Title        *string `json:"title,omitempty" bson:"title,omitempty"`
    Description  *string `json:"description,omitempty" bson:"description,omitempty"`
    Tags         *string `json:"tags,omitempty" bson:"tags,omitempty"`
	FileUpload   string `bson:"fileUpload,omitempty" json:"fileUpload"`
    Thumbnail    *string `json:"thumbnail,omitempty" bson:"thumbnail,omitempty"`
    GithubLink   *string `json:"githublink,omitempty" bson:"githublink,omitempty"`
	DemoLink      string `json:"demolink,omitempty" bson:"demolink,omitempty"`
    LiveDemoLink *string `json:"livedemolink,omitempty" bson:"livedemolink,omitempty"`
	BlogLink     string `bson:"blogLink,omitempty" json:"blogLink"`
	TeamMembers  string `bson:"teamMembers,omitempty" json:"teamMembers"`
	UpdatedAt    time.Time `bson:"updated_time" json:"updated_time"`
    // (no CreatedAt here)
}

type ProjectResponse struct {
    Title        *string `json:"title,omitempty" bson:"title,omitempty"`
    Description  *string `json:"description,omitempty" bson:"description,omitempty"`
    Tags         *string `json:"tags,omitempty" bson:"tags,omitempty"`
	FileUpload   string `bson:"fileUpload,omitempty" json:"fileUpload,omitempty"`
    Thumbnail    *string `json:"thumbnail,omitempty" bson:"thumbnail,omitempty"`
    GithubLink   *string `json:"githublink,omitempty" bson:"githublink,omitempty"`
	DemoLink      string `json:"demolink,omitempty" bson:"demolink,omitempty"`
    LiveDemoLink *string `json:"livedemolink,omitempty" bson:"livedemolink,omitempty"`
	BlogLink     string `bson:"blogLink,omitempty" json:"blogLink,omitempty"`
	TeamMembers  string `bson:"teamMembers,omitempty" json:"teamMembers,omitempty"`
	Id string `bson:"id,omitempty" json:"id"`
	UserId string `bson:"user_id,omitempty" json:"user_id"`
	CollectionId string `bson:"collection_id,omitempty" json:"collection_id"`
	CreatedAt    time.Time `bson:"created_time" json:"created_at"`
	UpdatedAt    time.Time `bson:"updated_time" json:"updated_at"`
    
}

func UpdatedProject() *ProjectUpdate{
	return &ProjectUpdate{
		UpdatedAt: time.Now().UTC(),
	}
}

// import "github.com/gofiber/fiber/v2"

// func CreatePresignedUrlAndUploadObject(bucketName string , objectKey string,data[]byte, contentType string) 

func GetFilePresignedUrl() fiber.Handler{
	return func(c *fiber.Ctx) error{
		envs:=env.NewEnv()
		accessKey := envs.AWS_ACCESS_KEY_ID
		secretKey :=envs.AWS_SECRET_ACCESS_KEY

		cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		log.Fatal("failed to load config:", err)
	}
	
		bucketName:=envs.S3_BUCKET_NAME
		var req struct {
			FileKey []string `json:"fileKey"`
		}
		if err := c.BodyParser(&req); err!=nil{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":"Invalid request body",
			})
		}
		client := s3.NewFromConfig(cfg)
		presignClient := s3.NewPresignClient(client)
		urls := []string{}
		for _, filename:=range req.FileKey{
			objectKey:="uploads/" + string(filename)
			params := &s3.PutObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(objectKey),
			}
			presignedURL, err := presignClient.PresignPutObject(context.TODO(), params, func(opts *s3.PresignOptions) {
						opts.Expires = time.Hour // expires in 1 hour
			})
			if err != nil {
            return c.Status(500).JSON(fiber.Map{"error": "failed presigned url"})
        }
			urls = append(urls, presignedURL.URL)
		
		}

		return c.JSON(fiber.Map{"urls": urls})
	}
}

func CreateProject() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// first get the user email , for inserting to that userid
		col_id:= c.Params("col_id")
		user_id, err:= FetchUserId(c)
		if err!=nil{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":"failed to fetch user id",
			})
		}
		
		var p models.Project
		if err := c.BodyParser(&p); err!=nil{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":"Invalid request body",
			})
		}

		project := models.Project{
			Id:primitive.NewObjectID().Hex(),
			UserId:user_id,
			CollectionId:col_id,
			Title: p.Title,
			Description: p.Description,
			Tags:p.Tags, // write tech staack here
			FileUpload: p.FileUpload,
			Thumbnail: p.Tags,
			GithubLink: p.GithubLink,
			DemoLink:p.DemoLink,
			LiveDemoLink: p.LiveDemoLink,
			BlogLink:p.BlogLink,
			TeamMembers: p.TeamMembers,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		_,err = connect.ProjectCollection.InsertOne(context.TODO(), project)
		if err!=nil{
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "It can be duplicacy error or you might have missed some information.",
		})

		
	}
	return c.JSON(fiber.Map{"success": "created", "data" : project})
	}
}

// get the col_id , check the project with this col_id exists?
func GetAllProject() fiber.Handler{
	return func(c *fiber.Ctx) error{
		// envs:=env.NewEnv()
		user_id, err:=FetchUserId(c)
		if err!=nil{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "error fetching userId",
			})
		}

		fmt.Print("user id: ", user_id)
	
	
		cursor, err := connect.ProjectCollection.Find(context.TODO(), bson.M{"user_id":user_id})
		if err!=nil{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "No collection could be found",
			})
		}
		var projects []models.Project
		if err := cursor.All(context.TODO(), &projects); err!=nil{
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to parse project data",
			})
		}
		return c.JSON(projects)

	}
}
func GetAllProjectOfCollectionId() fiber.Handler{
	return func(c *fiber.Ctx) error {
		col_id := c.Params("col_id")
		// var project_info models.Project
		cursor,err := connect.ProjectCollection.Find(context.TODO(), bson.M{"collection_id":col_id})

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
		
		upd:= UpdatedProject()
		if err := c.BodyParser(&upd); err!=nil{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":"Invalid JSON",
			})
		}

		// build the $set doc
		setDoc, err:=  setDoc(upd)
		 if err != nil {
            return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to prepare update"})
        }
        if len(setDoc) == 0 {
            return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "No fields provided to update"})
        }

		objId, err := primitive.ObjectIDFromHex(p_id)
		if err !=nil{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":"Failed to convert in primitive type",
			})
		}
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
		fmt.Println("project id: ", p_id)
		if p_id==""{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":"Please provide project id",
			})
		}
		
		var project models.Project
		err := connect.ProjectCollection.FindOne(context.Background(), bson.M{"id": p_id} ).Decode(&project)
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