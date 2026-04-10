package file

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/config"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/log"
	"go.uber.org/zap"
)

type SeaweedFS struct {
	log.Log
	ctx *config.Context
}

func NewSeaweedFS(ctx *config.Context) *SeaweedFS {
	return &SeaweedFS{
		Log: log.NewTLog("SeaweedFS"),
		ctx: ctx,
	}
}

// UploadFile 上传文件
func (s *SeaweedFS) UploadFile(filePath string, contentType string, copyFileWriter func(io.Writer) error) (map[string]interface{}, error) {
	fileDir, fileName := filepath.Split(filePath)
	s.Debug("filePath->", zap.String("filePath", filePath), zap.String("fileDir", fileDir), zap.String("fileName", fileName))
	newFileDir := fileDir
	// 确保路径以 / 开头，但避免重复添加
	if !strings.HasPrefix(newFileDir, "/") {
		newFileDir = fmt.Sprintf("/%s", newFileDir)
	}
	seaweedConfig := s.ctx.GetConfig().Seaweed

	// 检查并创建目录（如果不存在）
	dirURL := fmt.Sprintf("%s%s", seaweedConfig.URL, newFileDir)
	dirReq, err := http.NewRequest("MKCOL", dirURL, nil)
	if err != nil {
		s.Error("创建目录请求失败！", zap.String("dirURL", dirURL), zap.Error(err))
		// 继续执行，因为目录可能已经存在
	} else {
		dirResp, err := http.DefaultClient.Do(dirReq)
		if err != nil {
			s.Error("检测目录是否存在错误", zap.String("dirURL", dirURL), zap.Error(err))
			// 继续执行，因为目录可能已经存在
		} else {
			defer dirResp.Body.Close()
		}
	}

	resultMap, err := uploadFile(fmt.Sprintf("%s%s", seaweedConfig.URL, newFileDir), fileName, copyFileWriter)
	return resultMap, err
}

func (s *SeaweedFS) DownloadURL(path string, filename string) (string, error) {
	seaweedConfig := s.ctx.GetConfig().Seaweed
	rpath, _ := url.JoinPath(seaweedConfig.URL, path)
	return rpath, nil
}
