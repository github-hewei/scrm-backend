package dashboard

import (
	"context"
	"time"
)

// trendDays 趋势统计窗口天数
const trendDays = 30

// Service 仪表盘统计服务
type Service struct {
	repo *Repository
}

// NewService 创建仪表盘统计服务
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Stats 获取平台仪表盘统计(企业/用户/文件核心指标与近30天新增趋势)
func (s *Service) Stats(ctx context.Context) (*StatsResponse, error) {
	monthStart, trendStart := timeWindows()

	overview, err := s.countOverview(ctx, monthStart)
	if err != nil {
		return nil, err
	}

	storeTrend, err := s.repo.DailyStores(ctx, trendStart)
	if err != nil {
		return nil, err
	}
	userTrend, err := s.repo.DailyUsers(ctx, trendStart)
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

// StoreStats 获取企业仪表盘统计(会员/文章/文件核心指标与近30天新增趋势，限定当前企业)
func (s *Service) StoreStats(ctx context.Context, storeId uint32) (*StoreStatsResponse, error) {
	monthStart, trendStart := timeWindows()

	var overview StoreOverviewStats
	var err error
	if overview.MemberTotal, err = s.repo.CountMembers(ctx, storeId, 0); err != nil {
		return nil, err
	}
	if overview.MemberMonthlyNew, err = s.repo.CountMembers(ctx, storeId, monthStart); err != nil {
		return nil, err
	}
	if overview.ArticleTotal, err = s.repo.CountArticles(ctx, storeId, 0); err != nil {
		return nil, err
	}
	if overview.ArticleMonthlyNew, err = s.repo.CountArticles(ctx, storeId, monthStart); err != nil {
		return nil, err
	}
	if overview.FileTotal, err = s.repo.CountFiles(ctx, storeId, 0); err != nil {
		return nil, err
	}
	if overview.FileTotalSize, err = s.repo.SumFileSize(ctx, storeId); err != nil {
		return nil, err
	}

	memberTrend, err := s.repo.DailyMembers(ctx, storeId, trendStart)
	if err != nil {
		return nil, err
	}
	articleTrend, err := s.repo.DailyArticles(ctx, storeId, trendStart)
	if err != nil {
		return nil, err
	}
	fileTrend, err := s.repo.DailyFiles(ctx, storeId, trendStart)
	if err != nil {
		return nil, err
	}

	return &StoreStatsResponse{
		Overview: overview,
		Trends: StoreTrendStats{
			Member:  memberTrend,
			Article: articleTrend,
			File:    fileTrend,
		},
	}, nil
}

// countOverview 统计平台核心指标
func (s *Service) countOverview(ctx context.Context, monthStart int64) (OverviewStats, error) {
	var stats OverviewStats
	var err error

	if stats.StoreTotal, err = s.repo.CountStores(ctx, 0); err != nil {
		return stats, err
	}
	if stats.StoreMonthlyNew, err = s.repo.CountStores(ctx, monthStart); err != nil {
		return stats, err
	}
	if stats.UserTotal, err = s.repo.CountUsers(ctx, 0); err != nil {
		return stats, err
	}
	if stats.UserMonthlyNew, err = s.repo.CountUsers(ctx, monthStart); err != nil {
		return stats, err
	}
	if stats.FileTotal, err = s.repo.CountFiles(ctx, 0, 0); err != nil {
		return stats, err
	}
	if stats.FileTotalSize, err = s.repo.SumFileSize(ctx, 0); err != nil {
		return stats, err
	}

	return stats, nil
}

// timeWindows 计算本月起点与趋势窗口起点时间戳
func timeWindows() (monthStart, trendStart int64) {
	now := time.Now()
	loc := now.Location()
	monthStart = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc).Unix()
	trendStart = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, -(trendDays - 1)).Unix()
	return
}
