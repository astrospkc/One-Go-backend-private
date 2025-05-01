package services

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joho/godotenv"
)

func CreatePresignedUrlAndUploadObject(bucketName string , key string) (string, error){
	err := godotenv.Load(".env.prod") // or ".env.prod"
	if err != nil {
		log.Fatal("Error loading env file")
	}

	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		log.Fatal("failed to load config:", err)
	}

	client := s3.NewFromConfig(cfg)

	presignClient := s3.NewPresignClient(client)

	params := &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
		ACL: "public-read",
	
	}

	presignedURL, err := presignClient.PresignPutObject(context.TODO(), params, func(opts *s3.PresignOptions) {
		opts.Expires = time.Hour // expires in 1 hour
	})
	if err != nil {
		log.Fatal("failed to generate presigned URL:", err)
	}

	fmt.Println("Presigned URL:", presignedURL.URL)
	return presignedURL.URL, nil
}
