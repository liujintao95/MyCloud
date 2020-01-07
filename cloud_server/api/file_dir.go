package api

import (
	"MyCloud/cloud_server/models"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
)

func ShowDir(g *gin.Context) {
	dirIdStr := g.Query("dirId")

	curDirId, err := strconv.Atoi(dirIdStr)
	errCheck(g, err, "Sign:Failed to convert dir id", http.StatusInternalServerError)

	dirList, err := dirManager.GetByFiD(curDirId)
	errCheck(g, err, "Sign:Failed to get dir list", http.StatusInternalServerError)

	var resList [][]string
	for _, dirMate := range dirList {
		msg := []string{
			strconv.FormatInt(dirMate.Id, 10),
			strconv.FormatInt(dirMate.Fid, 10),
			dirMate.UserFileMap.FileName,
			strconv.Itoa(dirMate.UserFileMap.Star),
			dirMate.UserFileMap.Remark.String,
			dirMate.UserFileMap.FileInfo.Hash,
			strconv.FormatInt(dirMate.UserFileMap.FileInfo.Size, 10),
			strconv.Itoa(dirMate.IsDir),
			dirMate.DirName.String,
		}
		resList = append(resList, msg)
	}

	g.JSON(http.StatusOK, gin.H{
		"errmsg": "ok",
		"data":   resList,
	})
}

func SaveDir(g *gin.Context) {
	dirList := g.PostFormArray("dirList")
	userInter, _ := g.Get("userInfo")
	userMate := userInter.(models.UserInfo)

	maxId, err := dirManager.GetSqlMaxId()
	errCheck(g, err, "SaveDir:Failed to get max id", http.StatusInternalServerError)

	var idMap = make(map[string]int64)
	var dirMateList []models.FileDirectory
	for _, val := range dirList {
		maxId++

		valList := strings.Split(val, "|")
		curDir := valList[0]
		fileName := valList[2]
		fileHash := valList[3]

		idMap[curDir+fileName+"/"] = maxId

		dirMate := models.FileDirectory{}
		dirMate.Id = maxId

		if curDir == "/" {
			dirMate.Fid = -1
		} else {
			dirMate.Fid = idMap[curDir]
		}

		if fileHash != "NULL" {
			userFileMate, err := userFileManager.GetByUserFile(userMate.User, fileHash)
			errCheck(g, err, "SaveDir:Failed to get file info", http.StatusInternalServerError)

			dirMate.UserFileMap = userFileMate
			dirMate.IsDir = 0
		} else {
			dirMate.DirName.String = fileName
			dirMate.DirName.Valid = true
			dirMate.IsDir = 1
		}

		dirMateList = append(dirMateList, dirMate)
	}
	err = dirManager.Set(dirMateList)
	errCheck(g, err, "SaveDir:Failed to set dir list", http.StatusInternalServerError)
}

func ChangeDir(g *gin.Context) {

}

func RemoveDir(g *gin.Context) {

}
