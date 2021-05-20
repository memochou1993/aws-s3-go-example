package main

import (
	"log"
	"mime/multipart"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var (
	AccessKeyID     = ""
	SecretAccessKey = ""
	Region          = ""
	Bucket          = ""
)

func main() {
	sess := ConnectAws()

	http.Handle("/", http.FileServer(http.Dir("./public")))

	http.HandleFunc("/upload-file", func(w http.ResponseWriter, r *http.Request) {
		r.ParseMultipartForm(32 << 20)
		for _, fileHeaders := range r.MultipartForm.File {
			for _, header := range fileHeaders {
				Upload(sess, Bucket, header)
			}
		}
	})

	http.HandleFunc("/upload-files", func(w http.ResponseWriter, r *http.Request) {
		r.ParseMultipartForm(32 << 20)
		for _, fileHeaders := range r.MultipartForm.File {
			UploadMultiple(sess, Bucket, fileHeaders)
		}
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func ConnectAws() *session.Session {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(Region),
		Credentials: credentials.NewStaticCredentials(
			AccessKeyID,
			SecretAccessKey,
			"",
		),
	})
	if err != nil {
		panic(err)
	}
	return sess
}

func Upload(sess *session.Session, bucket string, header *multipart.FileHeader) error {
	svc := s3manager.NewUploader(sess)
	src, err := header.Open()
	if err != nil {
		return err
	}
	if _, err := svc.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		ACL:    aws.String("public-read"),
		Key:    aws.String(header.Filename),
		Body:   src,
	}); err != nil {
		return err
	}
	return nil
}

func UploadMultiple(sess *session.Session, bucket string, fileHeaders []*multipart.FileHeader) error {
	var objects []s3manager.BatchUploadObject
	for _, header := range fileHeaders {
		src, err := header.Open()
		if err != nil {
			return err
		}
		objects = append(objects, s3manager.BatchUploadObject{
			Object: &s3manager.UploadInput{
				Bucket: aws.String(bucket),
				ACL:    aws.String("public-read"),
				Key:    aws.String(header.Filename),
				Body:   src,
			},
		})
	}
	svc := s3manager.NewUploader(sess)
	iter := &s3manager.UploadObjectsIterator{Objects: objects}
	return svc.UploadWithIterator(aws.BackgroundContext(), iter)
}
