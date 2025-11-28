package controller

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"gobackend/connect"
	"gobackend/env"
	"gobackend/models"
	"gobackend/services"

	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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

func GenerateObjectKey(filename string) string {
    id := uuid.New().String()
    return fmt.Sprintf("uploads/%s-%s", id, filename)
}

func GetFilePresignedUrl() fiber.Handler{
	return func(c *fiber.Ctx) error{
		fmt.Println("get file presigned url")
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
		fmt.Println("req: ", req)
		client := s3.NewFromConfig(cfg)
		presignClient := s3.NewPresignClient(client)
		urls := []string{}
		for _, filename:=range req.FileKey{
			objectKey:=GenerateObjectKey(filename)
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
type GetAllProjectResponse struct{
	Projects []models.Project `json:"projects"`
	Code int `json:"code"`
	Page int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
	TotalPages int `json:"total_pages"`
}
func GetAllProject() fiber.Handler{
	return func(c *fiber.Ctx) error{
		// envs:=env.NewEnv()
		user_id, err:=FetchUserId(c)
		if err!=nil{
			return c.Status(fiber.StatusBadRequest).JSON(GetAllProjectResponse{
				Projects: nil,
				Code: fiber.StatusBadRequest,
				Page: 0,
				Limit: 0,
				Total: 0,
				TotalPages: 0,
			})
		}

		fmt.Print("user id: ", user_id)
		page,_:=strconv.Atoi(c.Query("page"))
		limit,_:=strconv.Atoi(c.Query("limit"))

		opts:=options.Find().
			SetLimit(int64(limit)).
			SetSkip(int64((page-1)*limit)).
			SetSort(bson.D{{"created_at", -1}})
	
		cursor, err := connect.ProjectCollection.Find(context.TODO(), bson.M{"user_id":user_id},opts)
		if err!=nil{
			return c.Status(fiber.StatusBadRequest).JSON(GetAllProjectResponse{
				Projects: nil,
				Code: fiber.StatusBadRequest,
				Page: page,
				Limit: limit,
				Total: 0,
				TotalPages: 0,
			})
		}
		total,err := connect.ProjectCollection.CountDocuments(context.TODO(), bson.M{"user_id":user_id})
		if err!=nil{
			return c.Status(fiber.StatusInternalServerError).JSON(GetAllProjectResponse{
				Projects: nil,
				Code: fiber.StatusInternalServerError,
				Page: page,
				Limit: limit,
				Total: 0,
				TotalPages: 0,
			})
		}
		var projects []models.Project
		var tot_pages int
		if limit>0{
		    tot_pages = (int(total) + limit - 1) / limit
		}else{
			tot_pages = 0
		}
		if err := cursor.All(context.TODO(), &projects); err!=nil{
			return c.Status(fiber.StatusInternalServerError).JSON(GetAllProjectResponse{
				Projects: nil,
				Code: fiber.StatusInternalServerError,
				Page: page,
				Limit: limit,
				Total: int(total),
				TotalPages: tot_pages,
			})
		}
		return c.JSON(GetAllProjectResponse{
			Projects: projects,
			Code: fiber.StatusOK,
			Page: page,
			Limit: limit,
			Total: int(total),
			TotalPages:tot_pages,
		})

	}
}

type GetAllProjectOfCollectionIdResponse struct{
	Projects []models.Project `json:"projects"`
	Code int `json:"code"`
	Page int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
	TotalPages int `json:"total_pages"`
}
func GetAllProjectOfCollectionId() fiber.Handler{
	return func(c *fiber.Ctx) error {
		col_id := c.Params("col_id")

		page,_:=strconv.Atoi(c.Query("page"))
		limit,_:=strconv.Atoi(c.Query("limit"))

		opts:=options.Find().SetSkip(int64((page-1)*limit)).SetLimit(int64(limit)).SetSort(bson.D{{"created_at", -1}})
		// var project_info models.Project
		cursor,err := connect.ProjectCollection.Find(context.TODO(), bson.M{"collection_id":col_id},opts)
		// cursor,err := connect.ProjectCollection.Find(context.TODO(), bson.M{"email":email})
		if err!=nil{
			return c.Status(fiber.StatusBadRequest).JSON(GetAllProjectOfCollectionIdResponse{
				Projects: nil,
				Code: fiber.StatusBadRequest,
				Page: page,
				Limit: limit,
				Total: 0,
				TotalPages: 0,
			})
		}
		total,err := connect.ProjectCollection.CountDocuments(context.TODO(), bson.M{"collection_id":col_id})
		if err!=nil{
			return c.Status(fiber.StatusInternalServerError).JSON(GetAllProjectOfCollectionIdResponse{
				Projects: nil,
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
		var projects []models.Project
		if err := cursor.All(context.TODO(), &projects); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(GetAllProjectOfCollectionIdResponse{
				Projects: nil,
				Code: fiber.StatusInternalServerError,
				Page: page,
				Limit: limit,
				Total: int(total),
				TotalPages: tot_pages,
			})
		}
		// fmt.Println(projects)
		return c.JSON(GetAllProjectOfCollectionIdResponse{
			Projects: projects,
			Code: fiber.StatusOK,
			Page: page,
			Limit: limit,
			Total: int(total),
			TotalPages: tot_pages,
		})
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
type ReadProjectResponse struct{
	Project models.Project `json:"project"`
	Code int `json:"code"`
}
func GetProjectByProjectId() fiber.Handler{
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
			return c.Status(fiber.StatusBadRequest).JSON(ReadProjectResponse{
				Project: project,
				Code: fiber.StatusBadRequest,
			})
		}
		return c.JSON(ReadProjectResponse{
			Project: project,
			Code: fiber.StatusOK,
		})
	}
}
// TODO: delete the files of the project also 
func DeleteProject() fiber.Handler{
	return func(c *fiber.Ctx) error{
		// get the project id
		envs:=env.NewEnv()
		p_id:= c.Params("projectid")
		if p_id==""{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":"Please provide project id",
			})
		}

		var project models.Project
		err := connect.ProjectCollection.FindOne(context.Background(), bson.M{"id": p_id} ).Decode(&project)
		if err!=nil{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error":"Project was not deleted successful"})
		}
		files:=project.FileUpload
		bucketName:=envs.S3_BUCKET_NAME
		if len(files)>0{
			for _,file:=range files{
				err:=services.DeleteFromS3(bucketName,file)
				if err!=nil{
					return c.Status(fiber.StatusInternalServerError).JSON("failed to delete file")
				}
			}
		}
		filter := bson.M{"id":p_id}
		result, err := connect.ProjectCollection.DeleteOne(context.TODO(),filter)
		if err!=nil{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"eror":"Project was not deleted successful"})
		}

		return c.JSON(result)
	}
}

func DeleteAllProject() fiber.Handler{
	return func(c *fiber.Ctx) error{
		envs:=env.NewEnv()
		col_id := c.Params("col_id")
		if col_id==""{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":"collection id needed",
			})
		}
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
type DeleteFileResponse struct{
	Message string `json:"message"`
	Code	int `json:"code"`
}
func DeleteFile() fiber.Handler{
	return func(c* fiber.Ctx)error{
		envs:=env.NewEnv()
		project_id:=c.Params("project_id")
		if(project_id==""){
			return c.Status(fiber.StatusBadRequest).JSON(DeleteFileResponse{
				Message:"project id needed",
				Code: fiber.StatusBadRequest,
			})
		}
		key:=c.Query("key")
		if key==""{
			return c.Status(fiber.StatusBadRequest).JSON(DeleteFileResponse{
				Message:"key needed",
				Code: fiber.StatusBadRequest,
			})
		}
		fmt.Println("key: ", key)
		
		
		bucket:=envs.S3_BUCKET_NAME
		err := services.DeleteFromS3(bucket,key)
		if err!=nil{
			return c.Status(fiber.StatusBadRequest).JSON(DeleteFileResponse{
				Message:"failed to delete file",
				Code: fiber.StatusBadRequest,
			})
		}

		filter:=bson.M{"id":project_id}
		update := bson.M{
			"$pull": bson.M{
				"fileUpload": key,
			},
		}
		_,err = connect.ProjectCollection.UpdateOne(context.Background(),filter,update)
		if err!=nil{
			return c.Status(fiber.StatusBadRequest).JSON(DeleteFileResponse{
				Message:"failed to delete file",
				Code: fiber.StatusBadRequest,
			})
		}

		return c.JSON(DeleteFileResponse{
			Message:"file deleted successfully",
			Code: fiber.StatusOK,
		})

	}
}