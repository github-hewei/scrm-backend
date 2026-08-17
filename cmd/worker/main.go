package main

import (
	"zero-backend/internal/config"
	"zero-backend/internal/modules/async"
	"zero-backend/internal/provider"
	"zero-backend/internal/worker"

	"github.com/241x/zero-kit/gormutil"
	"github.com/241x/zero-kit/job"
	"github.com/241x/zero-kit/mongodb"
	"github.com/241x/zero-kit/mysql"
	"github.com/241x/zero-kit/queue"
	"github.com/241x/zero-kit/redis"
)

func main() {
	config.Init()

	rdb := redis.New(provider.LoadRedisConfig())
	manager := queue.NewQueueManager(rdb)

	conn := mongodb.MustNewConn(provider.LoadMongoConfig())
	log := provider.NewLogger(conn.DB, "worker.log")

	gormLog := gormutil.NewLogger(log)
	db := mysql.MustNewDB(provider.LoadMySQLConfig(), gormLog)

	registry := worker.NewRegistry(log)
	registry.Register("example", worker.NewExampleHandler(log))

	// 作业执行器：企业微信数据同步（jobs 表由 SQLStore AutoMigrate 自动创建）
	jobStore, err := job.NewSQLStore(db)
	if err != nil {
		log.Err(err, "Failed to create job store")
		return
	}
	taskSvc := async.NewTaskService(async.NewAsyncTaskRepository(db), jobStore)
	executor := job.NewExecutor(jobStore,
		worker.NewWecomSyncJobHandler(provider.NewWecomSyncService(db, rdb, log), taskSvc, log),
		job.DefaultConfig()).
		WithLogger(log)

	// 启动服务
	worker.NewServer(manager, registry, log).AddExecutor(executor).Run()
}
