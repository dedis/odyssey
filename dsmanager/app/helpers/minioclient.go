package helpers

import (
	"os"

	"github.com/minio/minio-go/v6"
	"golang.org/x/xerrors"
)

// minioClient is the client that can talk to our cloud endpoint
var minioClient *minio.Client

// GetMinioClient return a static minio client
func GetMinioClient() (*minio.Client, error) {
	var err error
	if minioClient == nil {
		if os.Getenv("MINIO_ENDPOINT") == "" {
			return nil, xerrors.New("MINIO_ENDPOINT env variable is empty")
		}
		if os.Getenv("MINIO_KEY") == "" {
			return nil, xerrors.New("MINIO_KEY env variable is empty")
		}
		if os.Getenv("MINIO_SECRET") == "" {
			return nil, xerrors.New("MINIO_SECRET env variable is empty")
		}
		minioClient, err = minio.New(os.Getenv("MINIO_ENDPOINT"), os.Getenv("MINIO_KEY"), os.Getenv("MINIO_SECRET"), false)
		if err != nil {
			return nil, xerrors.Errorf("failed to create minio client: %v", err)
		}
	}
	return minioClient, nil
}
