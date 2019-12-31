package api

import (
	"MyCloud/cloud_server/models"
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"math"
	"net/http"
	"strconv"
	"time"
)

func initBlockUpload(g *gin.Context) {
	hash := g.PostForm("hash")
	sizeStr := g.PostForm("size")
	userInter, _ := g.Get("userInfo")
	userMate := userInter.(models.UserInfo)

	fileSize, _ := strconv.Atoi(sizeStr)

	// 判断用户是否已经关联该文件
	_, err := userFileManager.GetSqlByUserFile(userMate.User, hash)
	if err != sql.ErrNoRows {
		g.JSON(http.StatusOK, gin.H{
			"errmsg": "Upload:The file already exist",
			"data":   nil,
		})
		return
	}

	// 判断文件是否已存在
	fileMate, err := fileManager.GetSqlByHash(hash)
	if err != sql.ErrNoRows {
		errCheck(g, err, "initBlockUpload:Failed to read file info", http.StatusInternalServerError)
		userFileMate := models.UserFileMap{
			UserInfo: userMate,
			FileInfo: fileMate,
			FileName: fileMate.Name,
		}
		_, err = userFileManager.Set(userMate.User+hash, userFileMate)
		errCheck(g, err, "Upload:Failed to set user_file_map", http.StatusInternalServerError)

		g.JSON(http.StatusOK, gin.H{
			"errmsg": "ok",
			"data":   nil,
		})
	}

	// 判断是否重复上传
	var resList [...]int

	fileBlockMate, err := fileBlockManager.GetByUserHash(userMate.User, hash)
	if err != sql.ErrNoRows {
		errCheck(g, err, "initBlockUpload:Failed to read file block info", http.StatusInternalServerError)
		blockList, err := blockManager.GetByUploadId(fileBlockMate.UploadID)
		errCheck(g, err, "initBlockUpload:Failed to read block info", http.StatusInternalServerError)

		for i, blockMate := range blockList {
			resList[i] = blockMate.State
		}
	} else {
		newFileBlockMate := models.FileBlockInfo{
			Hash:       hash,
			UserInfo:   userMate,
			UploadID:   userMate.User + fmt.Sprintf("%x", time.Now().UnixNano()),
			FileSize:   fileSize,
			BlockSize:  5 * 1024 * 1024,
			BlockCount: int(math.Ceil(float64(fileSize) / (5 * 1024 * 1024))),
		}
		_, err = fileBlockManager.Set("bu_"+userMate.User+hash, newFileBlockMate)

		for i := 0; i < newFileBlockMate.BlockCount; i++ {
			resList[i] = 0
		}
	}

	g.JSON(http.StatusOK, gin.H{
		"errmsg": "ok",
		"data":   resList,
	})
}
