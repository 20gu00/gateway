package cert

import (
	"path/filepath"
	"runtime"
)

// basepath is the root directory of this package.
var basepath string

func init() {
	_, currentFile, _, _ := runtime.Caller(0) //Caller()报告当前go程调用栈所执行的函数的文件和行号信息。skip上溯的栈帧数，0表示Caller的调用者（Caller所在的调用栈）（0-当前函数，1-上一层函数，…）。
	//值是当前文件的目录路径
	basepath = filepath.Dir(currentFile)
}

func Path(rel string) string {
	if filepath.IsAbs(rel) { //判断返回路径是否是一个绝对路径。
		return rel
	}

	return filepath.Join(basepath, rel)
}
