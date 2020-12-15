package internal

import (
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// DefaultRegion to use for S3 credential creation
const defaultRegion = "us-east-1"

func buildSessionConfig() aws.Config {
	// By default make sure a region is specified, this is required for S3 operations
	sessionConfig := aws.Config{Region: aws.String(defaultRegion)}

	// if config.AccessKey != "" && config.SecretKey != "" {
	// 	sessionConfig.Credentials = credentials.NewStaticCredentials(config.AccessKey, config.SecretKey, "")
	// }

	return sessionConfig
}

func buildEndpointResolver(hostname string) endpoints.Resolver {
	defaultResolver := endpoints.DefaultResolver()

	fixedHost := hostname
	if !strings.HasPrefix(hostname, "http") {
		fixedHost = "https://" + hostname
	}

	return endpoints.ResolverFunc(func(service, region string, optFns ...func(*endpoints.Options)) (endpoints.ResolvedEndpoint, error) {
		if service == endpoints.S3ServiceID {
			return endpoints.ResolvedEndpoint{
				URL: fixedHost,
			}, nil
		}

		return defaultResolver.EndpointFor(service, region, optFns...)
	})
}

// SessionNew - Read the config for default credentials, if not provided use environment based variables
func SessionNew() *s3.S3 {
	sessionConfig := buildSessionConfig()
	hostBase := os.Getenv("S3_ENDPOINT_BASE")

	if hostBase != "" && hostBase != "s3.amazon.com" {
		sessionConfig.EndpointResolver = buildEndpointResolver(hostBase)
	}

	return s3.New(session.Must(session.NewSessionWithOptions(session.Options{
		Config:            sessionConfig,
		SharedConfigState: session.SharedConfigEnable,
	})))
}

// SessionForBucket - For a given S3 bucket, create an approprate session that references the region
// that this bucket is located in
func SessionForBucket(bucket string) (*s3.S3, error) {
	sessionConfig := buildSessionConfig()

	hostBucket := os.Getenv("S3_ENDPOINT")

	if hostBucket == "" || hostBucket == "%(bucket)s.s3.amazonaws.com" {
		svc := SessionNew()

		if loc, err := svc.GetBucketLocation(&s3.GetBucketLocationInput{Bucket: &bucket}); err != nil {
			return nil, err
		} else if loc.LocationConstraint == nil {
			// Use default service
			return svc, nil
		} else {
			sessionConfig.Region = loc.LocationConstraint
		}
	} else {
		host := strings.ReplaceAll(hostBucket, "%(bucket)s", bucket)

		sessionConfig.EndpointResolver = buildEndpointResolver(host)
	}

	return s3.New(session.Must(session.NewSessionWithOptions(session.Options{
		Config:            sessionConfig,
		SharedConfigState: session.SharedConfigEnable,
	}))), nil
}
