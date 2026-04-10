package file

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/config"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/log"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
)

// ServiceMinio 文件上传
type ServiceMinio struct {
	log.Log
	ctx            *config.Context
	downloadClient *http.Client
}

// NewServiceMinio NewServiceMinio
func NewServiceMinio(ctx *config.Context) *ServiceMinio {
	return &ServiceMinio{
		Log: log.NewTLog("File"),
		ctx: ctx,
		downloadClient: &http.Client{
			Timeout: time.Second * 30,
		},
	}
}

// UploadFile 上传文件
func (sm *ServiceMinio) UploadFile(filePath string, contentType string, copyFileWriter func(io.Writer) error) (map[string]interface{}, error) {
	buff := bytes.NewBuffer(make([]byte, 0))
	err := copyFileWriter(buff)
	if err != nil {
		sm.Error("复制文件内容失败！", zap.Error(err))
		return nil, err
	}

	minioConfig := sm.ctx.GetConfig().Minio

	ctx := context.Background()
	uploadUl, _ := url.Parse(minioConfig.URL)
	endpoint := uploadUl.Host
	accessKeyID := minioConfig.AccessKeyID
	secretAccessKey := minioConfig.SecretAccessKey
	useSSL := false

	if strings.HasPrefix(uploadUl.Scheme, "https") {
		useSSL = true
	}
	// 初使化minio client对象。
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		sm.Error("创建错误：", zap.Error(err))
		return nil, err
	}
	// 使用固定的bucket名称 "file"
	bucketName := "file"

	// 尝试检查bucket是否存在
	exists, err := minioClient.BucketExists(ctx, bucketName)
	if err != nil {
		sm.Error(fmt.Sprintf("检测 %s目录是否存在错误", bucketName), zap.Error(err))
		// 继续执行，尝试直接上传
	} else if !exists {
		// 尝试创建bucket
		err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			sm.Error(fmt.Sprintf("创建 %s目录失败", bucketName), zap.Error(err))
			// 继续执行，尝试直接上传
		} else {
			// 设置bucket策略
			policy := `{
				"Version": "2012-10-17",
				"Statement": [{
					"Effect": "Allow",
					"Principal": {
						"AWS": ["*"]
					},
					"Action": ["s3:GetBucketLocation", "s3:ListBucket", "s3:ListBucketMultipartUploads"],
					"Resource": ["arn:aws:s3:::%s"]
				}, {
					"Effect": "Allow",
					"Principal": {
						"AWS": ["*"]
					},
					"Action": ["s3:AbortMultipartUpload", "s3:DeleteObject", "s3:GetObject", "s3:ListMultipartUploadParts", "s3:PutObject"],
					"Resource": ["arn:aws:s3:::%s/*"]
				}]
			}`
			err = minioClient.SetBucketPolicy(context.Background(), bucketName, fmt.Sprintf(policy, bucketName, bucketName))
			if err != nil {
				sm.Error("设置minio文件读写权限错误", zap.Error(err))
				// 继续执行，尝试直接上传
			}
		}
	}

	// 使用完整的filePath作为object名称
	objectName := filePath
	// 移除可能的前导斜杠
	objectName = strings.TrimPrefix(objectName, "/")

	// 尝试上传文件
	n, err := minioClient.PutObject(ctx, bucketName, objectName, buff, int64(len(buff.Bytes())), minio.PutObjectOptions{ContentType: contentType, PartSize: 10 * 1024 * 1024})
	if err != nil {
		sm.Error("上传文件失败：", zap.Error(err))
		return map[string]interface{}{
			"path": "",
		}, err
	}
	return map[string]interface{}{
		"path": n.Key,
	}, err
}

func (sm *ServiceMinio) DownloadURL(ph string, filename string) (string, error) {
	minioConfig := sm.ctx.GetConfig().Minio

	// 确保使用HTTP协议
	minioURL := minioConfig.URL
	if strings.HasPrefix(minioURL, "https") {
		minioURL = "http" + minioURL[5:]
	}

	vals := url.Values{}
	vals.Set("response-content-disposition", fmt.Sprintf("inline; filename=\"%s\"", filename))
	// 包含bucket名称 "file" 在URL路径中
	result, _ := url.JoinPath(minioURL, "file", ph)
	return fmt.Sprintf("%s?%s", result, vals.Encode()), nil
}
