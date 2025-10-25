package controller

import (
	"context"
	"fmt"
	"gobackend/connect"
	"log"
	"reflect"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// func GetClaimsEmail(user interface{}) (string, error) {
// 	claims, ok := user.(jwt.MapClaims)
// 	if !ok {
// 		return "", errors.New("invalid JWT claims format")
// 	}

// 	email, ok := claims["aud"].(string)
// 	if !ok {
// 		return "", errors.New("invalid or missing 'aud' field")
// 	}

// 	return email, nil
// }
func FetchUserId(c *fiber.Ctx) (string , error){
	
		var user_id string
		userIdInterface:=c.Locals("user_id")
		// fmt.Println("user interfacce: ", userIdInterface)
		user_id, ok := userIdInterface.(string)
		if !ok {
			if user_id=="" {
				user := c.Locals("user")
				claims,ok := user.(jwt.MapClaims)
				if !ok {
					return "",c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
						"error": "Invalid or missing  aud field",
					})
				}
				user_id, ok = claims["aud"].(string)
				if !ok {
					return "",c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
						"error": "Invalid or missing  aud field",
					})
				}
				}
		
				fmt.Println("user_id in fn", user_id)
				
				return user_id, nil
		}
		return user_id, nil
		
		
		
		
}
func GetUserViaId(user_id string) (UserResponse, error)  {
	fmt.Println("user_id: ", user_id)
	
	
	var foundUser UserResponse
	err := connect.UsersCollection.FindOne(context.TODO(), bson.M{"id":user_id}).Decode(&foundUser)
	if err!=nil{
		log.Fatal("No such id exist")
	}
		fmt.Println(reflect.TypeOf(foundUser), foundUser)
	return foundUser, err
	
}


