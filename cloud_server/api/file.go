package api

import (
	"MyCloud/cloud_server/meta"
	"MyCloud/utils"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"os"
)

// 上传文件
func Upload(g *gin.Context) {
	file, header, err := g.Request.FormFile("file")
	errCheck(g, err, "Upload:Failed to get request", http.StatusBadRequest)
	defer file.Close()

	fileMate := meta.NewFileMeta(header.Filename, "/file/"+header.Filename)

	newFile, err := os.Create(fileMate.Location)
	errCheck(g, err, "Upload:Failed to create file", http.StatusInternalServerError)
	defer newFile.Close()

	fileMate.FileSize, err = io.Copy(newFile, file)
	errCheck(g, err, "Upload:Failed to copy file", http.StatusInternalServerError)

	// 移动句柄的位置
	_, err = newFile.Seek(0,0)
	errCheck(g, err, "Upload:Failed to seek file", http.StatusInternalServerError)

	// 生成哈希
	fileMate.FileSha1, err = utils.FileSha1(newFile)
	errCheck(g, err, "Upload:Failed to create sha1", http.StatusInternalServerError)
	meta.UpdateFileMeta(fileMate)

	g.JSON(http.StatusOK, gin.H{
		"errmsg": "ok",
		"data":   nil,
	})
}
