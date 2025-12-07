package storage

import "os"

type S3Config struct {
	Bucket string
	Region string
	CDNURL string
}

func LoadS3Config() *S3Config {
	return &S3Config{
		Bucket: os.Getenv("S3_BUCKET"),
		Region: os.Getenv("S3_REGION"),
		CDNURL: os.Getenv("CDN_URL"),
	}
}
