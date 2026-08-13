package main

import (
	"zero-backend/internal/cli"
	"zero-backend/internal/cli/runner"
	"zero-backend/internal/config"
	"zero-backend/internal/modules/rbac"
	wecomconfig "zero-backend/internal/modules/wecom/config"
	"zero-backend/internal/modules/wecom/group"
	wecomsync "zero-backend/internal/modules/wecom/sync"
	"zero-backend/internal/provider"

	"github.com/241x/zero-kit/gormutil"
	"github.com/241x/zero-kit/mongodb"
	"github.com/241x/zero-kit/mysql"
	"github.com/241x/zero-kit/queue"
	"github.com/241x/zero-kit/redis"
	"github.com/241x/zero-third/wecom"
	goredis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func main() {
	config.Init()

	conn := mongodb.MustNewConn(provider.LoadMongoConfig())
	log := provider.NewLogger(conn.DB, "cli.log")

	rdb := redis.New(provider.LoadRedisConfig())

	gormLog := gormutil.NewLogger(log)
	db := mysql.MustNewDB(provider.LoadMySQLConfig(), gormLog)

	app := cli.New(log, rdb)
	app.AddCommand(cli.MigrateCmd(db, log))
	app.AddCommand(cli.QueueCmd(queue.NewQueueManager(rdb)))
	app.AddCommand(cli.SyncApiCmd(runner.NewSyncApiRunner(log, rbac.NewRbacApiRepository(db))))
	app.AddCommand(cli.WecomSyncCmd(log, buildWecomSyncService(db, rdb)))
	app.Run()
}

// buildWecomSyncService 组装企微同步服务
func buildWecomSyncService(db *gorm.DB, rdb *goredis.Client) runner.WecomSyncService {
	configRepo := wecomconfig.NewWecomConfigRepository(db)
	appRepo := wecomconfig.NewWecomAppRepository(db)
	clientMgr := wecomsync.NewClientManager(configRepo, appRepo, wecom.NewRedisCache(rdb))
	groupSyncer := group.NewGroupSyncer(group.NewWecomGroupRepository(db), group.NewWecomGroupMemberRepository(db))
	return wecomsync.NewService(configRepo, clientMgr, groupSyncer)
}
