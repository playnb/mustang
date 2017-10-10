package log

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/cihub/seelog"
)

var panicdir string
var logDir string
var logName string

func LogDir() string {
	return logDir
}

var DevLog = false
var LtpLog = false
var LwLog = false

func init() {
	flag.StringVar(&panicdir, "panic_dir", `C:\code\server\panic`, "日志位置")
	flag.StringVar(&logDir, "log_dir", `C:\code\server\log`, "日志位置")
	flag.StringVar(&logName, "log_name", `cell`, "日志文件名称")

	flag.Set("alsologtostderr", "true")
	flag.Set("log_dir", `C:\code\server\log`)
	flag.Set("panic_dir", `C:\code\server\panic`)
	flag.Set("v", "3")
}

type IUserLogStruct interface {
	GetAccount() string
	GetUserId() uint64
	GetUserName() string
	GetLevel() uint32
}

var userLog seelog.LoggerInterface

func Init() {
	if len(logName) == 0 {
		logName = "app"
	}

	logConfig := `
<seelog>
    <outputs formatid="main">
		<filter levels="info,critical,error,debug,trace">
		`
	//只有windows
	if runtime.GOOS == "windows" {
		logConfig += `<console />`
	}
	logConfig += `
			<rollingfile type="date" filename="` + logDir + `/` + fmt.Sprintf("%s-%d.log", logName, os.Getpid()) + `" datepattern="2006.01.02" />
        </filter>
    </outputs>

    <formats>
        <format id="main" format="%Date %Time [%LEV] [%File:%Line] %Msg%n"/>
    </formats>
</seelog>
`

	logger, _ := seelog.LoggerFromConfigAsBytes([]byte(logConfig))
	seelog.UseLogger(logger)

	logConfigData := `
<seelog>
    <outputs formatid="main">
        <filter levels="info,critical,error">
			<rollingfile type="date" filename="` + logDir + `/sobjserver.log` + `" datepattern="20060102-15" />
        </filter>
		<filter levels="debug">
			<rollingfile type="date" filename="` + logDir + `/objserver.log` + `" datepattern="20060102-15" />
        </filter>
		<filter levels="trace">
			<rollingfile type="date" filename="` + logDir + `/taskflow.log` + `" datepattern="20060102-15" />
        </filter>
    </outputs>

    <formats>
        <format id="main" format="%Date %Time [%LEV] [%File:%Line] %Msg%n"/>
    </formats>
</seelog>
`
	logger_data, _ := seelog.LoggerFromConfigAsBytes([]byte(logConfigData))
	seelog.UseLogger2(logger_data)

	{
		_logCfg := `
	<seelog>
		<outputs formatid="main">
			<filter levels="trace,debug,info,critical,error">
				`
		if runtime.GOOS == "windows" {
			_logCfg += `<console />`
		}
		_logCfg += `
				<rollingfile type="date" filename="` + logDir + `/users.log` + `" datepattern="20060102-15" />
			</filter>
		</outputs>
	
		<formats>
			<format id="main" format="%Date %Time [%LEV] [%File:%Line] %Msg%n"/>
		</formats>
	</seelog>
	`
		userLog, _ = seelog.LoggerFromConfigAsBytes([]byte(_logCfg))
	}
}

func Ltp(format string, a ...interface{}) {
	if LtpLog {
		seelog.Debugf(format, a...)
	}
}

func Dev(format string, a ...interface{}) {
	if DevLog {
		seelog.Debugf(format, a...)
	}
}

func Lw(format string, a ...interface{}) {
	if LwLog {
		seelog.Debugf(format, a...)
	}
}

func Debug(format string, a ...interface{}) {
	seelog.Debugf(format, a...)
}

func Info(format string, a ...interface{}) {
	seelog.Infof(format, a...)
}

func Trace(format string, a ...interface{}) {
	seelog.Tracef(format, a...)
}

func Warning(format string, a ...interface{}) {
	seelog.Warnf(format, a...)
}

func Error(format string, a ...interface{}) {
	seelog.Errorf(format, a...)
}

func Fatal(format string, a ...interface{}) {
	seelog.Criticalf(format, a...)
}

func InfoData(format string, a ...interface{}) {
	seelog.InfofData(format, a...)
}

func TraceData(format string, a ...interface{}) {
	seelog.TracefData(format, a...)
}

func DebugData(format string, a ...interface{}) {
	seelog.DebugfData(format, a...)
}

func UserLog(user IUserLogStruct, stepType int32, stepId int32, stepName string, activite int32) {
	if userLog == nil {
		return
	}

	userLog.Tracef("%s | %d | %s | %d | %s | %s | %d | %d | %d | %s | %d",
		"全区", 0, user.GetAccount(), user.GetUserId(), user.GetUserName(), "通用", user.GetLevel(), stepType, stepId, stepName, activite)
}

func Flush() {
	seelog.Flush()
}

/*
  <struct  name="StepFlow" version="1" desc="(可选)客户端点击步骤流水">
    <entry name="LogTime"    type="datetime"                               desc="(必填)游戏事件的时间, 格式 YYYY-MM-DD HH:MM:SS" />
    <entry name="ZoneName"     type="string"   size="32"                     desc="(必填)服务信息"，用来唯一标示一个区，以机房号命名如：YP01 />
    <entry name="ZoneID"    type="int"      index="1"   defaultvalue="0"  desc="(必填)针对分区分服的游戏填写分区id,用来唯一标示一个区；非分区分服游戏请填写0"/>
    <entry name="account"        type="string"   size="64"                     desc="(必填)玩家" />
    <entry name="vCharID"        type="string"        size="64"    defaultvalue="NULL"     desc="(必填)玩家角色ID"/>
    <entry name="vCharName"      type="string"        size="64"    defaultvalue="NULL"     desc="(必填)玩家角色名"/>
    <entry name="iCareer"        type="int"                                            desc="(必填)玩家职业,详细:iCareerType项目自定义" />
    <entry name="iLevel"         type="int"                                    desc="(必填)等级"/>
    <entry name="iStepType"      type="int"                                    desc="(必填)步骤类型,详细:iSTEPTYPE项目自定义" />
    <entry name="iStepID"        type="int"                                    desc="(必填)步骤ID" />
    <entry name="vStepName"      type="string"        size="64"    defaultvalue="NULL"     desc="(必填) 步骤名"/>
    <entry  name="iActivate"     type="int"                                                desc="批次" />

*/
