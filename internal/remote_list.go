package internal

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

type FileObject struct {
	Source   int64 // used by sync
	Name     string
	Size     int64
	Checksum string
}

func remotePager(svc *s3.S3, uri string, delim bool, pager func(page *s3.ListObjectsV2Output) error) error {
	u, err := FileURINew(uri)
	if err != nil || u.Scheme != "s3" {
		return fmt.Errorf("requires buckets to be prefixed with s3://")
	}

	params := &s3.ListObjectsV2Input{
		Bucket:  aws.String(u.Bucket), // Required
		MaxKeys: aws.Int64(1000),
	}
	if u.Path != "" && u.Path != "/" {
		params.Prefix = u.Key()
	}
	if delim {
		params.Delimiter = aws.String("/")
	}

	wrapper := func(page *s3.ListObjectsV2Output, lastPage bool) bool {
		if err := pager(page); err != nil {
			// Would be nice to return the error
			return false
		}
		return true
	}

	sess, err := SessionForBucket(u.Bucket)
	if err != nil {
		return err
	}
	if err := sess.ListObjectsV2Pages(params, wrapper); err != nil {
		return err
	}
	return nil
}

func remoteList(svc *s3.S3, args []string) ([]FileObject, error) {
	result := make([]FileObject, 0)

	for _, arg := range args {
		pager := func(page *s3.ListObjectsV2Output) error {
			for _, obj := range page.Contents {
				result = append(result, FileObject{
					Name:     *obj.Key,
					Size:     *obj.Size,
					Checksum: *obj.ETag,
				})
			}

			return nil
		}

		remotePager(svc, arg, false, pager)
	}

	return result, nil
}
