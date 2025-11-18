package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gobackend/connect"
	"gobackend/env"
	"gobackend/models"
	"gobackend/services"
	"io"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/resend/resend-go/v3"

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
	OTP         string		`bson:"otp,omitempty" json:"otp"`
	OTPVerification string `bson:"otpVerification,omitempty" json:"otpVerification"`
	
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

type pendingUser struct {
	Name         string `json:"name"`
	Email        string `json:"email"`
	PasswordHash string `json:"password_hash"`
	Role         string `json:"role,omitempty"`
	ProfileKey   string `json:"profile_key,omitempty"` // temp S3 object key (if uploaded)
	OTP          string `json:"otp,omitempty"`
	CreatedAt    int64  `json:"created_at"`
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
	return tokenString, nil
}


func generateOtp() string{
	rand.Seed(time.Now().UnixNano())
	generatedOtp:=fmt.Sprintf("%06d",rand.Intn(900000)+100000)
	return generatedOtp
}

func SendOtp() fiber.Handler{
	return func(c *fiber.Ctx)error{
		envs:=env.NewEnv()
		resendApiKey:=envs.RESEND_API_KEY
		name:=c.FormValue("name")
		email:=c.FormValue("email")
		password:=c.FormValue("password")
		role:=c.FormValue("role")

		if email==""||password==""{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error":"email and password required"})
		}

		// check existing user
		var existing models.User
		err:=connect.UsersCollection.FindOne(context.TODO(), primitive.M{"email":email}).Decode(&existing)
		if err==nil{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error":"email already in use"})
		}

		// handle photo upload
		var imagekey string
		picHeader,_:=c.FormFile("file")
		if picHeader!=nil{
			file,err:=picHeader.Open()
			if err!=nil{
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error":"Failed to open uploaded file"})
			}
			defer file.Close()
			var buf bytes.Buffer
			if _,err := io.Copy(&buf, file); err!=nil{
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error":"failed to buffer uploaded file",
				})
			}
			filename:=picHeader.Filename
			ext:=strings.TrimPrefix(strings.ToLower(filepath.Ext(filename)),".")
			mimeType:=picHeader.Header.Get("Content-Type")
			imagekey:=fmt.Sprintf("temo/pic_%s.%s", time.Now().Format("20060102_150405"), ext)
			_,err = services.CreatePresignedUrlAndUploadObject(os.Getenv("S3_BUCKET_NAME"),imagekey,buf.Bytes(),mimeType)
			if err!=nil{
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error":"failed to upload profle picture"})
			}
		}

		// hashed password now and keep hashed in pending record
		hashedPass, err:=bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err!=nil{
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error":"failed to hash password"})
		}

		// generate otp and pending user record
		otp:=generateOtp()
		pending:=pendingUser{
			Name: name,
			Email: email,
			PasswordHash: string(hashedPass),
			Role: role,
			ProfileKey: imagekey,
			OTP: otp,
			CreatedAt: time.Now().Unix(),	
		}
		key:=fmt.Sprintf("pending_user:%s",strings.ToLower(email))
		b,_:=json.Marshal(pending)
		if err:=connect.RedisClient.Do(connect.Rctx ,connect.RedisClient.B().Set().Key(key).Value(string(b)).Ex(10*time.Minute).Build()).Error();err!=nil{
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error":"failed to store otp"})
		}

		// send otp
		resendClient:= resend.NewClient(resendApiKey)
		htmlBody := fmt.Sprintf(`<p>Your One-Go verification code is: <strong>%s</strong></p><p>This code expires in 10 minutes.</p>`, otp)
		params:=&resend.SendEmailRequest{
			From: "One-Go <no-reply@xastrosbuild.site>",
			To:[]string{email},
			Html:htmlBody,
			Subject: " Your One-Go verification code",
		}
		_,err =resendClient.Emails.Send(params)
		fmt.Println("params: ", params, err)
		if err!=nil{
			_=connect.RedisClient.B().CfDel().Key(key)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error":"failed to send Otp email"})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "OTP sent to email",
			"email":   email,
		})

	}
}

func VerifyOTP() fiber.Handler{
	return func(c*fiber.Ctx) error{
		envs:=env.NewEnv()
		resendApiKey:=envs.RESEND_API_KEY
		_=resendApiKey

		type bodyReq struct{
			Email  string  `json:"email"`
			Otp    string   `json:"otp"`
		}

		var req bodyReq
		if err:=c.BodyParser(&req); err!=nil{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error":"invalid request"})

		}

		email:=strings.TrimSpace(req.Email)
		otp:=strings.TrimSpace(req.Otp)
		if email==""||otp==""{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error":"email and otp must be filled"})
		}

		// fetch pending user from redis
		key:=fmt.Sprintf("pending_user:%s",strings.ToLower(email))
		val,err:=connect.RedisClient.Do(connect.Rctx,connect.RedisClient.B().Get().Key(key).Build()).ToString()
		if val==""{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error":"otp is not found, please insert correct otp"})
		}else if err !=nil{
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error":"valkey error"})
		}

		var pending pendingUser
		if err := json.Unmarshal([]byte(val), &pending); err!=nil{
			_=connect.RedisClient.B().CfDel().Key(key)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error":"internal error"})
		}

		// validate otp
		if pending.OTP!=otp{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error":"invalid otp"})
		}

		// valid -> create new user
		apikey, err:=GenerateApiKey()
		if err!=nil{
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error":"Failed to generate api key"})
		}

		user := models.User{
			Id:primitive.NewObjectID().Hex(),
			Name:pending.Name,
			Email:pending.Email,
			Password: pending.PasswordHash,
			ProfilePic: pending.ProfileKey,
			Role:pending.Role,
			APIkey: apikey,
			OTP: "",
			OTPVerification: "verified",
		}

		_,err = connect.UsersCollection.InsertOne(context.TODO(),user)
		if err!=nil{
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error":"failed to create user"})
		}

		// create api key record
		apiDoc:=models.APIkey{
			Id:primitive.NewObjectID().Hex(),
			UserId: user.Id,
			Key:apikey,
			UsageLimit: 50,
		}
		
		_,err = connect.APIKeyCollection.InsertOne(context.TODO(), apiDoc)
	
		if err!=nil{
			// rolling back user insertion in case api key creation failed
			_, _= connect.UsersCollection.DeleteOne(context.TODO(), bson.M{"id":user.Id})
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error":"failed to save api key"})
		}

		// 
		tokenString, err := CreateToken(user.Id)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create token",
			})
		}
		// set secure cookie 
		c.Cookie(&fiber.Cookie{
			Name:     "token",
			Value:    tokenString,
			HTTPOnly: true,
			Secure:   true,
			Path:     "/",
			MaxAge:   1000 * 60 * 60 * 24 * 5,
		})

		// remove pending record from redis
		_=connect.RedisClient.B().CfDel().Key(key)

		userRes := models.User{
			Id:         user.Id,
			Name:       user.Name,
			Email:      user.Email,
			ProfilePic: user.ProfilePic,
			Role:       user.Role,
			APIkey:     user.APIkey,
		}

		resp := struct {
			Message string      `json:"message"`
			Token   string      `json:"token"`
			User    models.User `json:"user"`
		}{
			Message: "verified and user created",
			Token:   tokenString,
			User:    userRes,
		}

		return c.Status(fiber.StatusOK).JSON(resp)
	}
}

type ForgotPasswordResponse struct {
	Message string `json:"message"`
	Code int `json:"code"`
}

func ForgotPassword() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// parse env
		envs := env.NewEnv()
		resendApiKey := envs.RESEND_API_KEY

		// parse request body
		type bodyReq struct {
			Email string `json:"email"`
		}

		var req bodyReq
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(
				ForgotPasswordResponse{
					Message: "Invalid request body",
					Code:    400,
				},
			)
		}

		// generate OTP
		otp := generateOtp()

		// send email
		resendClient := resend.NewClient(resendApiKey)
		htmlBody := fmt.Sprintf(
			`<p>Your One-Go verification code is: <strong>%s</strong></p><p>This code expires in 10 minutes.</p>`,
			otp,
		)

		params := &resend.SendEmailRequest{
			From:    "One-Go <no-reply@xastrosbuild.site>",
			To:      []string{req.Email},
			Html:    htmlBody,
			Subject: "Your One-Go verification code",
		}

		_, err := resendClient.Emails.Send(params)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(
				ForgotPasswordResponse{
					Message: "Failed to send email",
					Code:    500,
				},
			)
		}

		// success
		return c.Status(fiber.StatusOK).JSON(
			ForgotPasswordResponse{
				Message: "OTP sent to email",
				Code:    200,
			},
		)
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
