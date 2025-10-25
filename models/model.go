package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	Id 			string	`bson:"id,omitempty" json:"id"`
	Google_id   string  `bson:"google_id" json:"google_id"`
	Name 		string `bson:"name" json:"name"`
	Email 		string `bson:"email" json:"email"`
	ProfilePic  string `bson:"profile_pic,omitempty" json:"profile_pic"`
	Password 	string `bson:"password" json:"password"`
	Role 		string	`bson:"role" json:"role"`
	APIkey		string  `bson:"api_key" json:"api_key"`
	
}

type Collection struct{
	Id      string `bson:"id,omitempty" json:"id"`
	UserId	string 	`bson:"user_id" json:"user_id"`
	Title 		string `bson:"title" json:"title"`
	Description	string `bson:"description" json:"description"`
	CreatedAt	time.Time	`bson:"time" json:"time"`
}

type Project struct{
	Id           string `bson:"id,omitempty" json:"id"`
	UserId		 string	`bson:"user_id" json:"user_id"`
	CollectionId string	`bson:"collection_id" json:"collection_id"`
	Title		 string	`bson:"title" json:"title"`
	Description	 string	`bson:"description,omitempty" json:"description"`
	Tags		 string	`bson:"tags,omitempty" json:"tags"`
	FileUpload   string `bson:"fileUpload,omitempty" json:"fileUpload"`
	Thumbnail 	 string	`bson:"thumbnail,omitempty" json:"thumbnail"`
	GithubLink	 string	`bson:"githublink,omitempty" json:"githublink"`
	DemoLink     string `bson:"demolink,omitempty" json:"demolink"`
	LiveDemoLink string	`bson:"livedemolink,omitempty" json:"livedemolink"`
	BlogLink     string `bson:"blogLink,omitempty" json:"blogLink"`
	TeamMembers  string `bson:"teamMembers,omitempty" json:"teamMembers"`
	CreatedAt	time.Time	`bson:"time" json:"time"`
	UpdatedAt   time.Time  `bson:"updated_time" json:"updated_time"`
}

type Category struct {
	CollectionId primitive.ObjectID		`bson:"collection_id" json:"collection_id"`
	Blogs   Blog	`bson:"blog" json:"blog"`
	Links	string	`bson:"links" json:"links"`
	Media	string	`bson:"media" json:"media"`
	Resume	string 	`bson:"resume" json:"resume"`
}

type Blog struct{
	Id           string `bson:"id,omitempty" json:"id"`
	UserId		 string 	`bson:"user_id" json:"user_id"`
	CollectionId string		`bson:"collection_id" json:"collection_id"`
	Title		string	`bson:"title" json:"title"`
	Description	string  `bson:"description" json:"description"`
	Content		map[string]interface{}	`bson:"content" json:"content"`
	Tags		string	`bson:"tags,omitempty" json:"tags"`
	CoverImage 	string	`bson:"coverImage,omitempty" json:"coverImage"`
	Published 	time.Time	`bson:"published" json:"published"`
	CreatedAt	time.Time	`bson:"time" json:"time"`
	LastEdited  time.Time   `bson:"lastedited" json:"lastedited"`
	Status		string 		`bson:"status" json:"status"`
}

type Media struct{
	Id 			string	`bson:"id,omitempty" json:"id"`
	UserId		 string 	`bson:"user_id" json:"user_id"`
	CollectionId string		`bson:"collection_id" json:"collection_id"`
	Key         string 			`bson:"key" json:"key"`
	// File        multipart.File 	`bson:"file" json:"file"`
	Title 		string	`bson:"title" json:"title"`
	Content		string	`bson:"content" json:"content"`
	CreatedAt	time.Time	`bson:"time" json:"time"`
}
// type (image, video, audio, doc, pdf, etc.)
type Link struct{
	Id 			string	`bson:"id,omitempty" json:"id"`
	UserId		 string 	`bson:"user_id" json:"user_id"`
	CollectionId string		`bson:"collection_id" json:"collection_id"`
	Source		string	`bson:"source,omitempty" json:"source"`
	Title		string	`bson:"title,omitempty" json:"title"`
	Url			string	`bson:"url" json:"url"` 
	Description	string	`bson:"description,omitempty" json:"description"`
	Category	string	`bson:"category,omitempty" json:"category"`
}
// category (e.g., Social, Project, Resume)

type Resume struct {
	Id 			string	`bson:"id,omitempty" json:"id"`
	UserId 		string	`bson:"user_id" json:"user_id"`
	CollectionId string		`bson:"collection_id" json:"collection_id"`
	FileUrl		string	`bson:"fileurl" json:"fileurl"`
	UploadData	time.Time	`bson:"uploadData" json:"uploadData"`
}

type SubscriptionPlan struct {
	Id 			primitive.ObjectID	`bson:"id,omitempty" json:"id"`
	Name		string		`bson:"name" json:"name"`
	Price 		string		`bson:"price" json:"price"`
	Features 	[]string	`bson:"features" json:"features"`
	Duration 	string		`bson:"duration" json:"duration"`
}

type UserSubscription struct {
	Id 			primitive.ObjectID	`bson:"id,omitempty" json:"id"`
	UserId		string 	`bson:"user_id" json:"user_id"`
	PlanId 		string		`bson:"plan_id" json:"plan_id"`
	StartDate	time.Time	`bson:"startDate" json:"startDate"`
	EndDate		time.Time	`bson:"endDate" json:"endDate"`
	Status		string		`bson:"status" json:"status"`
}

// status- active or expired
type APIkey struct{
	Id 			string	`bson:"id,omitempty" json:"id"`
	UserId		string	`bson:"user_id" json:"user_id"`
	CollectionId string		`bson:"collection_id" json:"collection_id"`
	Key 		string	`bson:"key" json:"key"`
	UsageLimit	int64	`bson:"usagelimit" json:"usagelimit"`
	CreatedAt	time.Time	`bson:"createdat" json:"createdat"`
	Revoked		bool	`bson:"revoked" json:"revoked"`
	
}