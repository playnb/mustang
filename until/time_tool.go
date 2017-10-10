package until

import "time"

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

// 是否同一周
func IsTheSameWeek(rtime int64, ltime int64) bool {
	ryear, rmonth, rday := time.Unix(rtime, 0).Date()
	rtoday_zero_time := time.Date(ryear, rmonth, rday, 0, 0, 0, 0, time.Local)
	rweekday := rtoday_zero_time.Weekday()
	if rtoday_zero_time.Weekday() == 0 {
		rweekday = 7
	}

	lyear, lmonth, lday := time.Unix(ltime, 0).Date()
	ltoday_zero_time := time.Date(lyear, lmonth, lday, 0, 0, 0, 0, time.Local)
	lweekday := ltoday_zero_time.Weekday()
	if ltoday_zero_time.Weekday() == 0 {
		lweekday = 7
	}

	return rtoday_zero_time.AddDate(0, 0, int(8)-int(rweekday)) == ltoday_zero_time.AddDate(0, 0, int(8)-int(lweekday))
}
