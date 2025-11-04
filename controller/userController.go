package controller

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"gobackend/connect"
	"gobackend/env"
	"gobackend/models"
	"gobackend/services"
	"io"
	"log"
	"path/filepath"
	"strings"

	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

// TODO: later on add Project , category , links, blog, media, resume, subscription, usersubscription, apikey , all of these in UserResponse
type UserResponse struct {
	Id 			string	`bson:"id,omitempty" json:"id"`
	Name 		string `bson:"name,omitempty" json:"name"`
	Email 		string `bson:"email,omitempty" json:"email"`
	ProfilePic  string `bson:"profile_pic,omitempty" json:"profile_pic"`
	Role 		string	`bson:"role,omitempty" json:"role"`
	APIkey		string  `bson:"api_key,omitempty" json:"api_key"`
	
}

type Response struct{
	
	Token   string   `json:"token"`
	User		models.User `json:"user"`
}


type Project struct{
	Id       primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	Title		 *string	`bson:"title" json:"title"`
	Description	 *string	`bson:"description,omitempty" json:"description"`
	Tags		 *string	`bson:"tags,omitempty" json:"tags"`
	Thumbnail 	 *string	`bson:"thumbnail,omitempty" json:"thumbnail"`
	GithubLink	 *string	`bson:"githublink,omitempty" json:"githublink"`
	LiveDemoLink *string	`bson:"livedemolink,omitempty" json:"liveddemolink"`
	
}

type APIkey struct{
	Id 			primitive.ObjectID	`bson:"id,omitempty" json:"id"`
	Userid		string	`bson:"user_id" json:"user_id"`
	Key 		string	`bson:"key" json:"key"`
	UsageLimit	string	`bson:"usagelimit" json:"usagelimit"`
	CreatedAt	time.Time	`bson:"createdat" json:"createdat"`
	Revoked		bool	`bson:"revoked" json:"revoked"`
	
}




// first createtoken
func CreateToken(userid string) (string, error){
	envs:= env.NewEnv()
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userid,                    // Subject (user identifier)
		"iss": "One-Go",                  // Issuer
		"aud": userid,           // Audience (user role)
		"exp": time.Now().Add(time.Hour).Unix(), // Expiration time
		"iat": time.Now().Unix(),                 // Issued at
	})
	secret :=[]byte(envs.JWT_SECRET)
	tokenString, err := claims.SignedString(secret)
	if err!=nil{
		return "", err
	}
	fmt.Printf("Token claims added: %+v\n", claims)
	return tokenString, nil
}




func CreateUser() fiber.Handler {
	return func(c *fiber.Ctx) error {
		envs := env.NewEnv()

		// Generate API Key
		apiKey, err := GenerateApiKey()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to generate API key",
			})
		}

		// Read form values
		name := c.FormValue("name")
		email := c.FormValue("email")
		password := c.FormValue("password")
		role := c.FormValue("role")

		if email == "" || password == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Email and password are required",
			})
		}

		// Check if user already exists
		var existingUser models.User
		err = connect.UsersCollection.FindOne(context.TODO(), bson.M{"email": email}).Decode(&existingUser)
		if err == nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Email is already in use",
			})
		}

		// Handle optional profile picture
		var objectKey string
		picHeader, err := c.FormFile("file")
		if err == nil && picHeader != nil {
			file, err := picHeader.Open()
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Failed to open uploaded file",
				})
			}
			defer file.Close()

			var buf bytes.Buffer
			_, err = io.Copy(&buf, file)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to buffer uploaded file",
				})
			}

			filename := picHeader.Filename
			ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(filename)), ".")
			mimeType := picHeader.Header.Get("Content-Type")
			objectKey = fmt.Sprintf("uploads/pic_%s.%s", time.Now().Format("20060102_150405"), ext)

			_, err = services.CreatePresignedUrlAndUploadObject(envs.S3_BUCKET_NAME, objectKey, buf.Bytes(), mimeType)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to upload profile picture",
				})
			}
		}

		// Hash password
		hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to hash password",
			})
		}

		// Create user document
		user := models.User{
			Id:         primitive.NewObjectID().Hex(),
			Name:       name,
			Email:      email,
			Password:   string(hashedPass),
			ProfilePic: objectKey,
			Role:       role,
			APIkey:     apiKey,
		}

		// Save user
		_, err= connect.UsersCollection.InsertOne(context.TODO(), user)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create user",
			})
		}
		
		
		tokenString, err := CreateToken(user.Id)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create token",
			})
		}

		// Set secure cookie
		c.Cookie(&fiber.Cookie{
			Name:     "token",
			Value:    tokenString,
			HTTPOnly: true,
			Secure:   true,
			Path:     "/",
			MaxAge:   1000 * 60 * 60 * 24 * 5,
		})

		// Save API key doc
		api := models.APIkey{
			Id:         primitive.NewObjectID().Hex(),
			UserId:    	user.Id,
			Key:        apiKey,
			UsageLimit: 50,
		}
		_, err = connect.APIKeyCollection.InsertOne(context.TODO(), api)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to save API key",
			})
		}
		userRes := models.User{
			Id:         user.Id,
			Name:       name,
			Email:      email,
			ProfilePic: objectKey,
			Role:       role,
			APIkey:     apiKey,
		}
		// Final response
		resp := &Response{
			Token: tokenString,
			User:  userRes,
		}
		return c.JSON(resp)
	}
}




func Login() fiber.Handler{
	return func(c *fiber.Ctx) error{
		
		var d struct{
			Email string `bson:"email" json:"email"`
			Password string `bson:"password" json:"password"`
		}

		if err := c.BodyParser(&d); err !=nil{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":"Invalid request body",
			})
		}
		if d.Email==""||d.Password==""{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":"Email and Password are required",
			})
		}

		
		var user models.User
		err:= connect.UsersCollection.FindOne(context.TODO(), bson.M{"email":d.Email}).Decode(&user)
		if err!=nil{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":"NO user with this email",
			})
		}
		fmt.Println("user: ", user)
		fmt.Println("id in login :", user.Id)
		pass := d.Password
		password := []byte(pass)
		err = bcrypt.CompareHashAndPassword([]byte(user.Password),password )
		if err!=nil{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":"Password is incorrect, please try once more",
			})
		}
		tokenString, err := CreateToken(user.Id)
		if err!=nil{
			log.Println("failed to create token")
		}
		c.Cookie(&fiber.Cookie{
			Name: "token",
			Value:tokenString,
			HTTPOnly: true,
			Secure: true,
			Path:"/",
			MaxAge: 1000 * 60 * 60 * 24 * 5,
		})
		resp := &Response{
			Token:tokenString,
			User:user,
		}
		return c.JSON(resp)

	}
}

func Logout() fiber.Handler{
	return func(c *fiber.Ctx) error{
		c.Cookie(&fiber.Cookie{
			Name:     "token",
			Value:    "",
			Expires:  time.Now().Add(-time.Hour), // expire immediately
			HTTPOnly: true,
			Secure:   true,
			SameSite: "None",
		})
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Logged out successfully",
		})
	}
}




func GetUser() fiber.Handler{
	return func(c *fiber.Ctx) error {

		user_id,err:= FetchUserId(c)
		if err!=nil{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":"userId cannot be fetched",
			})
		}
		fmt.Println("userid: ", user_id)
		var user models.User
		err = connect.UsersCollection.FindOne(context.TODO(), bson.M{"id":user_id}).Decode(&user)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(user)
	}
}

func DeleteUser() fiber.Handler{
	return func(c *fiber.Ctx) error{
		user_id, err := FetchUserId(c)
		if err!=nil{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":"userId cannot be fetched",
			})
		}

		id,err:= primitive.ObjectIDFromHex(user_id)
		if err!=nil{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":"id format is not valid",
			})
		}
		result, err:= connect.UsersCollection.DeleteOne(context.TODO(), bson.M{"id":id})
		if err!=nil{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":"no user could be deleted",
			})
		}
		return c.JSON(result)
	}
}

func setUser(upd *UserResponse) (bson.M, error){
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

func UpdateUser() fiber.Handler{
	return func(c *fiber.Ctx) error{ 
		user_id, err := FetchUserId(c)
		if err!=nil{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":"userId cannot be fetched",
			})
		}

		id,err:= primitive.ObjectIDFromHex(user_id)
		if err!=nil{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":"id format is not valid",
			})
		}
		var userput UserResponse
		if err = c.BodyParser(&userput) ; err!= nil{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":"invalid JSON",
			})
		}

		setUser, err := setUser(&userput)
		if err!=nil{
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":"Failed to prepare update",
			})
		}
		if len(setUser)==0{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":"No fields provided to update",
			})
		}

		filter:= bson.M{
			"id":id,
		}
		update:=bson.M{
			"$set":setUser,
		}
		
		opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
		var updatedUser models.User

		err = connect.UsersCollection.FindOneAndUpdate(context.TODO(), filter, update, opts).Decode(&updatedUser)
		if err!=nil{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":"no user could be updae",
			})
		}
		return c.JSON(updatedUser)
	}
}


// TODO: Delete user and Update user
// getting user details by email id
