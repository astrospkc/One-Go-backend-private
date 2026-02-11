package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	Id              string `bson:"id,omitempty" json:"id"`
	Google_id       string `bson:"google_id" json:"google_id"`
	Name            string `bson:"name" json:"name"`
	Email           string `bson:"email" json:"email"`
	ProfilePic      string `bson:"profile_pic,omitempty" json:"profile_pic"`
	Password        string `bson:"password" json:"password"`
	Role            string `bson:"role" json:"role"`
	APIkey          string `bson:"api_key" json:"api_key"`
	OTP             string `bson:"otp,omitempty" json:"otp"`
	OTPVerification string `bson:"otpVerification,omitempty" json:"otpVerification"`
	Plan            string `bson:"plan,omitempty" json:"plan"`
}

type Collection struct {
	Id          string    `bson:"id,omitempty" json:"id"`
	UserId      string    `bson:"user_id" json:"user_id"`
	Title       string    `bson:"title" json:"title"`
	Description string    `bson:"description" json:"description"`
	CreatedAt   time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at" json:"updated_at"`
}

type Project struct {
	Id           string    `bson:"id,omitempty" json:"id"`
	UserId       string    `bson:"user_id" json:"user_id"`
	CollectionId string    `bson:"collection_id" json:"collection_id"`
	Title        string    `bson:"title" json:"title"`
	Description  string    `bson:"description,omitempty" json:"description"`
	Tags         string    `bson:"tags,omitempty" json:"tags"`
	FileUpload   []string  `bson:"fileUpload,omitempty" json:"fileUpload"`
	Thumbnail    string    `bson:"thumbnail,omitempty" json:"thumbnail"`
	GithubLink   string    `bson:"githublink,omitempty" json:"githublink"`
	DemoLink     string    `bson:"demolink,omitempty" json:"demolink"`
	LiveDemoLink string    `bson:"livedemolink,omitempty" json:"livedemolink"`
	BlogLink     string    `bson:"blogLink,omitempty" json:"blogLink"`
	TeamMembers  string    `bson:"teamMembers,omitempty" json:"teamMembers"`
	CreatedAt    time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time `bson:"updated_at" json:"updated_at"`
}

type Category struct {
	CollectionId primitive.ObjectID `bson:"collection_id" json:"collection_id"`
	Blogs        Blog               `bson:"blog" json:"blog"`
	Links        string             `bson:"links" json:"links"`
	Media        string             `bson:"media" json:"media"`
	Resume       string             `bson:"resume" json:"resume"`
}

type Blog struct {
	Id           string                 `bson:"id,omitempty" json:"id"`
	UserId       string                 `bson:"user_id" json:"user_id"`
	CollectionId string                 `bson:"collection_id" json:"collection_id"`
	Title        string                 `bson:"title" json:"title"`
	Description  string                 `bson:"description" json:"description"`
	Content      map[string]interface{} `bson:"content" json:"content"`
	Tags         string                 `bson:"tags,omitempty" json:"tags"`
	CoverImage   string                 `bson:"coverImage,omitempty" json:"coverImage"`
	Published    time.Time              `bson:"published" json:"published"`
	CreatedAt    time.Time              `bson:"created_at" json:"created_at"`
	LastEdited   time.Time              `bson:"lastedited" json:"lastedited"`
	Status       string                 `bson:"status" json:"status"`
}

type Media struct {
	Id           string `bson:"id,omitempty" json:"id"`
	UserId       string `bson:"user_id" json:"user_id"`
	CollectionId string `bson:"collection_id" json:"collection_id"`
	Key          string `bson:"key" json:"key"`
	// File        multipart.File 	`bson:"file" json:"file"`
	Title     string    `bson:"title" json:"title"`
	Content   string    `bson:"content" json:"content"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

// type (image, video, audio, doc, pdf, etc.)
type Link struct {
	Id           string `bson:"id,omitempty" json:"id"`
	UserId       string `bson:"user_id" json:"user_id"`
	CollectionId string `bson:"collection_id" json:"collection_id"`
	Source       string `bson:"source,omitempty" json:"source"`
	Title        string `bson:"title,omitempty" json:"title"`
	Url          string `bson:"url" json:"url"`
	Description  string `bson:"description,omitempty" json:"description"`
	Category     string `bson:"category,omitempty" json:"category"`
}

// category (e.g., Social, Project, Resume)


// status- active or expired
type APIkey struct {
	Id           string    `bson:"id,omitempty" json:"id"`
	UserId       string    `bson:"user_id" json:"user_id"`
	CollectionId string    `bson:"collection_id" json:"collection_id"`
	Key          string    `bson:"key" json:"key"`
	UsageLimit   int64     `bson:"usagelimit" json:"usagelimit"`
	CreatedAt    time.Time `bson:"createdat" json:"createdat"`
	Revoked      bool      `bson:"revoked" json:"revoked"`
}

type Subscription struct {
	Id          string    `bson:"id,omitempty" json:"id"`
	UserId      string    `bson:"user_id" json:"user_id"`
	Plan        string    `bson:"plan" json:"plan"`     // free, pro,team...
	Status      string    `bson:"status" json:"status"` // pending/active/canceled/expired
	StartAt     time.Time `bson:"start_at" json:"start_at"`
	EndAt       time.Time `bson:"end_at" json:"end_at"`
	AutoRenew   bool      `bson:"auto_renew" json:"auto_renew"`
	TrialEndsAt time.Time `bson:"trial_ends_at,omitempty" json:"trial_ends_at"`
	UpdatedAt   time.Time `bson:"updated_at" json:"updated_at"`
}

type SubscriptionPayment struct {
	Id              string	`bson:"id,omitempty" json:"id"`
	SubscriptionId  string	`bson:"subscription_id" json:"subscription_id"`
	UserId          string	`bson:"user_id" json:"user_id"`
	Amount          float64	`bson:"amount" json:"amount"`
	Currency        string	`bson:"currency" json:"currency"`
	Gateway         string	`bson:"gatewat" json:"gatewat"`
	PaymentRef      string	`bson:"payment_reference" json:"payment_reference"`
	Status          string	`bson:"status" json:"status"`
	PeriodStart     time.Time `bson:"period_start" json:"period_start"`
	PeriodEnd       time.Time `bson:"period_end" json:"period_end"`
	IdempotencyKey  string	`bson:"idempotency_key" json:"idempotency_key"`
	RefundAmount	float64	 `bson:"refund_amount" json:"refund_amount"`
	ExpiresAt		time.Time `bson:"expires_at" json:"expires_at"`
	CreatedAt       time.Time `bson:"created_at" json:"created_at"`
}

// Tracks storage & API usage for billing limits
type Usage struct {
	ID               string    `bson:"_id,omitempty" json:"id"`
	UserID           string    `bson:"user_id" json:"user_id"`
	StorageUsed      int64     `bson:"storage_used" json:"storage_used"`
	StorageLimit     int64     `bson:"storage_limit" json:"storage_limit"`
	APIRequestsCount int64     `bson:"api_requests_count" json:"api_requests_count"`
	APIRequestsLimit int64     `bson:"api_requests_limit" json:"api_requests_limit"`
	UpdatedAt        time.Time `bson:"updated_at" json:"updated_at"`
}

// Plan feature toggles
type FeatureAccess struct {
	ID                 string    `bson:"_id,omitempty" json:"id"`
	UserID             string    `bson:"user_id" json:"user_id"`
	ProjectsLimit      int       `bson:"projects_limit" json:"projects_limit"`
	TeamMembersAllowed int       `bson:"team_members_allowed" json:"team_members_allowed"`
	CustomDomain       bool      `bson:"custom_domain" json:"custom_domain"`
	AnalyticsEnabled   bool      `bson:"analytics_enabled" json:"analytics_enabled"`
	PrioritySupport    bool      `bson:"priority_support" json:"priority_support"`
	UpdatedAt          time.Time `bson:"updated_at" json:"updated_at"`
}


