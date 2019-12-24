package api

import (
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
		Name: header.Filename,
		Size: header.Size,
		Path: "./file/" + header.Filename,
		IsPublic: 0,
	}
	userFileMate := models.UserFileMap{
		UserInfo: userMate,
		FileInfo: fileMate,
		FileName: header.Filename,
	}
	if remark != "" {
		fileMate.Remark = sql.NullString{String: remark, Valid: true}
		userFileMate.Remark = sql.NullString{String: remark, Valid: true}
	}

	hash, err := utils.FileSha1(header)
	errCheck(g, err, "Upload:Failed to create sha1", http.StatusInternalServerError)
	fileMate.Hash = hash

	// 判断用户是否已经关联该文件
	_, err = userFileManager.SelectByUserFile(userMate.User, hash)
	if err != sql.ErrNoRows {
		g.JSON(http.StatusOK, gin.H{
			"errmsg": "Upload:The file already exist",
			"data":   nil,
		})
		return
	}

	// 判断文件是否已存在
	_, err = fileManager.SelectByHash(hash)
	if err == sql.ErrNoRows {
		err = g.SaveUploadedFile(header, fileMate.Path)
		errCheck(g, err, "Upload:Failed to save file", http.StatusInternalServerError)

		fileMate.Id, err = fileManager.Insert(fileMate)
		errCheck(g, err, "Upload:Failed to insert file information for mysql", http.StatusInternalServerError)
		err = fileManager.SetCache(fileMate.Hash, fileMate)
		errCheck(g, err, "Upload:Failed to set file information for redis", http.StatusInternalServerError)
	}

	_, err = userFileManager.Insert(userFileMate)
	errCheck(g, err, "Upload:Failed to insert user_file_map for mysql", http.StatusInternalServerError)
	err = userFileManager.SetCache(userMate.User + hash, userFileMate, )
	errCheck(g, err, "Upload:Failed to set user_file_map for redis", http.StatusInternalServerError)

	g.JSON(http.StatusOK, gin.H{
		"errmsg": "ok",
		"data":   fileMate.Hash,
	})
}

// 下载文件
func Download(g *gin.Context) {
	fileHash := g.Query("fileHash")
	userInter, _ := g.Get("userInfo")
	userMate := userInter.(models.UserInfo)

	userFileMate, err := userFileManager.GetCache(userMate.User + fileHash)
	errCheck(g, err, "Download:Failed to read user_file_map from redis", 0)
	haveCache := true
	if userFileMate.FileName == "" {
		haveCache = false
		userFileMate, err = userFileManager.SelectByUserFile(userMate.User, fileHash)
		if err == sql.ErrNoRows {
			g.JSON(http.StatusNotFound, gin.H{
				"errmsg": "Download:No files were found",
				"data":   nil,
			})
			return
		}
		errCheck(g, err, "Download:Failed to read user_file_map from mysql", http.StatusInternalServerError)
	}

	_, err = os.Stat(userFileMate.FileInfo.Path)
	errCheck(g, err, "Download:Failed to find file", http.StatusNotFound)

	if haveCache == false {
		err = userFileManager.SetCache(userMate.User+fileHash, userFileMate)
		errCheck(g, err, "Download:Failed to set redis", http.StatusInternalServerError)
	}

	g.Writer.Header().Add(
		"Content-Disposition", fmt.Sprintf(
			"attachment; filename=%s", userFileMate.FileInfo.Path,
		),
	)
	g.Writer.Header().Add("Content-Type", "application/octet-stream")
	g.File(userFileMate.FileInfo.Path)
}

// 下载公共文件
func PublicDownload(g *gin.Context) {
	fileHash := g.Query("fileHash")

	fileMate, err := fileManager.GetCache(fileHash)
	errCheck(g, err, "PublicDownload:Failed to read redis", 0)
	haveCache := true
	if fileMate.Path == "" {
		haveCache = false
		fileMate, err = fileManager.SelectByHash(fileHash)
		if err == sql.ErrNoRows {
			g.JSON(http.StatusNotFound, gin.H{
				"errmsg": "PublicDownload:No files were found",
				"data":   nil,
			})
			return
		}
		errCheck(g, err, "PublicDownload:Failed to read mysql", http.StatusInternalServerError)
	}

	if fileMate.IsPublic != 1 {
		g.JSON(http.StatusUnauthorized, gin.H{
			"errmsg": "PublicDownload:No files were found",
			"data":   nil,
		})
		return
	}

	_, err = os.Stat(fileMate.Path)
	errCheck(g, err, "PublicDownload:Failed to find file", http.StatusNotFound)

	if haveCache == false {
		err = fileManager.SetCache(fileMate.Hash, fileMate)
		errCheck(g, err, "PublicDownload:Failed to set redis", http.StatusInternalServerError)
	}

	g.Writer.Header().Add(
		"Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileMate.Path))
	g.Writer.Header().Add("Content-Type", "application/octet-stream")
	g.File(fileMate.Path)
}

func UpdateFileName(g *gin.Context) {
	fileHash := g.PostForm("fileHash")
	newFileName := g.PostForm("newFileName")
	userInter, _ := g.Get("userInfo")
	userMate := userInter.(models.UserInfo)

	userFileMate, err := userFileManager.GetCache(userMate.User + fileHash)
	errCheck(g, err, "UpdateFileName:Failed to read user_file_map from redis", 0)
	if userFileMate.FileName == "" {
		userFileMate, err = userFileManager.SelectByUserFile(userMate.User, fileHash)
		if err == sql.ErrNoRows {
			g.JSON(http.StatusNotFound, gin.H{
				"errmsg": "UpdateFileName:No files were found",
				"data":   nil,
			})
			return
		}
		errCheck(g, err, "UpdateFileName:Failed to read user_file_map from mysql", http.StatusInternalServerError)
	}

	userFileMate.FileName = newFileName

	err = userFileManager.Update(userFileMate)
	errCheck(g, err, "UpdateFileName:Failed to save mysql", http.StatusInternalServerError)
	err = userFileManager.SetCache(userMate.User+fileHash, userFileMate)
	errCheck(g, err, "UpdateFileName:Failed to save redis", http.StatusInternalServerError)

	g.JSON(http.StatusOK, gin.H{
		"errmsg": "ok",
		"data":   nil,
	})
}

func Delete(g *gin.Context) {
	fileHash := g.PostForm("fileHash")
	userInter, _ := g.Get("userInfo")
	userMate := userInter.(models.UserInfo)

	userFileMate, err := userFileManager.GetCache(userMate.User + fileHash)
	errCheck(g, err, "Delete:Failed to read user_file_map from redis", 0)
	if userFileMate.FileName == "" {
		userFileMate, err = userFileManager.SelectByUserFile(userMate.User, fileHash)
		if err == sql.ErrNoRows {
			g.JSON(http.StatusNotFound, gin.H{
				"errmsg": "Delete:No files were found",
				"data":   nil,
			})
			return
		}
		errCheck(g, err, "Delete:Failed to read user_file_map from mysql", http.StatusInternalServerError)
	}

	err = userFileManager.Delete(userFileMate)
	errCheck(g, err, "Delete:Failed to remove user_file_map from mysql", http.StatusInternalServerError)
	err = userFileManager.DelCache(userMate.User + fileHash)
	errCheck(g, err, "Delete:Failed to remove user_file_map from redis", http.StatusInternalServerError)

	_, err = userFileManager.SelectByFile(fileHash)
	if err == sql.ErrNoRows {
		err = fileManager.Delete(fileHash)
		errCheck(g, err, "Delete:Failed to remove file_info from mysql", 0)
		err = fileManager.DelCache(fileHash)
		errCheck(g, err, "Delete:Failed to remove file_info from redis", 0)
		err = os.Remove(userFileMate.FileInfo.Path)
		errCheck(g, err, "Delete:Failed to remove file", 0)
	}

	g.JSON(http.StatusOK, gin.H{
		"errmsg": "ok",
		"data":   nil,
	})
}
