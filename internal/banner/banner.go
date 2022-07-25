package banner

import (
	"fmt"
	"os"
	"runtime"
)

var (
	// BuildAt 编译时间, 遵循 time.RFC3339 格式
	// 转为 go 语言的 time.Time 示例:
	// buildAt, err := time.Parse(time.RFC3339, BuildAt)
	BuildAt string

	// Version 项目发布版本号
	// 项目每次发布版本后会打一个 tag, 这个版本号就来自 git 最新的 tag
	Version string

	// GitHead git 每次提交都会产生一个 id, 这个 GitHead 就来自于编译时最近的提交 id
	GitHead string
)

func Print() {
	fmt.Printf(logo, os.Getpid(), runtime.GOOS, runtime.GOARCH, Version, BuildAt, GitHead)
}
