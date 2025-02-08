package storage

import (
	"fmt"
	"log"
	"os"

	"github.com/IBM/ibm-cos-sdk-go/aws"
	"github.com/IBM/ibm-cos-sdk-go/aws/credentials"
	"github.com/IBM/ibm-cos-sdk-go/aws/session"
	"github.com/IBM/ibm-cos-sdk-go/service/s3"
)

type COSConfig struct {
	APIKey            string
	ServiceInstanceID string
	BucketName        string
	Endpoint          string
	Location          string
}

func loadCOSConfig() (*COSConfig, error) {
	config := &COSConfig{
		APIKey:            os.Getenv("IBM_COS_API_KEY"),
		ServiceInstanceID: os.Getenv("IBM_COS_INSTANCE_ID"),
		Endpoint:          os.Getenv("IBM_COS_ENDPOINT"),
		BucketName:        os.Getenv("IBM_COS_BUCKET_NAME"),
		Location:          os.Getenv("IBM_COS_LOCATION"),
	}

	// Validate required fields
	if config.APIKey == "" || config.ServiceInstanceID == "" || config.Endpoint == "" || config.BucketName == "" {
		return nil, fmt.Errorf("missing required COS configuration. Please ensure all required environment variables are set")
	}

	return config, nil
}

func uploadToCOS(localPath, objectKey string) error {
	config, err := loadCOSConfig()
	if err != nil {
		return fmt.Errorf("failed to load COS config: %w", err)
	}

	// Create a new session using IBM COS credentials
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(
			config.APIKey,
			config.ServiceInstanceID,
			"", // No session token needed
		),
		Endpoint:         aws.String(config.Endpoint),
		Region:           aws.String(config.Location),
		S3ForcePathStyle: aws.Bool(true),
	})
	if err != nil {
		return fmt.Errorf("failed to create COS session: %w", err)
	}

	// Create S3 service client
	svc := s3.New(sess)

	// Open the local file
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %w", err)
	}
	defer file.Close()

	// Get file size
	_, err = file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	// Upload the file
	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(config.BucketName),
		Key:    aws.String(objectKey),
		Body:   file,
		ACL:    aws.String("private"),
	})
	if err != nil {
		return fmt.Errorf("failed to upload to COS: %w", err)
	}

	log.Printf("Successfully uploaded %s to COS bucket %s", objectKey, config.BucketName)

	// Delete local file after successful upload
	if err := os.Remove(localPath); err != nil {
		log.Printf("Warning: Failed to delete local file %s: %v", localPath, err)
	}

	return nil
}
