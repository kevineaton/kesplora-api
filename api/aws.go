package api

import (
	"bytes"
	"context"
	"errors"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func checkAWSConfig() error {
	if config.AWSS3Client == nil {
		return errors.New("aws s3 not configured")
	}
	return nil
}

// ListFilesInBucket will list the files in the bucket. Note that this isn't the
// ideal way to list lal files for the account, as the files should be handled
// in the db
func ListFilesInBucket() ([]types.Object, error) {
	out := []types.Object{}
	err := checkAWSConfig()
	if err != nil {
		return out, err
	}
	output, err := config.AWSS3Client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(config.AWSS3Bucket),
	})
	if err != nil {
		return out, err
	}
	// note this doesn't handle IsTruncated yet
	return output.Contents, nil
}

// UploadFileToBucket uploads a file to a bucket
func UploadFileToBucket(key string, data []byte) error {
	err := checkAWSConfig()
	if err != nil {
		return err
	}

	reader := bytes.NewReader(data)
	_, err = config.AWSS3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(config.AWSS3Bucket),
		Key:         aws.String(key),
		Body:        reader,
		ContentType: aws.String("application/octet-stream"),
	})
	return err
}

// DeleteFileFromBucket deletes a file from a bucket
func DeleteFileFromBucket(key string) error {
	err := checkAWSConfig()
	if err != nil {
		return err
	}
	_, err = config.AWSS3Client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(config.AWSS3Bucket),
		Key:    aws.String(key),
	})
	return err
}

// GetFileFromBucket gets the raw bytes of an object from a bucket
func GetFileFromBucket(key string) ([]byte, error) {
	data := []byte{}
	err := checkAWSConfig()
	if err != nil {
		return data, err
	}
	result, err := config.AWSS3Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(config.AWSS3Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return data, err
	}
	defer result.Body.Close()
	return io.ReadAll(result.Body)
}
