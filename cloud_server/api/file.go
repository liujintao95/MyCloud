package api

import (
	"MyCloud/cloud_server/meta"
	"MyCloud/cloud_server/models"
	"MyCloud/utils"
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
)

// 上传文件
func Upload(g *gin.Context) {
	remark := g.PostForm("remark")
	header, err := g.FormFile("file")
	errCheck(g, err, "Upload:Failed to get request", http.StatusBadRequest)
	userInter, _ := g.Get("userInfo")
	userMate := userInter.(models.UserInfo)

	fileMate := models.FileInfo{
		Name:     header.Filename,
		Size:     header.Size,
		Path:     "./file/" + header.Filename,
	}
	userFileMate := models.UserFileMap{
		UserId:   userMate.Id,
		FileName: header.Filename,
	}
	if remark != ""{
		fileMate.Remark = sql.NullString{String: remark, Valid: true}
		userFileMate.Remark = sql.NullString{String: remark, Valid: true}
	}

	err = g.SaveUploadedFile(header, fileMate.Path)
	errCheck(g, err, "Upload:Failed to save file", http.StatusInternalServerError)

	// 生成哈希
	fileMate.Hash, err = utils.FileSha1(fileMate.Path)
	errCheck(g, err, "Upload:Failed to create sha1", http.StatusInternalServerError)

	_, err = fileManager.Insert(fileMate)
	errCheck(g, err, "Upload:Failed to insert mysql", http.StatusInternalServerError)

	err = fileManager.SetCache(fileMate.Hash, fileMate)
	errCheck(g, err, "Upload:Failed to set redis", http.StatusInternalServerError)

	g.JSON(http.StatusOK, gin.H{
		"errmsg": "ok",
		"data":   fileMate.Hash,
	})
}

// 下载文件
func Download(g *gin.Context) {
	fileHash := g.Query("fileHash")

	// 权限验证

	fileMate, err := fileManager.GetCache(fileHash)
	errCheck(g, err, "Download:Failed to read redis", 0)
	haveCache := true
	if fileMate.Path == "" {
		// 如果没有则取mysql中的数据
		haveCache = false
		fileMate, err = fileManager.SelectByHash(fileHash)
		if err == sql.ErrNoRows {
			g.JSON(http.StatusNotFound, gin.H{
				"errmsg": "Download:Failed to read mysql",
				"data":   nil,
			})
			return
		}
		errCheck(g, err, "Download:Failed to read mysql", http.StatusInternalServerError)
	}

	// 判断是否有文件
	_, err = os.Stat(fileMate.Path)
	errCheck(g, err, "Download:Failed to find file", http.StatusNotFound)

	if haveCache == false {
		err = fileManager.SetCache(fileMate.Hash, fileMate)
		errCheck(g, err, "Download:Failed to set redis", http.StatusInternalServerError)
	}

	g.Writer.Header().Add(
		"Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileMate.Path))
	g.Writer.Header().Add("Content-Type", "application/octet-stream")
	g.File(fileMate.Path)
}

func UpdateFileName(g *gin.Context) {
	fileHash := g.PostForm("fileHash")
	newFileName := g.PostForm("newFileName")

	// 权限验证

	// 获取文件信息
	fileMate, err := fileManager.SelectByHash(fileHash)
	errCheck(g, err, "UpdateFileName:Failed to read mysql", http.StatusInternalServerError)

	// 修改文件信息
	fileMate.Name = newFileName

	// 保存文件信息
	err = fileManager.Update(fileMate)
	errCheck(g, err, "UpdateFileName:Failed to save mysql", http.StatusInternalServerError)
	err = fileManager.SetCache(fileMate.Hash, fileMate)
	errCheck(g, err, "UpdateFileName:Failed to save redis", http.StatusInternalServerError)

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
