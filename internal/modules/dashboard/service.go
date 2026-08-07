package dashboard

import (
	"context"
	"time"

	"github.com/241x/zero-kit/apperror"
	"github.com/241x/zero-web/errcode"
	"gorm.io/gorm"
)

// trendDays 趋势统计窗口天数
const trendDays = 30

// dailyCountRow 数据库每日聚合行
type dailyCountRow struct {
	Date string
	Cnt  int64
}

// Service 仪表盘统计服务
type Service struct {
	db *gorm.DB
}

// NewService 创建仪表盘统计服务
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// Stats 获取平台仪表盘统计(企业/用户/文件核心指标与近30天新增趋势)
func (s *Service) Stats(ctx context.Context) (*StatsResponse, error) {
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Unix()
	trendStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -(trendDays - 1)).Unix()

	overview, err := s.countOverview(ctx, monthStart)
	if err != nil {
		return nil, err
	}

	storeTrend, err := s.countDaily(ctx, "gaz_rbac_store", trendStart)
	if err != nil {
		return nil, err
	}
	userTrend, err := s.countDaily(ctx, "gaz_rbac_user", trendStart)
	if err != nil {
		return nil, err
	}

	return &StatsResponse{
		Overview: overview,
		Trends: TrendStats{
			Store: storeTrend,
			User:  userTrend,
		},
	}, nil
}

// countOverview 统计核心指标
func (s *Service) countOverview(ctx context.Context, monthStart int64) (OverviewStats, error) {
	var stats OverviewStats
	var err error

	if stats.StoreTotal, err = s.countSince(ctx, "gaz_rbac_store", 0); err != nil {
		return stats, err
	}
	if stats.StoreMonthlyNew, err = s.countSince(ctx, "gaz_rbac_store", monthStart); err != nil {
		return stats, err
	}
	if stats.UserTotal, err = s.countSince(ctx, "gaz_rbac_user", 0); err != nil {
		return stats, err
	}
	if stats.UserMonthlyNew, err = s.countSince(ctx, "gaz_rbac_user", monthStart); err != nil {
		return stats, err
	}
	if stats.FileTotal, err = s.countSince(ctx, "gaz_upload_file", 0); err != nil {
		return stats, err
	}

	var fileSize int64
	err = s.db.WithContext(ctx).Table("gaz_upload_file").
		Where("deleted_at = 0").
		Select("COALESCE(SUM(file_size), 0)").
		Scan(&fileSize).Error
	if err != nil {
		return stats, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("统计文件存储失败"))
	}
	stats.FileTotalSize = fileSize

	return stats, nil
}

// countSince 统计表中 created_at 在 since 之后(含)且未删除的记录数，since=0 表示全部
func (s *Service) countSince(ctx context.Context, table string, since int64) (int64, error) {
	var count int64
	query := s.db.WithContext(ctx).Table(table).Where("deleted_at = 0")
	if since > 0 {
		query = query.Where("created_at >= ?", since)
	}
	if err := query.Count(&count).Error; err != nil {
		return 0, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("统计失败"))
	}
	return count, nil
}

// countDaily 统计自 since 以来按日新增数量，并补全近 trendDays 天的日期序列(空日补零)
func (s *Service) countDaily(ctx context.Context, table string, since int64) ([]DailyCount, error) {
	rows := make([]dailyCountRow, 0)
	err := s.db.WithContext(ctx).Table(table).
		Select("FROM_UNIXTIME(created_at, '%Y-%m-%d') AS `date`, COUNT(*) AS `cnt`").
		Where("deleted_at = 0 AND created_at >= ?", since).
		Group("`date`").
		Scan(&rows).Error
	if err != nil {
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("统计趋势失败"))
	}

	countMap := make(map[string]int64, len(rows))
	for _, row := range rows {
		countMap[row.Date] = row.Cnt
	}

	start := time.Unix(since, 0)
	result := make([]DailyCount, 0, trendDays)
	for i := range trendDays {
		date := start.AddDate(0, 0, i).Format("2006-01-02")
		result = append(result, DailyCount{Date: date, Count: countMap[date]})
	}
	return result, nil
}
