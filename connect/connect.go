package connect

import (
	"context"
	"fmt"
	"gobackend/env"

	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	dbName  = "CMS_portfolio"
	colCollection = "collections"
	colNameUsers = "users"
	colNameProjects = "projects"
	colNameBlogs = "blogs"
	colNameLinks = "links"
	colNameMedia = "media"
	colNameSubscription = "subscription"
	colNameAPI  = "apikey"
)

var UsersCollection *mongo.Collection
var ColCollection    *mongo.Collection
var ProjectCollection *mongo.Collection
var BlogsCollection *mongo.Collection
var LinksCollection *mongo.Collection
var MediaCollection *mongo.Collection
var SubscriptionCollection *mongo.Collection

var APIKeyCollection *mongo.Collection


func Connect(){
	envs := env.NewEnv()
	var uri string 
	uri =envs.MONGODB_URI
	
	if uri = envs.MONGODB_URI; uri == "" {
		log.Fatal("You must set your 'MONGODB_URI' environment variable. See\n\t https://docs.mongodb.com/drivers/go/current/usage-examples/")
	}

	// Use the SetServerAPIOptions() method to set the Stable API version to 1
	ctx, cancel:= context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)

	// Create a new client and connect to the server
	client, err := mongo.Connect(opts)

	if err != nil {
		log.Fatal("the error while connecting mongo client", err)
		return
	}
	// defer func() {
	// 	fmt.Println("Disconnecting")
	// 	if err = client.Disconnect(context.TODO()); err != nil {
	// 		log.Fatal("error while disconnecting: ", err)
	// 		return
	// 	}
	// }()
	
	err = client.Ping(ctx, nil)
	
	if err != nil{
		log.Fatal("ping :", err)
		return
	}
	UsersCollection = client.Database(dbName).Collection(colNameUsers)
	indexModel:=mongo.IndexModel{
		Keys:bson.D{{Key:"email", Value:  -1}},
		Options:options.Index().SetUnique(true),
	}
	_, err =UsersCollection.Indexes().CreateOne(context.TODO(), indexModel)
	if err != nil{
		
		log.Fatal("error occured while connecting to users collection : ", err)
	}
	
	ColCollection  = client.Database(dbName).Collection(colCollection)
	indexModel =mongo.IndexModel{
		Keys: bson.D{
			{Key:"user_id",Value:1},
			{Key:"title", Value: 1},
		},
		Options:options.Index().SetUnique(true),
	}
	_, err = ColCollection.Indexes().CreateOne(context.TODO(),indexModel)
	if err!=nil{
		log.Fatal("error occurred while connecting to collection collection: ", err)
	}

	ProjectCollection = client.Database(dbName).Collection(colNameProjects)
	indexModel =mongo.IndexModel{
		Keys: bson.D{
			
			{Key:"user_id",Value:1},
			{Key:"title", Value: 1},
		},
		Options:options.Index().SetUnique(true),
	}
	_, err = ProjectCollection.Indexes().CreateOne(context.TODO(),indexModel)
	if err!=nil{
		log.Fatal("error occurred while connecting to project collection: ", err)
	}
	BlogsCollection = client.Database(dbName).Collection(colNameBlogs)
	LinksCollection = client.Database(dbName).Collection(colNameLinks)
	MediaCollection = client.Database(dbName).Collection(colNameMedia)
	APIKeyCollection = client.Database(dbName).Collection(colNameAPI)
	SubscriptionCollection = client.Database(dbName).Collection(colNameSubscription)

	db_collection := []*mongo.Collection{ProjectCollection,BlogsCollection,LinksCollection,APIKeyCollection,ColCollection,UsersCollection,MediaCollection,SubscriptionCollection}
	im := mongo.IndexModel{
		Keys:bson.D{
			{Key:"id", Value: -1},
		},
	}

	for _, val := range db_collection {
		_, err = val.Indexes().CreateOne(context.TODO(), im)
		if err!=nil{
			log.Fatal("error occurred while connecting to index: ", err)
		}
	}


	fmt.Println("Set up is done")
	
}