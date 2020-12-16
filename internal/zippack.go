package internal

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type nameQueue struct {
	bucket string
	key    string
}

type dataQueue struct {
	name string
	data []byte
}

// ZipPack given a dest file, pack all of the source files into it, recursivly
func ZipPack(dest string, source []string) error {
	destFile, err := FileURINew(dest)
	if err != nil {
		return fmt.Errorf("buckets to be prefixed with s3://")
	}

	// FileGeneration -> Downloader -> Zip Adder -> Writer
	nameQueue := generateFiles(source)
	zipQueue := downloadFiles(nameQueue)
	reader := addToZip(zipQueue)

	defer reader.Close()
	wg := sync.WaitGroup{}

	if destFile.Scheme == "s3" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := uploadToS3(*destFile, reader); err != nil {
				log.Fatal(err)
			}
		}()
	} else if destFile.Scheme == "file" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := uploadToLocal(*destFile, reader); err != nil {
				log.Fatal(err)
			}
		}()
	} else {
		fmt.Println("ERROR unknown file destination")
		os.Exit(1)
	}

	wg.Wait()

	return nil
}

func uploadToS3(destFile FileURI, reader io.Reader) error {
	// Get a region based bucket
	dstSvc, err := SessionForBucket(destFile.Bucket)
	if err != nil {
		return err
	}

	uploader := s3manager.NewUploaderWithClient(dstSvc, func(u *s3manager.Uploader) {
		u.PartSize = 10 * 1024 * 1024
		u.Concurrency = 1
	})

	params := &s3manager.UploadInput{
		Bucket: aws.String(destFile.Bucket), // Required
		Key:    aws.String(destFile.Path),
		Body:   reader,
	}

	ctx := aws.BackgroundContext()
	_, err = uploader.UploadWithContext(ctx, params)
	if err != nil {
		return err
	}

	return nil
}

func uploadToLocal(destFile FileURI, reader io.Reader) error {
	f, err := os.OpenFile(destFile.Path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	if _, err := io.Copy(f, reader); err != nil {
		return err
	}

	return nil
}

func generateFiles(source []string) chan nameQueue {
	queue := make(chan nameQueue)

	go func() {
		defer close(queue)

		svc := SessionNew()
		seen := map[string]bool{}

		for _, sourceFile := range source {
			for todo := []string{sourceFile}; len(todo) != 0; {
				var item string
				item, todo = todo[0], todo[1:]

				remotePager(svc, item, true, func(page *s3.ListObjectsV2Output) error {
					// Anything that's a directory, traverse
					for _, item := range page.CommonPrefixes {
						uri := fmt.Sprintf("s3://%s/%s", *page.Name, *item.Prefix)

						todo = append(todo, uri)
					}
					if page.Contents != nil {
						for _, item := range page.Contents {
							// If we've traversed this path, don't od it twice
							path := fmt.Sprintf("s3://%s/%s", *page.Name, *item.Key)
							if seen[path] {
								continue
							}
							seen[path] = true

							queue <- nameQueue{bucket: *page.Name, key: *item.Key}
						}
					}
					return nil
				})
			}
		}
	}()

	return queue
}

func downloadFiles(namesQueue chan nameQueue) chan dataQueue {
	queue := make(chan dataQueue)
	waiter := sync.WaitGroup{}

	go func() {
		defer close(queue)

		svc := SessionNew()
		downloader := s3manager.NewDownloaderWithClient(svc)

		for elem := range namesQueue {
			bucket := elem.bucket
			key := elem.key
			waiter.Add(1)

			go func() {
				defer waiter.Done()

				buffer := aws.NewWriteAtBuffer([]byte{})
				params := &s3.GetObjectInput{
					Bucket: &bucket,
					Key:    &key,
				}

				_, err := downloader.Download(buffer, params)
				if err != nil {
					log.Fatal(err)
				}

				queue <- dataQueue{name: key, data: buffer.Bytes()}
			}()
		}

		waiter.Wait()
	}()

	return queue
}

func addToZip(queue chan dataQueue) io.ReadCloser {
	reader, writer := io.Pipe()

	go func() {
		zipper := NewZipper(writer)

		defer func() {
			zipper.Close()
			writer.Close()
		}()

		for elem := range queue {
			if err := zipper.AddFile(elem.name, elem.data); err != nil {
				log.Fatal(err)
			}
		}
	}()

	return reader
}
