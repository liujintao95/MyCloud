package api

import (
	"MyCloud/cloud_server/models"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"
)

func InitBlockUpload(g *gin.Context) {
	hash := g.PostForm("hash")
	fileName := g.PostForm("fileName")
	sizeStr := g.PostForm("size")
	userInter, _ := g.Get("userInfo")
	userMate := userInter.(models.UserInfo)

	fileSize, _ := strconv.ParseInt(sizeStr, 10, 64)

	var resList [...]int

	fileBlockMate := models.FileBlockInfo{
		Hash:       hash,
		FileName:   fileName,
		UserInfo:   userMate,
		UploadID:   userMate.User + fmt.Sprintf("%x", time.Now().UnixNano()),
		FileSize:   fileSize,
		BlockSize:  5 * 1024 * 1024,
		BlockCount: int(math.Ceil(float64(fileSize) / (5 * 1024 * 1024))),
	}

	_, err := fileBlockManager.Set(fileBlockMate)
	errCheck(g, err, "initBlockUpload:Failed to set file block info", http.StatusInternalServerError)

	var blockList []models.BlockInfo
	for i := 0; i < fileBlockMate.BlockCount; i++ {
		blockMate := models.BlockInfo{
			FileBlockInfo: fileBlockMate,
			Index:         i,
			Path:          "./file/tmp/" + fileBlockMate.UploadID + strconv.Itoa(i),
			Size:          5 * 1024 * 1024,
			State:         0,
			Recycled:      "N",
		}
		if i+1 == fileBlockMate.BlockCount {
			blockMate.Size = fileSize % (5 * 1024 * 1024)
		}

		blockList = append(blockList, blockMate)
		resList[i] = 0
	}

	err = blockManager.Set(blockList)

	data := make(map[string]interface{})
	data["blockState"] = resList
	data["uploadId"] = fileBlockMate.UploadID

	g.JSON(http.StatusOK, gin.H{
		"errmsg": "ok",
		"data":   data,
	})
}

func ResumeFromBreakPoint(g *gin.Context) {
	uploadId := g.Query("uploadId")

	blockList, err := blockManager.GetByUploadId(uploadId)
	errCheck(g, err, "Resume: Failed to read block info", http.StatusInternalServerError)

	var resList [...]int
	for i, blockMate := range blockList {
		resList[i] = blockMate.State
	}

	data := make(map[string]interface{})
	data["blockState"] = resList
	data["uploadId"] = uploadId

	g.JSON(http.StatusOK, gin.H{
		"errmsg": "ok",
		"data":   data,
	})
}

func BlockUpload(g *gin.Context) {
	uploadId := g.PostForm("uploadId")
	indexStr := g.PostForm("index")
	blockHeader, err := g.FormFile("block")
	errCheck(g, err, "blockUpload:Failed to get request", http.StatusBadRequest)

	index, _ := strconv.Atoi(indexStr)
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
	err = blockManager.Update(blockMate)
	errCheck(g, err, "blockUpload:Failed to update block state", http.StatusBadRequest)

	g.JSON(http.StatusOK, gin.H{
		"errmsg": "ok",
		"data":   "",
	})
}

func UploadProgress(g *gin.Context) {
	uploadId := g.Query("uploadId")

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

func BlockMerge(g *gin.Context) {
	uploadId := g.PostForm("uploadId")
	userInter, _ := g.Get("userInfo")
	userMate := userInter.(models.UserInfo)

	blockMateList, err := blockManager.GetByUploadId(uploadId)
	errCheck(g, err, "blockMerge:Failed to get block info", http.StatusInternalServerError)
	fileBlockMate, err := fileBlockManager.GetByUploadId(uploadId)
	errCheck(g, err, "blockMerge:Failed to get file block info", http.StatusInternalServerError)

	targetFile, err := os.OpenFile("./file/"+uploadId+fileBlockMate.FileName,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	errCheck(g, err, "blockMerge:Failed to open target file", http.StatusInternalServerError)
	defer targetFile.Close()
	for _, blockMate := range blockMateList {
		if blockMate.State == 0 {
			err = os.Remove("./file/" + uploadId + fileBlockMate.FileName)
			errCheck(g, err, "blockMerge:Failed to remove target file", 0)
			g.JSON(http.StatusInternalServerError, gin.H{
				"errmsg": "blockMerge:Failed to incomplete file",
				"data":   "",
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

	fileBlockMate.State = 1
	err = fileBlockManager.Update(fileBlockMate)
	errCheck(g, err, "blockMerge:Failed to update file block info", http.StatusInternalServerError)

	fileMate := models.FileInfo{
		Name:     fileBlockMate.FileName,
		Size:     fileBlockMate.FileSize,
		Path:     "./file/" + uploadId + fileBlockMate.FileName,
		Hash:     fileBlockMate.Hash,
		IsPublic: 0,
	}
	userFileMate := models.UserFileMap{
		UserInfo: userMate,
		FileInfo: fileMate,
		FileName: fileBlockMate.FileName,
	}

	fileMate.Id, err = fileManager.Set(fileMate)
	errCheck(g, err, "blockMerge:Failed to set file information", http.StatusInternalServerError)
	_, err = userFileManager.Set(userFileMate)
	errCheck(g, err, "blockMerge:Failed to set user_file_map", http.StatusInternalServerError)

	g.JSON(http.StatusOK, gin.H{
		"errmsg": "ok",
		"data":   "",
	})
}

func RemoveBlock(g *gin.Context) {
	uploadId := g.PostForm("uploadId")

	fileBlockMate, err := fileBlockManager.GetByUploadId(uploadId)
	errCheck(g, err, "RemoveBlock: Failed to get file block info", http.StatusInternalServerError)
	if fileBlockMate.State != 1 {
		g.JSON(http.StatusOK, gin.H{
			"errmsg": "RemoveBlock: Failed to merge block",
			"data":   "",
		})
		return
	}

	blockMateList, err := blockManager.GetByUploadId(uploadId)
	errCheck(g, err, "RemoveBlock: Failed to get block info", http.StatusInternalServerError)
	for _, blockMate := range blockMateList {
		err = os.Remove(blockMate.Path)
		errCheck(g, err, "RemoveBlock: Failed to remove block file", 0)
	}

	g.JSON(http.StatusOK, gin.H{
		"errmsg": "ok",
		"data":   "",
	})
}
