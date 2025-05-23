package services

import (
	"context"
	"fmt"
	"gobackend/env"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	// "github.com/joho/godotenv"
)

// func init() {
// 	err := godotenv.Load()
// 	if err != nil {
// 		log.Fatal("Error loading .env file")
// 	}
// }


func GetPresignedGetUrl(bucketName, fileKey string) (string, error) {
	fmt.Println("bucket name and filekey: ", bucketName, fileKey)
	envs:= env.NewEnv()

	accessKey := envs.AWS_ACCESS_KEY_ID
	secretKey :=envs.AWS_SECRET_ACCESS_KEY
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				accessKey, secretKey,
				"",
			),
		),
	)
	if err != nil {
		return "", fmt.Errorf("unable to load SDK config, %v", err)
	}

	client := s3.NewFromConfig(cfg)
	presignClient := s3.NewPresignClient(client)

	resp, err := presignClient.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(fileKey),
	})
	// , s3.WithPresignExpires(time.Duration(expiresInSeconds)*time.Second)

	if err != nil {
		return "", fmt.Errorf("failed to sign request: %v", err)
	}

	return resp.URL, nil
}


