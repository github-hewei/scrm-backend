package dashboard

// OverviewStats 平台核心指标
type OverviewStats struct {
	StoreTotal      int64 `json:"store_total"`       // 企业总数
	StoreMonthlyNew int64 `json:"store_monthly_new"` // 本月新增企业
	UserTotal       int64 `json:"user_total"`        // 租户用户总数
	UserMonthlyNew  int64 `json:"user_monthly_new"`  // 本月新增用户
	FileTotal       int64 `json:"file_total"`        // 文件总数
	FileTotalSize   int64 `json:"file_total_size"`   // 文件总存储占用(字节)
}

// DailyCount 按日统计
type DailyCount struct {
	Date  string `json:"date"`  // 日期(YYYY-MM-DD)
	Count int64  `json:"count"` // 当日数量
}

// TrendStats 近30天每日新增趋势
type TrendStats struct {
	Store []DailyCount `json:"store"` // 企业新增趋势
	User  []DailyCount `json:"user"`  // 用户新增趋势
}

// StatsResponse 仪表盘统计响应
type StatsResponse struct {
	Overview OverviewStats `json:"overview"`
	Trends   TrendStats    `json:"trends"`
}

// StoreOverviewStats 企业核心指标
type StoreOverviewStats struct {
	MemberTotal       int64 `json:"member_total"`        // 会员总数
	MemberMonthlyNew  int64 `json:"member_monthly_new"`  // 本月新增会员
	ArticleTotal      int64 `json:"article_total"`       // 文章总数
	ArticleMonthlyNew int64 `json:"article_monthly_new"` // 本月新增文章
	FileTotal         int64 `json:"file_total"`          // 文件总数
	FileTotalSize     int64 `json:"file_total_size"`     // 文件总存储占用(字节)
}

// StoreTrendStats 企业近30天每日新增趋势
type StoreTrendStats struct {
	Member  []DailyCount `json:"member"`  // 会员新增趋势
	Article []DailyCount `json:"article"` // 文章新增趋势
	File    []DailyCount `json:"file"`    // 文件上传趋势
}

// StoreStatsResponse 企业仪表盘统计响应
type StoreStatsResponse struct {
	Overview StoreOverviewStats `json:"overview"`
	Trends   StoreTrendStats    `json:"trends"`
}
