package controller

import (
	"context"
	"fmt"
	"gobackend/connect"
	"log"
	"reflect"

	"go.mongodb.org/mongo-driver/bson/primitive"
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

func GetUserViaId(user_id string) (UserResponse, error)  {
	fmt.Println("user_id: ", user_id)
	id,err:= primitive.ObjectIDFromHex(user_id)
	if err!=nil{
		log.Fatal("id format invalid")
	}
	var foundUser UserResponse
	err = connect.UsersCollection.FindOne(context.TODO(), bson.M{"id":id}).Decode(&foundUser)
	if err!=nil{
		log.Fatal("No such id exist")
	}
		fmt.Println(reflect.TypeOf(foundUser), foundUser)
	return foundUser, err
	
}
