package version

import "fmt"

//应用程序的版本信息
//用如下方式编译主语信息
/*
CurrentVersion=0.0.1

#Path是指向version的import路径
Path="github.com/playnb/mustang/version"
GitCommit=$(git rev-parse --short HEAD || echo unsupported)

go build -ldflags "-X $Path.Version=$CurrentVersion -X '$Path.BuildTime=`date "+%Y-%m-%d %H:%M:%S"`' -X '$Path.GoVersion=`go version`' -X $Path.GitCommit=$GitCommit"

*/
var (
	// Version should be updated by hand at each release
	Version = "0.0.0"

	GitCommit string
	GoVersion string
	BuildTime string
)

func FullVersion() string {
	return fmt.Sprintf("Version: %6s \nGit commit: %6s \nGo version: %6s \nBuild time: %6s \n",
		Version, GitCommit, GoVersion, BuildTime)
}
