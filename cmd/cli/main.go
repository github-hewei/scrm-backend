package main

import (
	"os"

	"zero-backend/internal/cli"
	"zero-backend/internal/cli/runner"
	"zero-backend/internal/config"
	"zero-backend/internal/modules/rbac"
	"zero-backend/internal/provider"

	"github.com/241x/zero-kit/gormutil"
	"github.com/241x/zero-kit/mongodb"
	"github.com/241x/zero-kit/mysql"
	"github.com/241x/zero-kit/queue"
	"github.com/241x/zero-kit/redis"
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
	app.AddCommand(cli.WecomSyncCmd(log, provider.NewWecomSyncService(db, rdb)))
	os.Exit(app.Run())
}
