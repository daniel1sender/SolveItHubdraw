package utils

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func UploadFileToS3(filePath, bucket, key string) error {
	// Load the AWS SDK configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(os.Getenv("AWS_REGION")))
	if err != nil {
		return fmt.Errorf("unable to load AWS SDK config: %v", err)
	}

	// Create an S3 client
	client := s3.NewFromConfig(cfg)

	// Open the file to upload
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("unable to open file %s: %v", filePath, err)
	}
	defer file.Close()

	// Upload the file to S3
	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   file,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file to S3: %v", err)
	}

	fmt.Println("File uploaded successfully!")
	return nil
}

func ListFilesInS3(bucket string) ([]string, error) {
	// Load AWS SDK configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %v", err)
	}

	// Create an S3 client
	client := s3.NewFromConfig(cfg)

	// List objects in the bucket
	resp, err := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list objects in bucket: %v", err)
	}

	// Collect file names
	var files []string
	for _, item := range resp.Contents {
		files = append(files, *item.Key)
	}

	return files, nil
}

func DownloadFileFromS3(bucket, filename string) (io.ReadCloser, error) {
	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(os.Getenv("AWS_REGION")))
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config: %v", err)
	}

	// Create an S3 client
	client := s3.NewFromConfig(cfg)

	// Get the file from S3
	resp, err := client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(filename),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to download file from S3: %v", err)
	}

	return resp.Body, nil
}