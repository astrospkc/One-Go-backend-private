package controller

import (
	"context"
	"fmt"
	"gobackend/connect"
	"gobackend/models"
	"log"
	"os"

	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
	"golang.org/x/crypto/bcrypt"
)

// TODO: later on add Project , category , links, blog, media, resume, subscription, usersubscription, apikey , all of these in UserResponse
type UserResponse struct {
	Id       primitive.ObjectID `bson:"id,omitempty" json:"id"`
	Name 		string `bson:"name" json:"name"`
	Email 		string `bson:"email" json:"email"`
	ProfilePic  string `bson:"profile_pic,omitempty" json:"profile_pic"`
	Role 		string	`bson:"role" json:"role"`
	APIkey		string  `bson:"api_key" json:"api_key"`
	
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

// var secretKey = []byte(os.Getenv("JWT_SECRET"))


// first createtoken
func CreateToken(userid string) (string, error){
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userid,                    // Subject (user identifier)
		"iss": "One-Go",                  // Issuer
		"aud": userid,           // Audience (user role)
		"exp": time.Now().Add(time.Hour).Unix(), // Expiration time
		"iat": time.Now().Unix(),                 // Issued at
	})
	secret :=[]byte(os.Getenv("JWT_SECRET"))
	tokenString, err := claims.SignedString(secret)
	if err!=nil{
		return "", err
	}
	fmt.Printf("Token claims added: %+v\n", claims)
	return tokenString, nil
}




func CreateUser() fiber.Handler {
	return func(c *fiber.Ctx) error {

		apikey,err:= GenerateApiKey()
		if err!=nil{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":"Failed to generate key",
			})
		}

    	var d models.User
    	if err := c.BodyParser(&d); err!=nil{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":"Invalid request body",
			})
		}

		
		if d.Email == "" || d.Password == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Email and Password are required",
			})
		}


		// hashing the password
		password:=[]byte(d.Password)
		hashedPass, err := bcrypt.GenerateFromPassword(password,bcrypt.DefaultCost)
		if err!=nil{
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Could not hash password",
			})
		}
		fmt.Println(string(hashedPass))


		if d.ProfilePic == "" {
			d.ProfilePic = "https://cdn.example.com/default-avatar.png"
		}
		hash:= string(hashedPass)
		
		user := models.User{
			Id:primitive.NewObjectID(),
			Name:d.Name,
			Email:d.Email,
			ProfilePic: d.ProfilePic,
			Password:hash ,
			Role:d.Role,
			APIkey: apikey,
		}
		fmt.Println("user: ", user)
		_,err = connect.UsersCollection.InsertOne(context.Background(), user)
		if err!=nil{
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "looks like email address is already in use",
			})
			
		}
		
		
		tokenString,err := CreateToken(user.Id.Hex())
		if err!=nil{
			log.Println("failed to create token")
		}
		
    	c.Cookie(&fiber.Cookie{
			Name: "token",
			Value:tokenString,
			HTTPOnly: true,
			Secure: true,
			Path:"/",
			MaxAge: 3600,
		})

		_ = models.APIkey{
			Id: primitive.NewObjectID(),
			Userid: user.Id.Hex(),
			Key: apikey,
			UsageLimit: 50,
		}

		
		
		resp := &Response{
			Token:tokenString,
			User:user,
		}
		return c.JSON(resp)
	}
}


func Login() fiber.Handler{
	return func(c *fiber.Ctx) error{
		
		var d models.User
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
		tokenString, err := CreateToken(user.Id.Hex())
		if err!=nil{
			log.Println("failed to create token")
		}
		c.Cookie(&fiber.Cookie{
			Name: "token",
			Value:tokenString,
			HTTPOnly: true,
			Secure: true,
			Path:"/",
			MaxAge: 3600,
		})
		resp := &Response{
			Token:tokenString,
			User:user,
		}
		return c.JSON(resp)

	}
}


func GetUser() fiber.Handler{
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
		
		user_info,err := GetUserViaId(user_id)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "User not found"})
		}
		return c.JSON(user_info)

		
	}
}



// getting user details by email id
