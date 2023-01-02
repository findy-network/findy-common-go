package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

type S3Client struct {
	ctx context.Context
	*s3.Client
	*manager.Downloader
}

func NewS3Client() *S3Client {
	ctx := context.TODO()
	cfg := try.To1(config.LoadDefaultConfig(ctx))

	svc := s3.NewFromConfig(cfg)
	return &S3Client{
		ctx:        ctx,
		Client:     svc,
		Downloader: manager.NewDownloader(svc),
	}
}

func (c *S3Client) S3ListBucketFiles(bucketName string) (*s3.ListObjectsV2Output, error) {
	return c.ListObjectsV2(
		c.ctx,
		&s3.ListObjectsV2Input{Bucket: aws.String(bucketName)},
	)
}

func (c *S3Client) S3DownloadBucketFiles(
	bucketName, subfolder, targetFolder string,
	input *s3.ListObjectsV2Output,
) (
	err error,
) {
	defer err2.Handle(&err)

	for _, item := range input.Contents {
		if subfolder == "" || strings.HasPrefix(*item.Key, subfolder) {
			fmt.Println("Name:          ", *item.Key)

			// Create a file to write the S3 Object contents to.
			filename := targetFolder + "/" + *item.Key
			if _, err := os.Stat(filename); err == nil {
				return fmt.Errorf("File (%s) already exists, cleanup target first!", filename)
			}

			try.To(os.MkdirAll(filepath.Dir(filename), os.ModePerm))
			f := try.To1(os.Create(filename))

			// Write the contents of S3 Object to the file
			n := try.To1(c.Download(c.ctx, f, &s3.GetObjectInput{
				Bucket: aws.String(bucketName),
				Key:    aws.String(*item.Key),
			}))
			fmt.Printf("file downloaded, %d bytes\n", n)

			f.Close()
		}
	}
	return nil
}
