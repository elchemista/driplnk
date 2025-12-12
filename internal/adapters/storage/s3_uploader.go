package storage

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"io"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/chai2010/webp"
	"github.com/disintegration/imaging"
)

type S3Uploader struct {
	client     *s3.Client
	uploader   *manager.Uploader
	bucket     string
	region     string
	cdnURL     string
	folderPath string
}

type S3UploaderConfig struct {
	Bucket     string
	Region     string
	CDNURL     string // Optional: If set, used to generate public URLs
	FolderPath string // Optional: Prefix for uploaded files (e.g., "media")
}

func NewS3Uploader(ctx context.Context, cfg S3UploaderConfig) (*S3Uploader, error) {
	awsCfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(cfg.Region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg)
	uploader := manager.NewUploader(client)

	return &S3Uploader{
		client:     client,
		uploader:   uploader,
		bucket:     cfg.Bucket,
		region:     cfg.Region,
		cdnURL:     cfg.CDNURL,
		folderPath: cfg.FolderPath,
	}, nil
}

func (u *S3Uploader) Upload(ctx context.Context, file io.Reader, filename string) (string, error) {
	key := filename
	if u.folderPath != "" {
		key = path.Join(u.folderPath, filename)
	}

	_, err := u.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(u.bucket),
		Key:    aws.String(key),
		Body:   file,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to s3: %w", err)
	}

	return key, nil
}

// UploadProfileImage uploads a profile image, resizing it and converting to WebP.
func (u *S3Uploader) UploadProfileImage(ctx context.Context, file io.Reader, filename string) (string, error) {
	// 1. Decode image
	img, _, err := image.Decode(file)
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	// 2. Resize to 500x500 (Cover) - Smart cropping
	resized := imaging.Fill(img, 500, 500, imaging.Center, imaging.Lanczos)

	// 3. Encode to WebP
	var buf bytes.Buffer
	if err := webp.Encode(&buf, resized, &webp.Options{Lossless: false, Quality: 80}); err != nil {
		return "", fmt.Errorf("failed to encode webp: %w", err)
	}

	// 4. Determine new filename (replace extension with .webp)
	ext := path.Ext(filename)
	base := strings.TrimSuffix(filename, ext)
	newFilename := base + ".webp"
	key := newFilename

	if u.folderPath != "" {
		key = path.Join(u.folderPath, newFilename)
	}

	// Debug logging
	fmt.Printf("[DEBUG] S3 Upload: bucket=%s, key=%s, region=%s, folderPath=%s\n", u.bucket, key, u.region, u.folderPath)

	// 5. Upload to S3 using direct client (like s3_backup.go does)
	_, err = u.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(u.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(buf.Bytes()),
		ContentType: aws.String("image/webp"),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to s3: %w", err)
	}

	return u.GetURL(key), nil
}

func (u *S3Uploader) GetURL(key string) string {
	if u.cdnURL != "" {
		return fmt.Sprintf("%s/%s", u.cdnURL, key)
	}
	// Fallback to standard S3 URL
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", u.bucket, u.region, key)
}
