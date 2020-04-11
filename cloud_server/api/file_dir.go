package api

import (
	"MyCloud/cloud_server/models"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func ShowDir(g *gin.Context) {
	dirIdStr := g.Query("dirId")

	curDirId, err := strconv.ParseInt(dirIdStr, 64, 10)
	errCheck(g, err, "Sign:Failed to convert dir id", http.StatusInternalServerError)

	dirList, err := dirManager.GetByFid(curDirId)
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
			strconv.Itoa(dirMate.IsFolder),
			dirMate.DirName,
		}
		resList = append(resList, msg)
	}

	g.JSON(http.StatusOK, gin.H{
		"errmsg": "ok",
		"data":   resList,
	})
}

func SaveDir(g *gin.Context) {
	dirStr := g.PostForm("dirList")
	userInter, _ := g.Get("userInfo")
	userMate := userInter.(models.UserInfo)

	type dirMap struct {
		CurDir   string `json:"curDir"`
		FileName string `json:"fileName"`
		FileHash string `json:"fileHash"`
	}
	var dirList []dirMap
	err := json.Unmarshal([]byte(dirStr), &dirList)
	errCheck(g, err, "SaveDir:Failed to Unmarshal json", http.StatusInternalServerError)

	maxId, err := dirManager.GetSqlMaxId()
	errCheck(g, err, "SaveDir:Failed to get max id", http.StatusInternalServerError)

	var idMap = make(map[string]int64)
	var dirMateList []models.FileDirectory
	for _, val := range dirList {
		maxId++

		idMap[val.CurDir+val.FileName] = maxId

		dirMate := models.FileDirectory{}
		dirMate.Id = maxId
		dirMate.DirName = val.CurDir + val.FileName

		if val.CurDir == "" {
			dirMate.Fid = -1
		} else {
			dirMate.Fid = idMap[val.CurDir]
		}

		if val.FileHash != "" {
			userFileMate, err := userFileManager.GetByUserFile(userMate.User, val.FileHash)
			errCheck(g, err, "SaveDir:Failed to get file info", http.StatusInternalServerError)

			dirMate.UserFileMap = userFileMate
			dirMate.IsFolder = 0
		} else {
			dirMate.IsFolder = 1
		}

		dirMateList = append(dirMateList, dirMate)
	}
	err = dirManager.Set(dirMateList)
	errCheck(g, err, "SaveDir:Failed to set dir list", http.StatusInternalServerError)

	g.JSON(http.StatusOK, gin.H{
		"errmsg": "ok",
		"data":   "",
	})
}

func ChangeDir(g *gin.Context) {
	idStr := g.PostForm("id")
	targetIdStr := g.PostForm("targetId")

	id, err := strconv.ParseInt(idStr, 64, 10)
	targetId, err := strconv.ParseInt(targetIdStr, 64, 10)
	errCheck(g, err, "ChangeDir:Failed to convert dir id", http.StatusInternalServerError)

	dirMate, err := dirManager.GetSqlById(id)
	errCheck(g, err, "ChangeDir:Failed to get dir info", http.StatusInternalServerError)

	dirMate.Fid = targetId

	err = dirManager.Update(dirMate)
	errCheck(g, err, "ChangeDir:Failed to update dir info", http.StatusInternalServerError)

	g.JSON(http.StatusOK, gin.H{
		"errmsg": "ok",
		"data":   "",
	})
}

func RemoveDir(g *gin.Context) {
	idStr := g.PostForm("id")

	id, err := strconv.ParseInt(idStr, 64, 10)
	errCheck(g, err, "RemoveDir:Failed to convert dir id", http.StatusInternalServerError)

	err = dirManager.DeleteById(id)
	errCheck(g, err, "RemoveDir:Failed to del dir info", http.StatusInternalServerError)

	g.JSON(http.StatusOK, gin.H{
		"errmsg": "ok",
		"data":   "",
	})
}
