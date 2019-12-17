package meta

import "time"

// 文件信息结构
type FileMeta struct {
	FileName string
	FileSha1 string
	FileSize int64
	Location string
	Ctime string
}

var fileMates map[string]FileMeta

func NewFileMeta(name string, location string, ) *FileMeta {
	nowTime := time.Now().Format("1997-01-01 12:00:00")
	return &FileMeta{FileName: name, Location: location, Ctime: nowTime}
}

func init()  {
	fileMates = make(map[string]FileMeta)
}

// 新增或者更新文件信息
func UpdateFileMeta(fmeta *FileMeta){
	fileMates[fmeta.FileSha1] = *fmeta
}

// 通过sha1值获取文件对象
func GetFileMeta(fileSha1 string) *FileMeta{
	fmeta := fileMates[fileSha1]
	return &fmeta
}