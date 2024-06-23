package s3

import (
	"context"
	"mime/multipart"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/sirupsen/logrus"
)

type Profile struct {
	s3        *s3.Client
	presigner *s3.PresignClient

	bucketName string

	log *logrus.Logger
}

func NewProfile(s3 *s3.Client, presign *s3.PresignClient, bucketName string, log *logrus.Logger) *Profile {
	return &Profile{
		s3:        s3,
		presigner: presign,

		bucketName: bucketName + "/user-avatars-develop/",

		log: log,
	}
}

func (a *Profile) UploadAvatar(ctx context.Context, file multipart.File, key string) error {
	if file == nil {
		return nil
	}

	if _, err := a.s3.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(a.bucketName),
		Key:    aws.String(key),
		Body:   file,
	}); err != nil {
		return err
	}

	return nil
}

func (p *Profile) GetAvatars(ctx context.Context, key string) (*v4.PresignedHTTPRequest, error) {
	resp, err := p.presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(p.bucketName),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(60 * int64(time.Second))
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (p *Profile) DeleteAvatar(ctx context.Context, key string) error {
	if _, err := p.s3.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(p.bucketName),
		Key:    aws.String(key),
	}); err != nil {
		return err
	}

	return nil
}
