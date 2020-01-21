package api

import (
	"MyCloud/cloud_server/models"
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"strconv"
	"time"
)

// 初始化上传文件
func InitFile(g *gin.Context) {
	fileName := g.PostForm("fileName")
	sizeStr := g.PostForm("size")
	hash := g.PostForm("hash")
	userInter, _ := g.Get("userInfo")
	userMate := userInter.(models.UserInfo)

	fileSize, _ := strconv.ParseInt(sizeStr, 10, 64)
	uploadId := userMate.User + fmt.Sprintf("%x", time.Now().UnixNano())

	fileMate := models.FileInfo{
		Name: fileName,
		Size: fileSize,
		Path: "./file/" + uploadId + fileName,
		Hash: hash,
	}
	userFileMate := models.UserFileMap{
		UserInfo: userMate,
		FileInfo: fileMate,
		FileName: fileName,
	}

	lastId, err := fileManager.Set(fileMate)
	errCheck(g, err, "InitFile:Failed to set file information", http.StatusInternalServerError)
	userFileMate.FileInfo.Id = lastId

	_, err = userFileManager.Set(userFileMate)
	errCheck(g, err, "InitFile:Failed to set user_file_map", http.StatusInternalServerError)

	g.JSON(http.StatusOK, gin.H{
		"errmsg": "ok",
		"data":   nil,
	})
}

// 上传文件
func Upload(g *gin.Context) {
	hash := g.PostForm("hash")
	header, err := g.FormFile("file")
	errCheck(g, err, "Upload:Failed to get request", http.StatusBadRequest)
	userInter, _ := g.Get("userInfo")
	userMate := userInter.(models.UserInfo)

	fileMate, err := fileManager.GetByHash(hash)
	errCheck(g, err, "Upload:Failed to get file info", http.StatusInternalServerError)

	userFileMate, err := userFileManager.GetByUserFile(userMate.User, hash)
	errCheck(g, err, "Upload:Failed to get file info", http.StatusInternalServerError)

	err = g.SaveUploadedFile(header, fileMate.Path)
	errCheck(g, err, "Upload:Failed to save file", http.StatusInternalServerError)

	fileMate.State = 1
	err = fileManager.Update(fileMate)
	errCheck(g, err, "Upload:Failed to update file state", http.StatusInternalServerError)

	userFileMate.State = 1
	err = userFileManager.Update(userFileMate)
	errCheck(g, err, "Upload:Failed to update user file state", http.StatusInternalServerError)

	g.JSON(http.StatusOK, gin.H{
		"errmsg": "ok",
		"data":   nil,
	})
}

// 秒传
func RapidUpload(g *gin.Context) {
	hash := g.PostForm("hash")
	userInter, _ := g.Get("userInfo")
	userMate := userInter.(models.UserInfo)

	// 判断用户是否已经关联该文件
	_, err := userFileManager.GetByUserFile(userMate.User, hash)
	if err != sql.ErrNoRows {
		g.JSON(http.StatusOK, gin.H{
			"errmsg": "RapidUpload:The file already exist",
			"data":   nil,
		})
		return
	}

	// 判断文件是否已存在
	fileMate, err := fileManager.GetByHash(hash)
	if err != sql.ErrNoRows {
		errCheck(g, err, "RapidUpload:Failed to read file info", http.StatusInternalServerError)
		userFileMate := models.UserFileMap{
			UserInfo: userMate,
			FileInfo: fileMate,
			FileName: fileMate.Name,
		}
		_, err = userFileManager.Set(userFileMate)
		errCheck(g, err, "RapidUpload:Failed to set user_file_map", http.StatusInternalServerError)

		g.JSON(http.StatusOK, gin.H{
			"errmsg": "ok",
			"data":   true,
		})
	}

	g.JSON(http.StatusOK, gin.H{
		"errmsg": "ok",
		"data":   false,
	})
}

// 下载文件
func Download(g *gin.Context) {
	fileHash := g.Query("fileHash")
	userInter, _ := g.Get("userInfo")
	userMate := userInter.(models.UserInfo)

	userFileMate, err := userFileManager.GetByUserFile(userMate.User, fileHash)
	if err == sql.ErrNoRows {
		g.JSON(http.StatusNotFound, gin.H{
			"errmsg": "Download:No files were found",
			"data":   nil,
		})
		return
	}
	errCheck(g, err, "Download:Failed to read user_file_map", http.StatusInternalServerError)

	_, err = os.Stat(userFileMate.FileInfo.Path)
	errCheck(g, err, "Download:Failed to find file", http.StatusNotFound)

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

	fileMate, err := fileManager.GetByHash(fileHash)
	if err == sql.ErrNoRows {
		g.JSON(http.StatusNotFound, gin.H{
			"errmsg": "Download:No files were found",
			"data":   nil,
		})
		return
	}
	errCheck(g, err, "PublicDownload:Failed to read file info", http.StatusInternalServerError)

	if fileMate.IsPublic != 1 {
		g.JSON(http.StatusUnauthorized, gin.H{
			"errmsg": "PublicDownload:No files were found",
			"data":   nil,
		})
		return
	}

	_, err = os.Stat(fileMate.Path)
	errCheck(g, err, "PublicDownload:Failed to find file", http.StatusNotFound)

	g.Writer.Header().Add(
		"Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileMate.Path))
	g.Writer.Header().Add("Content-Type", "application/octet-stream")
	g.File(fileMate.Path)
}

// 更新文件名称
func UpdateFileName(g *gin.Context) {
	fileHash := g.PostForm("fileHash")
	newFileName := g.PostForm("newFileName")
	userInter, _ := g.Get("userInfo")
	userMate := userInter.(models.UserInfo)

	userFileMate, err := userFileManager.GetByUserFile(userMate.User, fileHash)
	if err == sql.ErrNoRows {
		g.JSON(http.StatusNotFound, gin.H{
			"errmsg": "Download:No files were found",
			"data":   nil,
		})
		return
	}
	errCheck(g, err, "UpdateFileName:Failed to read user_file_map", http.StatusInternalServerError)

	userFileMate.FileName = newFileName

	err = userFileManager.Update(userFileMate)
	errCheck(g, err, "UpdateFileName:Failed to update file", http.StatusInternalServerError)

	g.JSON(http.StatusOK, gin.H{
		"errmsg": "ok",
		"data":   nil,
	})
}

// 删除文件
func Delete(g *gin.Context) {
	fileHash := g.PostForm("fileHash")
	userInter, _ := g.Get("userInfo")
	userMate := userInter.(models.UserInfo)

	userFileMate, err := userFileManager.GetByUserFile(userMate.User, fileHash)
	if err == sql.ErrNoRows {
		g.JSON(http.StatusNotFound, gin.H{
			"errmsg": "Delete:No files were found",
			"data":   nil,
		})
		return
	}
	errCheck(g, err, "Delete:Failed to read user_file_map", 0)

	err = userFileManager.DeleteByUserFile(userMate.User, fileHash)
	errCheck(g, err, "Delete:Failed to remove user_file_map", http.StatusInternalServerError)

	_, err = userFileManager.GetSqlByFile(fileHash)
	if err == sql.ErrNoRows {
		err = fileManager.DeleteByHash(fileHash)
		errCheck(g, err, "Delete:Failed to remove file_info", 0)
		err = os.Remove(userFileMate.FileInfo.Path)
		errCheck(g, err, "Delete:Failed to remove file", 0)
	}

	g.JSON(http.StatusOK, gin.H{
		"errmsg": "ok",
		"data":   nil,
	})
}

// 显示上传列表
func UploadShow(g *gin.Context) {
	pageStr := g.Query("page")
	userInter, _ := g.Get("userInfo")
	userMate := userInter.(models.UserInfo)

	page, err := strconv.Atoi(pageStr)
	errCheck(g, err, "UploadShow:Failed to conv page string", http.StatusInternalServerError)

	userFileList, err := userFileManager.GetByUser(userMate.User, page)
	if err == sql.ErrNoRows {
		g.JSON(http.StatusOK, gin.H{
			"errmsg": "UploadShow:No files were found",
			"data":   nil,
		})
		return
	}
	errCheck(g, err, "UploadShow:Failed to read user_file_map", http.StatusInternalServerError)

	fileBlockList, err := fileBlockManager.GetByUser(userMate.User)
	if err != nil && err != sql.ErrNoRows {
		errCheck(g, err, "UploadShow:Failed to read file_block_info", http.StatusInternalServerError)
	}

	var hashDict map[string]models.FileBlockInfo
	for _, fileBlockMate := range fileBlockList {
		hashDict[fileBlockMate.Hash] = fileBlockMate
	}

	var resList []map[string]interface{}
	for _, userFileMate := range userFileList {
		resDict := make(map[string]interface{})
		resDict["name"] = userFileMate.FileName
		resDict["hash"] = userFileMate.FileInfo.Hash
		resDict["progress"] = 0
		resDict["state"] = userFileMate.FileInfo.State

		fileSize := float64(userFileMate.FileInfo.Size)
		unitList := []string{"Bytes", "KB", "MB", "GB", "TB"}
		index := 0
		for index < len(unitList) {
			if fileSize > 1024 {
				fileSize = float64(fileSize) / float64(1024)
				index++
			} else {
				break
			}
		}
		resDict["size"] = fmt.Sprintf("%.2f", fileSize) + unitList[index]

		if uploadId := hashDict[userFileMate.FileInfo.Hash].UploadID; uploadId != "" {
			resDict["uploadId"] = uploadId
			blockMateList, err := blockManager.GetByUploadId(uploadId)
			errCheck(g, err, "blockUpload:Failed to get block info", http.StatusInternalServerError)

			amount := 0
			finishAmount := 0
			for _, blockMate := range blockMateList {
				amount++
				if blockMate.State == 1 {
					finishAmount++
				}
			}
			resDict["progress"] = fmt.Sprintf("%.2f", float64(finishAmount)/float64(amount))
		}

		resList = append(resList, resDict)
	}

	count, err := userFileManager.GetCountByUser(userMate.User)
	if err != nil && err != sql.ErrNoRows {
		errCheck(g, err, "UploadShow:Failed to read file_block_info", http.StatusInternalServerError)
	}

	g.JSON(http.StatusOK, gin.H{
		"errmsg": "ok",
		"data":   resList,
		"count":  count,
	})
}
