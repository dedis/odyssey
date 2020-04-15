package helpers

import (
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/minio/minio-go/v6"
	"golang.org/x/xerrors"
)

// minioClient is the client that can talk to our cloud endpoint
var minioClient *minio.Client

// CloudClient defines the primitives needed to use a cloud client, This
// abstraction allows us to mock the traditional minio client.
type CloudClient interface {
	PutObject(bucketName, objectName string, reader io.Reader, objectSize int64,
		opts interface{}) (n int64, err error)
	GetObject(bucketName, objectName string, opts interface{}) (CloudObject, error)
	BucketExists(bucketName string) (bool, error)
	MakeBucket(bucketName string, location string) (err error)
}

// MinioCloudClient is the default implementation for a CloudClient
type MinioCloudClient struct {
	client *minio.Client
}

// NewMinioCloudClient return a new minio cloud client
func NewMinioCloudClient() (CloudClient, error) {
	client, err := GetMinioClient()
	if err != nil {
		return nil, xerrors.Errorf("failed to get minio client: %v", err)
	}

	return MinioCloudClient{client: client}, nil
}

// PutObject puts an object on the cloud
func (mcc MinioCloudClient) PutObject(bucketName, objectName string, reader io.Reader, objectSize int64,
	opts interface{}) (n int64, err error) {

	minioOpts, ok := opts.(minio.PutObjectOptions)
	if !ok {
		return 0, xerrors.Errorf("unkown opts: %v", opts)
	}

	return mcc.client.PutObject(bucketName, objectName, reader, objectSize, minioOpts)
}

// GetObject gets an object
func (mcc MinioCloudClient) GetObject(bucketName, objectName string,
	opts interface{}) (CloudObject, error) {

	minioOpts, ok := opts.(minio.GetObjectOptions)
	if !ok {
		return nil, xerrors.Errorf("unkown opts: %v", opts)
	}

	return mcc.client.GetObject(bucketName, objectName, minioOpts)
}

// BucketExists tells if a bucket exists
func (mcc MinioCloudClient) BucketExists(bucketName string) (bool, error) {
	return mcc.client.BucketExists(bucketName)
}

// MakeBucket makes a bucket
func (mcc MinioCloudClient) MakeBucket(bucketName string, location string) (err error) {
	return mcc.client.MakeBucket(bucketName, location)
}

// GetMinioClient return a static minio client
func GetMinioClient() (*minio.Client, error) {
	if minioClient == nil {
		if os.Getenv("MINIO_ENDPOINT") == "" {
			return nil, xerrors.New("MINIO_ENDPOINT env variable is empty")
		}
		if os.Getenv("MINIO_ACCESS_KEY") == "" {
			return nil, xerrors.New("MINIO_ACCESS_KEY env variable is empty")
		}
		if os.Getenv("MINIO_SECRET_KEY") == "" {
			return nil, xerrors.New("MINIO_SECRET_KEY env variable is empty")
		}

		// Endpoint must be host:port, if they give us a url, parse those out
		// for them.
		ep := os.Getenv("MINIO_ENDPOINT")
		if strings.Contains(ep, "://") {
			u, err := url.Parse(ep)
			if err != nil {
				return nil, xerrors.Errorf("failed to parse minion endpoint url: %v", err)
			}
			ep = u.Host
		}

		var err error
		minioClient, err = minio.New(ep, os.Getenv("MINIO_ACCESS_KEY"), os.Getenv("MINIO_SECRET_KEY"), false)
		if err != nil {
			return nil, xerrors.Errorf("failed to create minio client: %v", err)
		}
	}
	return minioClient, nil
}

// CloudObject is the abstraction of an element that we get from the cloud
type CloudObject interface {
	Close() (err error)
	Stat() (minio.ObjectInfo, error)
	Read(p []byte) (n int, err error)
}
