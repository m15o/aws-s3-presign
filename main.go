package main

import (
	"os"
	"fmt"
	"log"
	"flag"
	"time"
	"strings"
	"net/url"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"
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

func (s *S3ObjectSigner) Presign() (*url.URL, error) {
	awsSession, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return nil, err
	}

	s3c := s3.New(awsSession)

	op := &request.Operation{
		Name:       "Presign",
		HTTPMethod: s.HTTPMethod,
		HTTPPath:   "/{Bucket}/{Key+}",
	}

	in := &s3.PutObjectInput{Bucket: aws.String(s.Bucket), Key: aws.String(s.Key)}
	var out interface{}
	r := s3c.NewRequest(op, in, out)
	r.Presign(s.Expire)

	return r.HTTPRequest.URL, nil
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
		return &S3Object{Bucket: u.Hostname(), Key: u.Path}, nil
	} else {

		if strings.HasPrefix(u.Hostname(), "s3-") {
			parts := strings.SplitN(u.Path, "/", 2)
			if parts == nil || len(parts) != 2 {
				return nil, errors.WithStack(fmt.Errorf("invalid url: %v", u))
			}
			return &S3Object{Bucket: parts[0], Key: parts[1]}, nil
		} else {
			return nil, errors.WithStack(fmt.Errorf("invalid url: %v", u))
		}
	}
}

func parse() (*S3ObjectSigner, error) {
	var httpMethod string
	var expire time.Duration

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: presign [options] (bucket key|object_url)")
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

	url, err := signer.Presign()
	if err != nil {
		log.Fatalf("%+v\n", err)
	}
	fmt.Printf("%s\n", url.String())
}
