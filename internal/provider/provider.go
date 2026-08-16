package provider

import (
	"zero-backend/internal/config"
	"zero-backend/internal/modules/captcha"
	"zero-backend/internal/modules/setting"
	wecomconfig "zero-backend/internal/modules/wecom/config"
	"zero-backend/internal/modules/wecom/contact"
	"zero-backend/internal/modules/wecom/customer"
	"zero-backend/internal/modules/wecom/group"
	wecomsync "zero-backend/internal/modules/wecom/sync"

	"github.com/241x/zero-kit/apperror"
	"github.com/241x/zero-kit/logger"
	"github.com/241x/zero-third/wecom"
	"github.com/241x/zero-web/errcode"
	"github.com/241x/zero-web/server"
	goredis "github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

// NewLogger 加载日志组件。filename 为空时使用配置文件中的默认文件名。
func NewLogger(db *mongo.Database, filename string) *logger.ZeroLogger {
	type fileLogConfigAdapter struct {
		Path       string `mapstructure:"path"`
		Filename   string `mapstructure:"filename"`
		MaxSize    int    `mapstructure:"max_size"`
		MaxAge     int    `mapstructure:"max_age"`
		MaxBackups int    `mapstructure:"max_backups"`
		Compress   bool   `mapstructure:"compress"`
		LocalTime  bool   `mapstructure:"local_time"`
	}
	type adapter struct {
		Level   string               `mapstructure:"level"`
		Writers []string             `mapstructure:"writers"`
		File    fileLogConfigAdapter `mapstructure:"file"`
	}
	var a adapter
	config.UnmarshalKey("logger", &a)
	if filename != "" {
		a.File.Filename = filename
	}

	options := []logger.Option{}
	for _, writer := range a.Writers {
		switch writer {
		case "console":
			options = append(options, logger.WithConsole())
		case "file":
			options = append(options, logger.WithFileWithConfig(logger.FileConfig{
				Path:       a.File.Path,
				Filename:   a.File.Filename,
				MaxSize:    a.File.MaxSize,
				MaxAge:     a.File.MaxAge,
				MaxBackups: a.File.MaxBackups,
				Compress:   a.File.Compress,
				LocalTime:  a.File.LocalTime,
			}))
		case "mongodb":
			options = append(options, logger.WithMongo(db))
		}
	}
	level := logger.Disabled
	switch a.Level {
	case "info":
		level = logger.InfoLevel
	case "debug":
		level = logger.DebugLevel
	case "warn":
		level = logger.WarnLevel
	case "error":
		level = logger.ErrorLevel
	}

	options = append(options, logger.WithLevel(level))
	return logger.New(options...)
}

// ProvideBindErrCode 提供绑定组件错误码
func ProvideBindErrCode() apperror.Code {
	return errcode.InvalidInput
}

// ProvideServerOptions 提供服务器选项
func ProvideServerOptions() []server.Option {
	return []server.Option{}
}

// NewSettingService 创建设置服务
func NewSettingService(db *gorm.DB) *setting.Service {
	return setting.NewService(setting.NewRepository(db), setting.NewDefaultRepository(db))
}

// NewCaptchaService 创建验证码服务
func NewCaptchaService(rdb *goredis.Client, cfg captcha.Config) (*captcha.Service, error) {
	return captcha.NewService(rdb, cfg, "ZAG:CAPTCHA")
}

// MustNewCaptchaService 创建验证码服务
func MustNewCaptchaService(rdb *goredis.Client, cfg captcha.Config) *captcha.Service {
	svc, err := NewCaptchaService(rdb, cfg)
	if err != nil {
		panic(err)
	}
	return svc
}

// NewWecomSyncService 组装企微数据同步服务（CLI/worker/web 共用同一内核）
func NewWecomSyncService(db *gorm.DB, rdb *goredis.Client) *wecomsync.Service {
	configRepo := wecomconfig.NewWecomConfigRepository(db)
	appRepo := wecomconfig.NewWecomAppRepository(db)
	clientMgr := wecomsync.NewClientManager(configRepo, appRepo, wecom.NewRedisCache(rdb))
	contactSyncer := contact.NewContactSyncer(
		contact.NewWecomDepartmentRepository(db),
		contact.NewWecomMemberRepository(db),
		contact.NewWecomMemberDepartmentRepository(db),
	)
	groupSyncer := group.NewGroupSyncer(group.NewWecomGroupRepository(db), group.NewWecomGroupMemberRepository(db))
	customerSyncer := customer.NewCustomerSyncer(customer.NewWecomCustomerRepository(db), customer.NewWecomCustomerFollowRepository(db))
	return wecomsync.NewService(configRepo, clientMgr, contactSyncer, groupSyncer, customerSyncer)
}
