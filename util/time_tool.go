package util

import (
	"cell/common/mustang/log"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	January = 1 + iota
	February
	March
	April
	May
	June
	July
	August
	September
	October
	November
	December
)

var DaySeconds = 24 * 60 * 60

var DelayTime = int64(0)

var daysBefore = [...]int32{
	0,
	31,
	31 + 28,
	31 + 28 + 31,
	31 + 28 + 31 + 30,
	31 + 28 + 31 + 30 + 31,
	31 + 28 + 31 + 30 + 31 + 30,
	31 + 28 + 31 + 30 + 31 + 30 + 31,
	31 + 28 + 31 + 30 + 31 + 30 + 31 + 31,
	31 + 28 + 31 + 30 + 31 + 30 + 31 + 31 + 30,
	31 + 28 + 31 + 30 + 31 + 30 + 31 + 31 + 30 + 31,
	31 + 28 + 31 + 30 + 31 + 30 + 31 + 31 + 30 + 31 + 30,
	31 + 28 + 31 + 30 + 31 + 30 + 31 + 31 + 30 + 31 + 30 + 31,
}

func IsLeap(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

func NowTimestamp() int64 {
	return time.Now().Unix() + DelayTime
}

//App要求的毫秒级时间戳
func AppTimestamp() int64 {
	return time.Now().UnixNano()/int64(time.Millisecond) + DelayTime
}

func MonthDay(m int, year int) int {
	if !(m >= int(January) && m <= int(December)) {
		return 0
	}

	if m == int(February) && IsLeap(year) {
		return 29
	}
	return int(daysBefore[m] - daysBefore[m-1])
}

// 今日0点时间
func GetTodayZeroTime() time.Time {
	year, month, day := time.Now().Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.Local)
}

// 明日零点时间
func GetTomorrowZeroTime() time.Time {
	today_zero_time := GetTodayZeroTime()
	return today_zero_time.AddDate(0, 0, 1)
}

func GetCertainDayZeroTime(dayTimes int64) time.Time {
	year, month, day := time.Unix(dayTimes, 0).Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.Local)
}

// 是否同一天(按秒来计算)
func IsTheSameDay(rtime int64, ltime int64) bool {
	ryear, rmonth, rday := time.Unix(rtime, 0).Date()
	lyear, lmonth, lday := time.Unix(ltime, 0).Date()

	return time.Date(ryear, rmonth, rday, 0, 0, 0, 0, time.Local) == time.Date(lyear, lmonth, lday, 0, 0, 0, 0, time.Local)
}

// 本周一零点时间
func GetThisWeekMondayZeroTime() time.Time {
	today_zero_time := GetTodayZeroTime()
	weekday := today_zero_time.Weekday()
	if today_zero_time.Weekday() == 0 {
		weekday = 7
	}

	return today_zero_time.AddDate(0, 0, int(1)-int(weekday))
}

// 下周一零点时间
func GetNextWeekMondayZeroTime() time.Time {
	today_zero_time := GetTodayZeroTime()
	weekday := today_zero_time.Weekday()
	if today_zero_time.Weekday() == 0 {
		weekday = 7
	}

	return today_zero_time.AddDate(0, 0, int(8)-int(weekday))
}

// 某一天本周一零点时间
func GetCertainDayThisWeekMondayZeroTime(dayTimes int64) time.Time {
	year, month, day := time.Unix(dayTimes, 0).Date()
	day_zero_time := time.Date(year, month, day, 0, 0, 0, 0, time.Local)
	weekday := day_zero_time.Weekday()
	if day_zero_time.Weekday() == 0 {
		weekday = 7
	}

	return day_zero_time.AddDate(0, 0, int(1)-int(weekday))
}

// 某一天下周一零点时间
func GetCertainDayNextWeekMondayZeroTime(dayTimes int64) time.Time {
	year, month, day := time.Unix(dayTimes, 0).Date()
	day_zero_time := time.Date(year, month, day, 0, 0, 0, 0, time.Local)
	weekday := day_zero_time.Weekday()
	if day_zero_time.Weekday() == 0 {
		weekday = 7
	}

	return day_zero_time.AddDate(0, 0, int(8)-int(weekday))
}

// 是否同一周
func IsTheSameWeek(rtime int64, ltime int64) bool {
	rmonday_zero_time := GetCertainDayThisWeekMondayZeroTime(rtime).Unix()
	lmonday_zero_time := GetCertainDayThisWeekMondayZeroTime(ltime).Unix()

	//return rtoday_zero_time.AddDate(0, 0, int(8)-int(rweekday)) == ltoday_zero_time.AddDate(0, 0, int(8)-int(lweekday))
	return rmonday_zero_time == lmonday_zero_time
}

// 相差几周
func AbsWeeks(rtime int64, ltime int64) int64 {
	if rtime < ltime {
		temp := rtime
		rtime = ltime
		ltime = temp
	}

	rmonday_zero_time := GetCertainDayThisWeekMondayZeroTime(rtime).Unix()
	lmonday_zero_time := GetCertainDayThisWeekMondayZeroTime(ltime).Unix()

	return (rmonday_zero_time - lmonday_zero_time) / int64(7*24*60*60)
}

// 时间字符转换为时间戳  2018-01-05 12:00:00 ---->
func TimeStrToTimestamp(timeStr string) uint64 {
	if len(timeStr) == 0 {
		return 0
	}

	tm, _ := time.ParseInLocation("2006-01-02 15:04:05", timeStr, time.Local)
	return uint64(tm.Unix())
}

func TimestampToTimeStr(timestamp uint64) time.Time {

	tm := time.Unix(int64(timestamp), 0)
	return tm
}

// 获得每日时间字符串的时间戳 12:00:00 ---->
func DayTimeStrToTimestamp(timeStr string) uint64 {
	if len(timeStr) == 0 {
		return 0
	}

	ss := strings.Split(timeStr, ":")
	if len(ss) == 3 {
		hour, _ := strconv.ParseUint(ss[0], 10, 64)
		minute, _ := strconv.ParseUint(ss[1], 10, 64)
		second, _ := strconv.ParseUint(ss[2], 10, 64)

		return hour*60*60 + minute*60 + second
	}

	return 0
}

func FormatTimeCH(timestamp_msec int64) string {
	timestamp := timestamp_msec * 1000 * 1000

	var str string
	day := int64(time.Duration(timestamp).Hours() / 24)
	if day > 0 {
		str += fmt.Sprintf("%d天", day)
	}

	hour := int64(time.Duration(timestamp).Hours()) % 24
	if hour > 0 {
		str += fmt.Sprintf("%d小时", hour)
	}

	min := int64(time.Duration(timestamp).Minutes()) % 60
	if min > 0 {
		str += fmt.Sprintf("%d分钟", min)
	}
	return str
}

//获得当月0点时间
func GetFirstDateOfMonth(d time.Time) time.Time {
	d = d.AddDate(0, 0, -d.Day()+1)
	return GetZeroTime(d)
}

//获取传入的时间所在月份的最后一天，即某月最后一天的0点。如传入time.Now(), 返回当前月份的最后一天0点时间。
func GetLastDateOfMonth(d time.Time) time.Time {
	return GetFirstDateOfMonth(d).AddDate(0, 1, -1)
}

//获取某一天的0点时间
func GetZeroTime(d time.Time) time.Time {
	return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location())
}

////////////////////////////////////////////////////////////////////

func init() {
	go func() {
		for {
			_sysTime = time.Now().UnixNano()
			_sysTimeMillisecond = _sysTime / int64(time.Millisecond)
			time.Sleep(time.Millisecond)
		}
	}()

	go func() {
		tc := time.NewTicker(time.Second * 60)
		for {
			select {
			case <-tc.C:
				DumpFunctionCost()
			}
		}
	}()
}

////////////////////////////////////////////////////////////////////

var _sysTime int64
var _sysTimeMillisecond int64

type FunctionTime struct {
	Name      string
	beginTime int64
	thres     int64
}

func (ft *FunctionTime) End() {
	els := (_sysTimeMillisecond - ft.beginTime)
	if els > ft.thres {
		log.Trace("[FunctionTime] 超时|%s|%d|%d|%d", ft.Name, ft.beginTime, ft.thres, els)
	}
}

func NewFunctionTime(name string, thres int64) *FunctionTime {
	return &FunctionTime{
		Name:      name,
		beginTime: _sysTimeMillisecond,
		thres:     thres,
	}
}

////////////////////////////////////////////////////////////////////
//用于统计函数耗时

type _functionTime struct {
	msg          string
	cost         int64
	count        int64
	functionName string
}

func (f *_functionTime) String() string {
	t := ""
	n := f.cost / f.count
	i := 0
	for n > 0 {
		if i%3 == 0 && i != 0 {
			t = "," + t
		}
		t = strconv.FormatInt(int64(n%10), 10) + t
		i++

		n = n / 10
	}
	t = strings.Trim(t, ",")
	return fmt.Sprintf("(%d:\t%s)Function:%s msg:%s, count:%d AllCost:%d AverageCost:%d",
		f.count,
		t,
		f.functionName,
		f.msg,
		f.count,
		f.cost,
		f.cost/f.count)
}

var _allFunctionCost = make(map[string]*_functionTime)
var _allFunctionCostMutex sync.Mutex

func FunctionCost(msg string) func() {
	start := time.Now()
	return func() {
		_allFunctionCostMutex.Lock()
		if t, ok := _allFunctionCost[msg]; ok {
			t.cost += int64(time.Since(start))
			t.count += 1
		} else {
			_allFunctionCost[msg] = &_functionTime{
				cost:         int64(time.Since(start)),
				msg:          msg,
				count:        1,
				functionName: GetFuncName(3),
			}
		}
		_allFunctionCostMutex.Unlock()
	}
}

func DumpFunctionCost() {
	_allFunctionCostMutex.Lock()
	tmp := _allFunctionCost
	_allFunctionCost = make(map[string]*_functionTime)
	_allFunctionCostMutex.Unlock()

	log.Trace("[FunctionCost] DumpFunctionCost Begin")

	all := make([]*_functionTime, 0, len(tmp))
	for _, v := range tmp {
		all = append(all, v)
	}
	sort.Slice(all, func(i, j int) bool {
		return all[i].cost/(all[i].count+1) > all[j].cost/(all[j].count+1)
	})
	for _, v := range all {
		_ = v
		//log.Trace("[FunctionCost] %s", v.String())
		fmt.Printf("[FunctionCost] %s\n", v.String())
	}
	log.Trace("[FunctionCost] DumpFunctionCost End")
}
