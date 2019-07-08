package log

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"

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
	flag.StringVar(&panicdir, "panic_dir", `../panic`, "日志位置")
	flag.StringVar(&logDir, "log_dir", `../log`, "日志位置")
	flag.StringVar(&logName, "log_name", `cell`, "日志文件名称")

	flag.Set("alsologtostderr", "true")
	flag.Set("v", "3")
}

type IUserLog interface {
	GetAccount() string
	GetAccid() string
	GetUserId() uint64
	GetUserName() string
	GetLevel() uint32
}

var userLog seelog.LoggerInterface
var billLog seelog.LoggerInterface
var sobjLog seelog.LoggerInterface
var objLog seelog.LoggerInterface
var taskflowLog seelog.LoggerInterface
var luaLog seelog.LoggerInterface

func Init() {
	var err error
	if len(logName) == 0 {
		logName = "app"
	}

	logConfig := `
<seelog>
    <outputs formatid="main">
		<filter levels="info,critical,error,debug,trace,warn">
		`
	//只有windows
	if runtime.GOOS == "windows" {
		logConfig += `<console />`
	}
	logConfig += `
			<rollingfile type="date" filename="` + logDir + `/` + fmt.Sprintf("%s.log", logName) + `" datepattern="2006.01.02-15" />
        </filter>
    </outputs>

    <formats>
        <format id="main" format="%Date %Time [%LEV] [%File:%Line] %Msg%n"/>
    </formats>
</seelog>
`

	logger, _ := seelog.LoggerFromConfigAsBytes([]byte(logConfig))
	seelog.UseLogger(logger)
	fieldLog = NewFieldLog(logger)

	// 	logConfigData := `
	// <seelog>
	//     <outputs formatid="main">
	// 		<filter levels="debug">
	// 			<rollingfile type="date" filename="` + logDir + fmt.Sprintf(`/objserver-%d.log`, os.Getpid()) + `" datepattern="20060102-15" />
	//         </filter>
	//     </outputs>

	//     <formats>
	//         <format id="main" format="%Date %Time [%LEV] [%File:%Line] %Msg%n"/>
	//     </formats>
	// </seelog>
	// `
	// 	logger_data, _ := seelog.LoggerFromConfigAsBytes([]byte(logConfigData))
	// 	seelog.Trace(logConfigData)
	// 	seelog.UseLogger2(logger_data)

	{
		_logCfg := `
<seelog>
    <outputs formatid="main">
		<filter levels="info,critical,error,debug,trace,warn">
		`
		//只有windows
		if runtime.GOOS == "windows" {
			_logCfg += `<console />`
		}
		_logCfg += `
			<rollingfile type="date" filename="` + logDir + `/` + fmt.Sprintf("%s.log", "user-flow") + `" datepattern="2006.01.02-15" />
        </filter>
    </outputs>

    <formats>
        <format id="main" format="%Date %Time %Msg%n"/>
    </formats>
</seelog>
	`
		userLog, err = seelog.LoggerFromConfigAsBytes([]byte(_logCfg))
		if err != nil {
			fmt.Println(_logCfg)
			fmt.Println(err)
		}
		userFieldLog = NewFieldLog(userLog)
	}

	{
		_logCfg := `
	<seelog>
		<outputs formatid="main">
			<filter levels="trace,debug,info,critical,error,warn">
				`
		if runtime.GOOS == "windows" {
			_logCfg += `<console />`
		}
		_logCfg += `
				<rollingfile type="date" filename="` + logDir + fmt.Sprintf(`/trade-%d.log`, os.Getpid()) + `" datepattern="20060102-15" />
			</filter>
		</outputs>
	
		<formats>
			<format id="main" format="%Date %Time [%LEV] [%File:%Line] %Msg%n"/>
		</formats>
	</seelog>
	`
		billLog, _ = seelog.LoggerFromConfigAsBytes([]byte(_logCfg))
		//seelog.Trace(_logCfg)
	}

	{
		_logCfg := `
	<seelog>
		<outputs formatid="main">
			<filter levels="trace,debug,info,critical,error,warn">
				`
		if runtime.GOOS == "windows" {
			_logCfg += `<console />`
		}
		_logCfg += `
				<rollingfile type="date" filename="` + logDir + fmt.Sprintf(`/sobjserver-%d.log`, os.Getpid()) + `" datepattern="20060102-15" />
			</filter>
		</outputs>
	
		<formats>
			<format id="main" format="%Date %Time [%LEV] [%File:%Line] %Msg%n"/>
		</formats>
	</seelog>
	`
		sobjLog, _ = seelog.LoggerFromConfigAsBytes([]byte(_logCfg))
		//seelog.Trace(_logCfg)
	}

	{
		_logCfg := `
	<seelog>
		<outputs formatid="main">
			<filter levels="trace,debug,info,critical,error,warn">
				`
		if runtime.GOOS == "windows" {
			_logCfg += `<console />`
		}
		_logCfg += `
				<rollingfile type="date" filename="` + logDir + fmt.Sprintf(`/objserver-%d.log`, os.Getpid()) + `" datepattern="20060102-15" />
			</filter>
		</outputs>
	
		<formats>
			<format id="main" format="%Date %Time [%LEV] [%File:%Line] %Msg%n"/>
		</formats>
	</seelog>
	`
		objLog, _ = seelog.LoggerFromConfigAsBytes([]byte(_logCfg))
		//seelog.Trace(_logCfg)
	}

	{
		_logCfg := `
	<seelog>
		<outputs formatid="main">
			<filter levels="trace,debug,info,critical,error,warn">
				`
		if runtime.GOOS == "windows" {
			_logCfg += `<console />`
		}
		_logCfg += `
				<rollingfile type="date" filename="` + logDir + fmt.Sprintf(`/taskflow2-%d.log`, os.Getpid()) + `" datepattern="20060102-15" />
			</filter>
		</outputs>
	
		<formats>
			<format id="main" format="%Date %Time [%LEV] [%File:%Line] %Msg%n"/>
		</formats>
	</seelog>
	`
		taskflowLog, _ = seelog.LoggerFromConfigAsBytes([]byte(_logCfg))
		//seelog.Trace(_logCfg)
	}

	{
		_logCfg := `
	<seelog>
		<outputs formatid="main">
			<filter levels="trace,debug,info,critical,error,warn">
				`
		if runtime.GOOS == "windows" {
			_logCfg += `<console />`
		}
		_logCfg += `
				<rollingfile type="date" filename="` + logDir + fmt.Sprintf(`/lua-%d.log`, os.Getpid()) + `" datepattern="20060102" />
			</filter>
		</outputs>
	
		<formats>
			<format id="main" format="%Date %Time [%LEV] [%File:%Line] %Msg%n"/>
		</formats>
	</seelog>
	`
		luaLog, _ = seelog.LoggerFromConfigAsBytes([]byte(_logCfg))
		//seelog.Trace(_logCfg)
	}

	Debug("===========================>LOAD LOG CONFIG(init)")
}

type UserAction string

const (
	UserAction_PostMoment         = UserAction("发表动态")
	UserAction_DelMoment          = UserAction("删除动态")
	UserAction_PostComment        = UserAction("发表评论")
	UserAction_DelComment         = UserAction("删除评论")
	UserAction_PatchUserInfo      = UserAction("修改个人信息")
	UserAction_PatchGameCharactor = UserAction("修改展示游戏角色")
	UserAction_LikeMoment         = UserAction("点赞动态")
	UserAction_UnLikeMoment       = UserAction("取消点赞动态")
	UserAction_LikeComment        = UserAction("点赞评论")
	UserAction_UnLikeComment      = UserAction("取消点赞评论")
	UserAction_FollowSomebody     = UserAction("关注某人")
	UserAction_UnFollowSomebody   = UserAction("取消关注某人")
)

type UserFlowTrace struct {
	Accid     uint32      `json:"accid"`
	Action    UserAction  `json:"action"`
	Target    interface{} `json:"target"`
	MomentID  string      `json:"moment_id"`
	CommentID string      `json:"comment_id"`
	Content   interface{} `json:"content"`
}

func UserTrace(uft *UserFlowTrace) {
	UserLog().WithField(Fields{
		"accid":  uft.Accid,
		"action": uft.Action,
		"target": uft.Target,

		"moment_id":  uft.MomentID,
		"comment_id": uft.CommentID,
		"content":    uft.Content,
		"timestamp":  time.Now().Unix(),
	}).Log("")
}

func Ltp(format string, a ...interface{}) {
	if LtpLog {
		//seelog.Debugf(format, a...)
		seelog.Debug(fmt.Sprintf(format, a...))
	}
}

func Dev(format string, a ...interface{}) {
	//seelog.Debugf(format, a...)
	seelog.Debug(fmt.Sprintf(format, a...))
}

func Lw(format string, a ...interface{}) {
	if LwLog {
		//seelog.Debugf(format, a...)
		seelog.Debug(fmt.Sprintf(format, a...))
	}
}

func Debug(format string, a ...interface{}) {
	//seelog.Debugf(format, a...)
	seelog.Debug(fmt.Sprintf(format, a...))
}

func Info(format string, a ...interface{}) {
	//seelog.Infof(format, a...)
	seelog.Info(fmt.Sprintf(format, a...))
}

func Trace(format string, a ...interface{}) {
	//seelog.Tracef(format, a...)
	seelog.Trace(fmt.Sprintf(format, a...))
}

func Warning(format string, a ...interface{}) {
	//seelog.Warnf(format, a...)
	seelog.Warn(fmt.Sprintf(format, a...))
}

func Error(format string, a ...interface{}) {
	//seelog.Errorf(format, a...)
	seelog.Error(fmt.Sprintf(format, a...))
}

func Fatal(format string, a ...interface{}) {
	//seelog.Criticalf(format, a...)
	seelog.Critical(fmt.Sprintf(format, a...))
}

// func InfoData(format string, a ...interface{}) {
// 	//seelog.InfofData(format, a...)
// 	seelog.InfofData("%s", fmt.Sprintf(format, a...))
// }

// func TraceData(format string, a ...interface{}) {
// 	//seelog.TracefData(format, a...)
// 	seelog.TracefData("%s", fmt.Sprintf(format, a...))
// }

// func DebugData(format string, a ...interface{}) {
// 	//seelog.DebugfData(format, a...)
// 	seelog.DebugfData("%s", fmt.Sprintf(format, a...))
// }

func SObjLog(format string, a ...interface{}) {
	sobjLog.Infof(fmt.Sprintf(format, a...))
}

func ObjLog(format string, a ...interface{}) {
	objLog.Tracef(fmt.Sprintf(format, a...))
}

func TaskLog(format string, a ...interface{}) {
	taskflowLog.Debugf(fmt.Sprintf(format, a...))
}

func LuaLog(format string, a ...interface{}) {
	luaLog.Debugf(fmt.Sprintf(format, a...))
}

func GaBillLog(params map[string]string, format string, a ...interface{}) {
	_getMapStr := func(param map[string]string, key string) string {
		if p, ok := param[key]; ok {
			return p
		}
		return ""
	}

	billLog.Infof("GABill|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s\n",
		_getMapStr(params, "account"),
		_getMapStr(params, "amount"),
		_getMapStr(params, "order_id"),
		_getMapStr(params, "product_id"),
		_getMapStr(params, "extra"),
		_getMapStr(params, "time"),
		_getMapStr(params, "transaction_id"),
		_getMapStr(params, "openid"),
		_getMapStr(params, "zone_id"),
		_getMapStr(params, "channel"),
		_getMapStr(params, "game_id"),
		fmt.Sprintf(format, a...))
}

func Flush() {
	seelog.Flush()
	if userLog != nil {
		userLog.Flush()
	}
	if billLog != nil {
		billLog.Flush()
	}
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
