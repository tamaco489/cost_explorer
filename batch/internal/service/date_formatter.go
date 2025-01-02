package service

import "time"

// DailyReportDateFormatter: 日次コストレポートのための日時情報を保持する構造体
type DailyReportDateFormatter struct {
	Yesterday   string // 昨日の日付
	StartDate   string // 今月の開始日付
	EndDate     string // 今月の終了日付
	CurrentDay  int    // 今日までの日数
	DaysInMonth int    // 今月の総日数
}

// WeeklyReportDateFormatter: 週次コストレポートのための日時情報を保持する構造体
type WeeklyReportDateFormatter struct {
	LastWeekStartDate       string // 先週の開始日付
	LastWeekEndDate         string // 先週の終了日付
	WeekBeforeLastStartDate string // 先々週の開始日付
	WeekBeforeLastEndDate   string // 先々週の終了日付
}

// NewDailyReportDateFormatter: DailyReportDateFormatter のコンストラクタ
//
// 実行日時からコスト算出に必要な各基準日を取得
//
// Yesterday: 昨日の日付 (string)
//
// StartDate: 今月の開始日付 (string)
//
// EndDate: 今月の終了日付 (string)
//
// CurrentDay: 今日までの日数 (int)
//
// DaysInMonth: 今月の総日数 (int)
func (ds *DailyCostExplorerService) NewDailyReportDateFormatter(execTime time.Time) DailyReportDateFormatter {
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

// NewWeeklyReportDateFormatter: WeeklyReportDateFormatter のコンストラクタ
//
// lastWeekStartDate: 先週の開始日付 (string)
//
// lastWeekEndDate: 先週の終了日付 (string)
//
// weekBeforeLastStartDate: 先々週の開始日付 (string)
//
// weekBeforeLastEndDate: 先々週の終了日付 (string)
func (ws *WeeklyCostExplorerService) NewWeeklyReportDateFormatter(execTime time.Time) WeeklyReportDateFormatter {
	return WeeklyReportDateFormatter{
		LastWeekStartDate:       execTime.AddDate(0, 0, -13).Format("2006-01-02"),
		LastWeekEndDate:         execTime.AddDate(0, 0, -7).Format("2006-01-02"),
		WeekBeforeLastStartDate: execTime.AddDate(0, 0, -20).Format("2006-01-02"),
		WeekBeforeLastEndDate:   execTime.AddDate(0, 0, -14).Format("2006-01-02"),
	}
}
