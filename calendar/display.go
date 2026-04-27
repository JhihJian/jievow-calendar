package calendar

var lunarMonthNames = [...]string{
	"", "正月", "二月", "三月", "四月", "五月", "六月",
	"七月", "八月", "九月", "十月", "冬月", "腊月",
}

var lunarDayNames = [...]string{
	"", "初一", "初二", "初三", "初四", "初五", "初六", "初七", "初八", "初九", "初十",
	"十一", "十二", "十三", "十四", "十五", "十六", "十七", "十八", "十九", "二十",
	"廿一", "廿二", "廿三", "廿四", "廿五", "廿六", "廿七", "廿八", "廿九", "三十",
}

func LunarMonthDisplay(month int, isLeap bool) string {
	m := month
	if m < 0 {
		m = -m
	}
	if m < 1 || m > 12 {
		return ""
	}
	name := lunarMonthNames[m]
	if isLeap {
		name = "闰" + name
	}
	return name
}

func LunarDayDisplay(day int) string {
	if day < 1 || day > 30 {
		return ""
	}
	return lunarDayNames[day]
}

func LunarDisplay(month, day int, isLeap bool) string {
	return LunarMonthDisplay(month, isLeap) + LunarDayDisplay(day)
}

func LunarYearDisplay(yearGanzhi string, month, day int, isLeap bool) string {
	return yearGanzhi + "年" + LunarDisplay(month, day, isLeap)
}
