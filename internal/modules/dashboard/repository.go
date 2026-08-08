package dashboard

import (
	"context"
	"time"

	"github.com/241x/zero-kit/apperror"
	"github.com/241x/zero-web/errcode"
	"gorm.io/gorm"
)

// dailyCountRow 数据库每日聚合行
type dailyCountRow struct {
	Date string
	Cnt  int64
}

// Repository 仪表盘统计数据访问层
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建仪表盘统计数据访问层
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// CountStores 企业总数(平台)
func (r *Repository) CountStores(ctx context.Context, since int64) (int64, error) {
	return r.count(ctx, "gaz_rbac_store", 0, since)
}

// CountUsers 租户用户总数(平台)
func (r *Repository) CountUsers(ctx context.Context, since int64) (int64, error) {
	return r.count(ctx, "gaz_rbac_user", 0, since)
}

// CountMembers 会员总数
func (r *Repository) CountMembers(ctx context.Context, storeId uint32, since int64) (int64, error) {
	return r.count(ctx, "gaz_user", storeId, since)
}

// CountArticles 文章总数
func (r *Repository) CountArticles(ctx context.Context, storeId uint32, since int64) (int64, error) {
	return r.count(ctx, "gaz_article", storeId, since)
}

// CountFiles 文件总数
func (r *Repository) CountFiles(ctx context.Context, storeId uint32, since int64) (int64, error) {
	return r.count(ctx, "gaz_upload_file", storeId, since)
}

// SumFileSize 文件总存储占用(字节)
func (r *Repository) SumFileSize(ctx context.Context, storeId uint32) (int64, error) {
	var size int64
	query := r.db.WithContext(ctx).Table("gaz_upload_file").Where("deleted_at = 0")
	if storeId > 0 {
		query = query.Where("store_id = ?", storeId)
	}
	if err := query.Select("COALESCE(SUM(file_size), 0)").Scan(&size).Error; err != nil {
		return 0, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("统计文件存储失败"))
	}
	return size, nil
}

// DailyStores 企业每日新增趋势(平台)
func (r *Repository) DailyStores(ctx context.Context, since int64) ([]DailyCount, error) {
	return r.daily(ctx, "gaz_rbac_store", 0, since)
}

// DailyUsers 租户用户每日新增趋势(平台)
func (r *Repository) DailyUsers(ctx context.Context, since int64) ([]DailyCount, error) {
	return r.daily(ctx, "gaz_rbac_user", 0, since)
}

// DailyMembers 会员每日新增趋势
func (r *Repository) DailyMembers(ctx context.Context, storeId uint32, since int64) ([]DailyCount, error) {
	return r.daily(ctx, "gaz_user", storeId, since)
}

// DailyArticles 文章每日新增趋势
func (r *Repository) DailyArticles(ctx context.Context, storeId uint32, since int64) ([]DailyCount, error) {
	return r.daily(ctx, "gaz_article", storeId, since)
}

// DailyFiles 文件每日上传趋势
func (r *Repository) DailyFiles(ctx context.Context, storeId uint32, since int64) ([]DailyCount, error) {
	return r.daily(ctx, "gaz_upload_file", storeId, since)
}

// count 通用计数查询(storeId>0限定企业，since=0表示全部)
func (r *Repository) count(ctx context.Context, table string, storeId uint32, since int64) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Table(table).Where("deleted_at = 0")
	if storeId > 0 {
		query = query.Where("store_id = ?", storeId)
	}
	if since > 0 {
		query = query.Where("created_at >= ?", since)
	}
	if err := query.Count(&count).Error; err != nil {
		return 0, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("统计失败"))
	}
	return count, nil
}

// daily 通用按日聚合查询，并补全近 trendDays 天的日期序列(空日补零)
func (r *Repository) daily(ctx context.Context, table string, storeId uint32, since int64) ([]DailyCount, error) {
	rows := make([]dailyCountRow, 0)
	query := r.db.WithContext(ctx).Table(table).
		Select("FROM_UNIXTIME(created_at, '%Y-%m-%d') AS `date`, COUNT(*) AS `cnt`").
		Where("deleted_at = 0 AND created_at >= ?", since)
	if storeId > 0 {
		query = query.Where("store_id = ?", storeId)
	}
	if err := query.Group("`date`").Scan(&rows).Error; err != nil {
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
