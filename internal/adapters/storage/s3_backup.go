package storage

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Store struct {
	client *s3.Client
	bucket string
}

func NewS3Store(ctx context.Context, bucket string, region string) (*S3Store, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, err
	}
	return &S3Store{
		client: s3.NewFromConfig(cfg),
		bucket: bucket,
	}, nil
}

// Backup zips the directory at localPath and uploads it to S3 as 'driplnk_backup.zip'
func (s *S3Store) Backup(ctx context.Context, localPath string) error {
	zipPath := localPath + ".zip"
	if err := zipDirectory(localPath, zipPath); err != nil {
		return fmt.Errorf("failed to zip directory: %w", err)
	}
	defer os.Remove(zipPath) // Clean up zip after upload

	file, err := os.Open(zipPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String("driplnk_backup.zip"),
		Body:   file,
	})
	if err != nil {
		return fmt.Errorf("s3 upload failed: %w", err)
	}
	return nil
}

// Restore downloads 'driplnk_backup.zip' from S3 and unzips it to localPath
func (s *S3Store) Restore(ctx context.Context, localPath string) error {
	// Download
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String("driplnk_backup.zip"),
	})
	if err != nil {
		// If not found, it might be a fresh install
		// Check for NoSuchKey error structure or string
		if strings.Contains(err.Error(), "NoSuchKey") || strings.Contains(err.Error(), "NotFound") {
			return nil // No backup exists, start fresh
		}
		return fmt.Errorf("s3 download failed: %w", err)
	}
	defer result.Body.Close()

	zipPath := localPath + ".zip"
	outFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	// Copy body to file
	if _, err := io.Copy(outFile, result.Body); err != nil {
		outFile.Close()
		return err
	}
	outFile.Close()
	defer os.Remove(zipPath)

	// Unzip
	if err := unzipDirectory(zipPath, localPath); err != nil {
		return fmt.Errorf("failed to unzip: %w", err)
	}
	return nil
}

// Helpers for Zip/Unzip

func zipDirectory(source, target string) error {
	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name, _ = filepath.Rel(filepath.Dir(source), path)
		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})
}

func unzipDirectory(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", fpath)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
