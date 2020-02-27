package helpers

import (
	"net/url"
	"os"
	"strings"

	"github.com/minio/minio-go/v6"
	"golang.org/x/xerrors"
)

// minioClient is the client that can talk to our cloud endpoint
var minioClient *minio.Client

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

		// Endpoint must be host:port, if they give us a url, parse those out for them.
		ep := os.Getenv("MINIO_ENDPOINT")
		if strings.Contains(ep, "://") {
			u, err := url.Parse(ep)
			if err == nil {
				ep = u.Host
			}
		}

		var err error
		minioClient, err = minio.New(ep, os.Getenv("MINIO_ACCESS_KEY"), os.Getenv("MINIO_SECRET_KEY"), false)
		if err != nil {
			return nil, xerrors.Errorf("failed to create minio client: %v", err)
		}
	}
	return minioClient, nil
}
