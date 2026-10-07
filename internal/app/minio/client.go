package minio

import (
    "context"
    "log"
    "os"

    "github.com/minio/minio-go/v7"
    "github.com/minio/minio-go/v7/pkg/credentials"
)

var Client *minio.Client

func Init() {
    endpoint := os.Getenv("MINIO_ENDPOINT")
    accessKey := os.Getenv("MINIO_ACCESS_KEY")
    secretKey := os.Getenv("MINIO_SECRET_KEY")

    var err error
    Client, err = minio.New(endpoint, &minio.Options{
        Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
        Secure: false,
    })
    if err != nil {
        log.Fatal(err)
    }
    log.Println("MinIO client initialized for", endpoint)
}

func UploadFile(objectName, filePath, contentType string) error {
    _, err := Client.FPutObject(
        context.Background(),
        "media",
        objectName,
        filePath,
        minio.PutObjectOptions{ContentType: contentType},
    )
    return err
}
