package storage

import "os"

type S3Config struct {
	Bucket       string
	Region       string
	CDNURL       string
	Folder       string // For avatars/uploads
	BackupFolder string // For backups
}

func LoadS3Config() *S3Config {
	return &S3Config{
		Bucket:       os.Getenv("S3_BUCKET"),
		Region:       os.Getenv("S3_REGION"),
		CDNURL:       os.Getenv("CDN_URL"),
		Folder:       os.Getenv("S3_FOLDER"),
		BackupFolder: os.Getenv("S3_BACKUP_FOLDER"),
	}
}
