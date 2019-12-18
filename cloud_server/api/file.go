package api

import (
	"MyCloud/cloud_server/meta"
	"MyCloud/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
)

// 上传文件
func Upload(g *gin.Context) {
	header, err := g.FormFile("file")
	errCheck(g, err, "Upload:Failed to get request", http.StatusBadRequest)

	fileMate := meta.NewFileMeta(header.Filename, "./file/"+header.Filename)

	err = g.SaveUploadedFile(header, fileMate.Location)
	errCheck(g, err, "Upload:Failed to save file", http.StatusInternalServerError)

	// 生成哈希
	fileMate.FileSha1, err = utils.FileSha1(fileMate.Location)
	errCheck(g, err, "Upload:Failed to create sha1", http.StatusInternalServerError)
	meta.UpdateFileMeta(fileMate)

	g.JSON(http.StatusOK, gin.H{
		"errmsg": "ok",
		"data":   fileMate.FileSha1,
	})
}

// 下载文件
func Download(g *gin.Context) {
	fileHash := g.Query("fileHash")

	fileMate := meta.GetFileMeta(fileHash)

	// 判断空

	g.Writer.Header().Add(
		"Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileMate.Location))
	g.Writer.Header().Add("Content-Type", "application/octet-stream")
	g.File(fileMate.Location)
}

func Update(g *gin.Context) {
	fileHash := g.PostForm("fileHash")
	newFileName := g.PostForm("newFileName")

	// 权限判断

	// 获取文件信息
	fileMate := meta.GetFileMeta(fileHash)

	// 修改文件信息
	fileMate.FileName = newFileName

	// 保存文件信息
	meta.UpdateFileMeta(fileMate)

	g.JSON(http.StatusOK, gin.H{
		"errmsg": "ok",
		"data":   nil,
	})
}

func Delete(g *gin.Context) {
	fileHash := g.PostForm("fileHash")

	fileMate := meta.GetFileMeta(fileHash)

	meta.RemoveFileMeta(fileHash)

	err := os.Remove(fileMate.Location)
	errCheck(g, err, "Delete:Failed to remove file", 0)

	g.JSON(http.StatusOK, gin.H{
		"errmsg": "ok",
		"data":   nil,
	})
}
