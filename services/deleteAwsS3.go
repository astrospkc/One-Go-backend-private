package services

import (
	"context"
	"gobackend/env"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func DeleteFromS3(bucket, key string) error {
	envs := env.NewEnv()

	accessKey := envs.AWS_ACCESS_KEY_ID
	secretKey :=envs.AWS_SECRET_ACCESS_KEY

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		log.Fatal("failed to load config:", err)
	}
	

	client := s3.NewFromConfig(cfg)
	presignClient := s3.NewPresignClient(client)

	
	pre_url,err:=presignClient.PresignDeleteObject(context.TODO(),&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:aws.String(key),
	})
	if err!=nil{
		log.Fatal("failed to delete:", err)
	}

	request, err := http.NewRequest(http.MethodDelete, pre_url.URL, nil)
	if err !=nil{
		return err
	}
	resp, err := http.DefaultClient.Do(request)
	if err!=nil{
		return err
	}
	
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
	log.Fatalf("Delete failed: %s", resp.Status)
	}

	
	return nil
}