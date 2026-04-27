package calendar

import "testing"

func TestLunarMonthDisplay(t *testing.T) {
	tests := []struct {
		month  int
		isLeap bool
		want   string
	}{
		{1, false, "正月"},
		{2, false, "二月"},
		{3, false, "三月"},
		{6, false, "六月"},
		{10, false, "十月"},
		{11, false, "冬月"},
		{12, false, "腊月"},
		{4, true, "闰四月"},
		{6, true, "闰六月"},
		{0, false, ""},
		{13, false, ""},
	}
	for _, tt := range tests {
		got := LunarMonthDisplay(tt.month, tt.isLeap)
		if got != tt.want {
			t.Errorf("LunarMonthDisplay(%d, %v) = %q, want %q", tt.month, tt.isLeap, got, tt.want)
		}
	}
}

func TestLunarDayDisplay(t *testing.T) {
	tests := []struct {
		day  int
		want string
	}{
		{1, "初一"}, {2, "初二"}, {9, "初九"}, {10, "初十"},
		{11, "十一"}, {15, "十五"}, {19, "十九"}, {20, "二十"},
		{21, "廿一"}, {29, "廿九"}, {30, "三十"},
		{0, ""}, {31, ""},
	}
	for _, tt := range tests {
		got := LunarDayDisplay(tt.day)
		if got != tt.want {
			t.Errorf("LunarDayDisplay(%d) = %q, want %q", tt.day, got, tt.want)
		}
	}
}

func TestLunarDisplay(t *testing.T) {
	tests := []struct {
		month  int
		day    int
		isLeap bool
		want   string
	}{
		{3, 4, false, "三月初四"},
		{1, 1, false, "正月初一"},
		{12, 30, false, "腊月三十"},
		{6, 1, true, "闰六月初一"},
	}
	for _, tt := range tests {
		got := LunarDisplay(tt.month, tt.day, tt.isLeap)
		if got != tt.want {
			t.Errorf("LunarDisplay(%d, %d, %v) = %q, want %q", tt.month, tt.day, tt.isLeap, got, tt.want)
		}
	}
}

func TestLunarYearDisplay(t *testing.T) {
	got := LunarYearDisplay("丙午", 3, 4, false)
	want := "丙午年三月初四"
	if got != want {
		t.Errorf("LunarYearDisplay(...) = %q, want %q", got, want)
	}
}
