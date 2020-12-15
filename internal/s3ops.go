package internal

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

func isS3File(source string) (bool, error) {
	fileURI, err := FileURINew(source)
	if err != nil {
		return false, err
	}
	svc, err := SessionForBucket(fileURI.Bucket)

	params := &s3.HeadObjectInput{
		Bucket: aws.String(fileURI.Bucket), // Required
		Key:    fileURI.Key(),
	}
	_, err = svc.HeadObject(params)

	if err != nil {
		return false, err
	}
	return true, nil
}

func isS3Directory(source string) (bool, error) {
	if !strings.HasSuffix(source, "/") {
		source = source + "/"
	}
	fileURI, err := FileURINew(source)
	if err != nil {
		return false, err
	}

	params := &s3.ListObjectsV2Input{
		Bucket:    aws.String(fileURI.Bucket), // Required
		MaxKeys:   aws.Int64(10),
		Delimiter: aws.String("/"),
		Prefix:    fileURI.Key(),
	}
	svc, err := SessionForBucket(fileURI.Bucket)
	if err != nil {
		return false, err
	}
	objs, err := svc.ListObjectsV2(params)
	if err != nil {
		return false, err
	}

	return len(objs.CommonPrefixes) == 1 || len(objs.Contents) != 0, nil
}
