package service

import (
	"time"
)

// DailyReportDateFormatter: 日次コストレポートのための日時情報を保持する構造体
type DailyReportDateFormatter struct {
	Yesterday   string // 昨日の日付
	StartDate   string // 今月の開始日付
	EndDate     string // 今月の終了日付
	CurrentDay  int    // 今日までの日数
	DaysInMonth int    // 今月の総日数
}

// NewDailyReportDateFormatter: DailyReportDateFormatter のコンストラクタ
//
// 実行日時からコスト算出に必要な各基準日を取得
//
// yesterday: 昨日の日付 (string)
//
// startDate: 今月の開始日付 (string)
//
// endDate: 今月の終了日付 (string)
//
// currentDay: 今日までの日数 (int)
//
// daysInMonth: 今月の総日数 (int)
func (s *DailyCostExplorerService) NewDailyReportDateFormatter(execTime time.Time) DailyReportDateFormatter {
	currentYear, currentMonth, _ := execTime.Date()
	daysInMonth := time.Date(currentYear, currentMonth+1, 0, 0, 0, 0, 0, time.UTC).Day()

	return DailyReportDateFormatter{
		Yesterday:   execTime.AddDate(0, 0, -1).Format("2006-01-02"),
		StartDate:   time.Date(execTime.Year(), execTime.Month(), 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02"),
		EndDate:     execTime.Format("2006-01-02"),
		CurrentDay:  execTime.Day(),
		DaysInMonth: daysInMonth,
	}
}
