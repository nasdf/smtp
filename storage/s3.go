package storage

import (
	"io"

  "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// S3Storage holds settings for the S3 storage provider.
type S3Storage struct {
	Bucket   string
	session  *session.Session
	uploader *s3manager.Uploader
}

// NewS3Storage creates a new config for uploading and retrieving files.
func NewS3Storage(bucket string) *S3Storage {
	sess := session.Must(session.NewSession())
	uploader := s3manager.NewUploader(sess)
	return &S3Storage{Bucket: bucket, session: sess, uploader: uploader}
}

// Upload stores the 
func (c *S3Storage) Put(name string, body io.Reader) error {
	params := &s3manager.UploadInput{Bucket: &c.Bucket, Key: &name, Body: body}
	_, err := c.uploader.Upload(params)
	return err
}