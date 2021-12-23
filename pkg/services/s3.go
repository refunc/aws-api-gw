package services

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/minio/minio-go"
	"github.com/refunc/refunc/pkg/env"
)

func SetFunctionCode(code map[string]string, ns string, name string) (string, int64, string, error) {
	bucket, bucket_ok := code["S3Bucket"]
	key, key_ok := code["S3Key"]
	if bucket_ok && key_ok {
		return setFunctionS3BucketCode(bucket, key)
	}
	blob, ok := code["ZipFile"]
	if ok {
		S3KeyPrefix := fmt.Sprintf("%s/funcs/%s/%s", env.GlobalScopeRoot, ns, name)
		return setFunctionBlobCode(S3KeyPrefix, blob)
	}
	return "", 0, "", errors.New("function code type error")
}

func setFunctionS3BucketCode(bucket string, key string) (string, int64, string, error) {
	mc := env.GlobalMinioClient()
	stat, err := mc.StatObject(bucket, key, minio.StatObjectOptions{})
	if err != nil {
		return "", 0, "", err
	}
	return fmt.Sprintf("s3://%s/%s", bucket, key), stat.Size, stat.ETag, nil
}

func setFunctionBlobCode(S3KeyPrefix, blob string) (string, int64, string, error) {
	mc := env.GlobalMinioClient()
	//TODO should big size blob in memory?
	bts, err := base64.StdEncoding.DecodeString(blob)
	if err != nil {
		return "", 0, "", err
	}
	sha256sum := sha256.Sum256(bts)
	etag := hex.EncodeToString(sha256sum[:])
	size := int64(len(bts))
	key := fmt.Sprintf("%s/%s.zip", S3KeyPrefix, etag)
	putSize, err := mc.PutObject(env.GlobalBucket, key, bytes.NewReader(bts), size, minio.PutObjectOptions{})
	if err != nil {
		return "", 0, "", err
	}
	if putSize != size {
		return "", 0, "", errors.New("put code blob error")
	}
	return fmt.Sprintf("s3://%s%s", env.GlobalBucket, key), size, etag, nil
}
