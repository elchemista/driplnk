package storage

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// TestS3Connectivity tests basic S3 operations with current credentials
// Run with: go test -v ./internal/adapters/storage -run TestS3Connectivity
func TestS3Connectivity(t *testing.T) {
	bucket := os.Getenv("S3_BUCKET")
	region := os.Getenv("S3_REGION")

	if bucket == "" || region == "" {
		t.Skip("S3_BUCKET or S3_REGION not set, skipping S3 test")
	}

	ctx := context.Background()

	// Load AWS config
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		t.Fatalf("Failed to load AWS config: %v", err)
	}

	client := s3.NewFromConfig(cfg)

	// Print debug info
	fmt.Printf("\n=== S3 Test Configuration ===\n")
	fmt.Printf("Bucket: %s\n", bucket)
	fmt.Printf("Region: %s\n", region)
	fmt.Printf("AWS_ACCESS_KEY_ID set: %v\n", os.Getenv("AWS_ACCESS_KEY_ID") != "")
	fmt.Printf("AWS_SECRET_ACCESS_KEY set: %v\n", os.Getenv("AWS_SECRET_ACCESS_KEY") != "")
	fmt.Printf("=============================\n\n")

	// Test 1: List objects in bucket (to verify read access)
	t.Run("ListObjects", func(t *testing.T) {
		result, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
			Bucket:  aws.String(bucket),
			MaxKeys: aws.Int32(5),
		})
		if err != nil {
			t.Fatalf("Failed to list objects: %v", err)
		}
		fmt.Printf("Listed %d objects in bucket\n", len(result.Contents))
		for _, obj := range result.Contents {
			fmt.Printf("  - %s\n", *obj.Key)
		}
	})

	// Test 2: Upload a test file to avatars folder
	t.Run("UploadToAvatars", func(t *testing.T) {
		testKey := fmt.Sprintf("avatars/test_%d.txt", time.Now().Unix())
		testContent := []byte("test upload from driplnk")

		_, err := client.PutObject(ctx, &s3.PutObjectInput{
			Bucket:      aws.String(bucket),
			Key:         aws.String(testKey),
			Body:        bytes.NewReader(testContent),
			ContentType: aws.String("text/plain"),
		})
		if err != nil {
			t.Fatalf("Failed to upload to avatars folder: %v", err)
		}
		fmt.Printf("Successfully uploaded: %s\n", testKey)

		// Cleanup - delete the test file
		_, err = client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(testKey),
		})
		if err != nil {
			t.Logf("Warning: Failed to delete test file: %v", err)
		} else {
			fmt.Printf("Successfully deleted: %s\n", testKey)
		}
	})

	// Test 3: Upload to root (like backup does)
	t.Run("UploadToRoot", func(t *testing.T) {
		testKey := fmt.Sprintf("test_root_%d.txt", time.Now().Unix())
		testContent := []byte("test upload to root")

		_, err := client.PutObject(ctx, &s3.PutObjectInput{
			Bucket:      aws.String(bucket),
			Key:         aws.String(testKey),
			Body:        bytes.NewReader(testContent),
			ContentType: aws.String("text/plain"),
		})
		if err != nil {
			t.Fatalf("Failed to upload to root: %v", err)
		}
		fmt.Printf("Successfully uploaded to root: %s\n", testKey)

		// Cleanup
		client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(testKey),
		})
	})
}
