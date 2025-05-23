package services

import (
	"bytes"
	"context"
	"fmt"
	"gobackend/env"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func CreatePresignedUrlAndUploadObject(bucketName string , objectKey string,data[]byte, contentType string) (string, error){
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

	params := &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	}
	// handle progress operation also
	// this is generating presigned url
	presignedURL, err := presignClient.PresignPutObject(context.TODO(), params, func(opts *s3.PresignOptions) {
		opts.Expires = time.Hour // expires in 1 hour
	})
	if err != nil {
		log.Fatal("failed to generate presigned URL:", err)
	}

	// now put operation to upload it to s3
	httpreq, err:= http.NewRequest("PUT",presignedURL.URL, bytes.NewReader(data))
	if err!=nil{
		return "",err
	}

	httpreq.Header.Set("Content-Type",contentType)
	clientHttp:= &http.Client{}
	resp,err:=clientHttp.Do(httpreq)
	if err!=nil{
		return "",err 
	}
	defer resp.Body.Close()

	if resp.StatusCode !=http.StatusOK{
		body,_ := io.ReadAll(resp.Body)
		return "",fmt.Errorf("error while uploading %s %s", resp.Status, body)
	}




	// fmt.Println("Presigned URL:", presignedURL.URL)
	return  presignedURL.URL, nil
}

// here I am gettng presignedUrl , which is letting me to upload file 
