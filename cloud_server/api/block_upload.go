package api

import (
	"MyCloud/cloud_server/models"
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"
)

func initBlockUpload(g *gin.Context) {
	hash := g.PostForm("hash")
	fileName := g.PostForm("fileName")
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
		fileBlockMate = models.FileBlockInfo{
			Hash:       hash,
			FileName:       fileName,
			UserInfo:   userMate,
			UploadID:   userMate.User + fmt.Sprintf("%x", time.Now().UnixNano()),
			FileSize:   fileSize,
			BlockSize:  5 * 1024 * 1024,
			BlockCount: int(math.Ceil(float64(fileSize) / (5 * 1024 * 1024))),
		}

		_, err = fileBlockManager.Set("fb_"+fileBlockMate.UploadID, fileBlockMate)
		errCheck(g, err, "initBlockUpload:Failed to set file block info", http.StatusInternalServerError)

		for i := 0; i < fileBlockMate.BlockCount; i++ {
			blockMate := models.BlockInfo{
				FileBlockInfo: fileBlockMate,
				Index:         i,
				Path:"./file/tmp"+fileBlockMate.UploadID+strconv.Itoa(i),
				Size:          5 * 1024 * 1024,
				State:         0,
			}
			if i + 1 == fileBlockMate.BlockCount{
				blockMate.Size = fileSize % (5 * 1024 * 1024)
			}
			_, err = blockManager.Set(
				"bl"+fileBlockMate.UploadID+strconv.Itoa(i),
				blockMate,
			)

			resList[i] = 0
		}
	}

	data := make(map[string]interface{})
	data["blockState"] = resList
	data["uploadId"] = fileBlockMate.UploadID

	g.JSON(http.StatusOK, gin.H{
		"errmsg": "ok",
		"data":   data,
	})
}

func blockUpload(g *gin.Context) {
	uploadId := g.PostForm("uploadId")
	index := g.PostForm("index")
	blockHeader, err := g.FormFile("block")
	errCheck(g, err, "blockUpload:Failed to get request", http.StatusBadRequest)

	blockMate, err := blockManager.GetByUploadIdIndex(uploadId, index)
	errCheck(g, err, "blockUpload:Failed to get block info", http.StatusBadRequest)
	if blockMate.State == 1 {
		g.JSON(http.StatusOK, gin.H{
			"errmsg": "ok",
			"data":   "",
		})
		return
	}

	err = g.SaveUploadedFile(blockHeader, blockMate.Path)
	errCheck(g, err, "blockUpload:Failed to save block", http.StatusInternalServerError)

	blockMate.State = 1
	err = blockManager.Update("bl_"+uploadId+index, blockMate)
	errCheck(g, err, "blockUpload:Failed to update block state", http.StatusBadRequest)

	g.JSON(http.StatusOK, gin.H{
		"errmsg": "ok",
		"data":   "",
	})
}

func uploadProgress(g *gin.Context) {
	uploadId := g.PostForm("uploadId")

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
	g.JSON(http.StatusOK, gin.H{
		"errmsg": "ok",
		"data": map[string]int{
			"amount":       amount,
			"finishAmount": finishAmount,
		},
	})
}


func blockMerge(g *gin.Context){
	uploadId := g.PostForm("uploadId")

	blockMateList, err := blockManager.GetByUploadId(uploadId)
	errCheck(g, err, "blockMerge:Failed to get block info", http.StatusInternalServerError)
	fileBlockMate, err := fileBlockManager.GetByUploadId(uploadId)
	errCheck(g, err, "blockMerge:Failed to get file block info", http.StatusInternalServerError)

	targetFile, err := os.OpenFile("./file/"+fileBlockMate.FileName+uploadId,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	errCheck(g, err, "blockMerge:Failed to open target file", http.StatusInternalServerError)
	defer targetFile.Close()
	for _, blockMate := range blockMateList {
		if blockMate.State == 0{
			err = os.Remove("./file/" + fileBlockMate.FileName + uploadId)
			errCheck(g, err, "blockMerge:Failed to remove target file", 0)
			g.JSON(http.StatusInternalServerError, gin.H{
				"errmsg": "blockMerge:Failed to incomplete file",
				"data": "",
			})
			return
		}
		block, err := os.Open(blockMate.Path)
		errCheck(g, err, "blockMerge:Failed to get file handler", http.StatusInternalServerError)
		b, err := ioutil.ReadAll(block)
		errCheck(g, err, "blockMerge:Failed to read file", http.StatusInternalServerError)
		_, err = targetFile.Write(b)
		errCheck(g, err, "blockMerge:Failed to write file", http.StatusInternalServerError)
		block.Close()
	}
	g.JSON(http.StatusOK, gin.H{
		"errmsg": "ok",
		"data": "",
	})
}