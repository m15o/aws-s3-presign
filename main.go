package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"golang.org/x/xerrors"
)

type S3ObjectSigner struct {
	*S3Object
	HTTPMethod string
	Expire     time.Duration
}

type S3Object struct {
	Bucket string
	Key    string
}

func (s *S3ObjectSigner) Presign(ctx context.Context) (*string, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	s3Client := s3.NewFromConfig(cfg)
	presignClient := s3.NewPresignClient(s3Client)
	var presignRequest *v4.PresignedHTTPRequest
	switch s.HTTPMethod {
	case "HEAD":
		input := s3.HeadObjectInput{Bucket: aws.String(s.Bucket), Key: aws.String(s.Key)}
		r, err := presignClient.PresignHeadObject(ctx, &input)
		if err != nil {
			return nil, err
		}
		presignRequest = r

	case "GET":
		input := s3.GetObjectInput{Bucket: aws.String(s.Bucket), Key: aws.String(s.Key)}
		r, err := presignClient.PresignGetObject(ctx, &input)
		if err != nil {
			return nil, err
		}
		presignRequest = r

	case "PUT":
		input := s3.PutObjectInput{Bucket: aws.String(s.Bucket), Key: aws.String(s.Key)}
		r, err := presignClient.PresignPutObject(ctx, &input)
		if err != nil {
			return nil, err
		}
		presignRequest = r

	case "DELETE":
		input := s3.DeleteObjectInput{Bucket: aws.String(s.Bucket), Key: aws.String(s.Key)}
		r, err := presignClient.PresignDeleteObject(ctx, &input)
		if err != nil {
			return nil, err
		}
		presignRequest = r

	default:
		return nil, fmt.Errorf("unssuported method: %s", s.HTTPMethod)
	}

	return &presignRequest.URL, nil
}

func (s *S3Object) NewSigner(httpMethod string, expire time.Duration) *S3ObjectSigner {
	return &S3ObjectSigner{s, httpMethod, expire}
}

func parseS3Url(s3url string) (*S3Object, error) {
	u, err := url.Parse(s3url)
	if err != nil {
		return nil, err
	}

	if u.Scheme == "s3" {
		log.Printf("bucket: %s", u.Hostname())
		log.Printf("path: %s", u.Path)
		key := strings.TrimLeft(u.Path, "/")
		return &S3Object{Bucket: u.Hostname(), Key: key}, nil
	} else {
		if strings.HasPrefix(u.Hostname(), "s3-") {
			parts := strings.SplitN(u.Path, "/", 2)
			if parts == nil || len(parts) != 2 {
				return nil, xerrors.Errorf("invalid url: %v", u)
			}
			return &S3Object{Bucket: parts[0], Key: parts[1]}, nil
		} else {
			return nil, xerrors.Errorf("invalid url: %v", u)
		}
	}
}

func parse() (*S3ObjectSigner, error) {
	var httpMethod string
	var expire time.Duration

	flag.Usage = func() {
		_, _ = fmt.Fprintln(os.Stderr, "Usage: presign [options] (bucket key|object_url)")
		flag.PrintDefaults()
	}

	flag.StringVar(&httpMethod, "method", "GET", "HTTP Method")
	flag.DurationVar(&expire, "expire", 5*time.Minute, "Expire duration")
	flag.Parse()

	if flag.NArg() == 0 || flag.NArg() > 2 {
		flag.Usage()
		os.Exit(1)
	}

	if flag.NArg() == 1 {
		s3object, err := parseS3Url(flag.Arg(0))
		if err != nil {
			return nil, err
		}
		return s3object.NewSigner(httpMethod, expire), nil
	} else {
		bucket := flag.Arg(0)
		key := flag.Arg(1)
		s3object := &S3Object{bucket, key}
		return s3object.NewSigner(httpMethod, expire), nil
	}
}

func main() {
	signer, err := parse()
	if err != nil {
		log.Fatalf("%+v\n", err)
	}

	signedUrl, err := signer.Presign(context.Background())
	if err != nil {
		log.Fatalf("%+v\n", err)
	}
	fmt.Printf("%s\n", *signedUrl)
}
