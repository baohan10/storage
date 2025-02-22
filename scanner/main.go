package main

import (
	"fmt"
	"os"

	"gitlink.org.cn/cloudream/common/pkgs/logger"
	stgglb "gitlink.org.cn/cloudream/storage/common/globals"
	"gitlink.org.cn/cloudream/storage/common/pkgs/db"
	"gitlink.org.cn/cloudream/storage/common/pkgs/distlock"
	agtrpc "gitlink.org.cn/cloudream/storage/common/pkgs/grpc/agent"
	scmq "gitlink.org.cn/cloudream/storage/common/pkgs/mq/scanner"
	"gitlink.org.cn/cloudream/storage/scanner/internal/config"
	"gitlink.org.cn/cloudream/storage/scanner/internal/event"
	"gitlink.org.cn/cloudream/storage/scanner/internal/mq"
	"gitlink.org.cn/cloudream/storage/scanner/internal/tickevent"
)

func main() {
	err := config.Init()
	if err != nil {
		fmt.Printf("init config failed, err: %s", err.Error())
		os.Exit(1)
	}

	err = logger.Init(&config.Cfg().Logger)
	if err != nil {
		fmt.Printf("init logger failed, err: %s", err.Error())
		os.Exit(1)
	}

	db, err := db.NewDB(&config.Cfg().DB)
	if err != nil {
		logger.Fatalf("new db failed, err: %s", err.Error())
	}

	stgglb.InitMQPool(&config.Cfg().RabbitMQ)

	stgglb.InitAgentRPCPool(&agtrpc.PoolConfig{})

	distlockSvc, err := distlock.NewService(&config.Cfg().DistLock)
	if err != nil {
		logger.Warnf("new distlock service failed, err: %s", err.Error())
		os.Exit(1)
	}
	go serveDistLock(distlockSvc)

	eventExecutor := event.NewExecutor(db, distlockSvc)
	go serveEventExecutor(&eventExecutor)

	agtSvr, err := scmq.NewServer(mq.NewService(&eventExecutor), &config.Cfg().RabbitMQ)
	if err != nil {
		logger.Fatalf("new agent server failed, err: %s", err.Error())
	}
	agtSvr.OnError(func(err error) {
		logger.Warnf("agent server err: %s", err.Error())
	})
	go serveScannerServer(agtSvr)

	tickExecutor := tickevent.NewExecutor(tickevent.ExecuteArgs{
		EventExecutor: &eventExecutor,
		DB:            db,
	})
	startTickEvent(&tickExecutor)

	forever := make(chan struct{})
	<-forever
}

func serveEventExecutor(executor *event.Executor) {
	logger.Info("start serving event executor")

	err := executor.Execute()

	if err != nil {
		logger.Errorf("event executor stopped with error: %s", err.Error())
	}

	logger.Info("event executor stopped")

	// TODO 仅简单结束了程序
	os.Exit(1)
}

func serveScannerServer(server *scmq.Server) {
	logger.Info("start serving scanner server")

	err := server.Serve()

	if err != nil {
		logger.Errorf("scanner server stopped with error: %s", err.Error())
	}

	logger.Info("scanner server stopped")

	// TODO 仅简单结束了程序
	os.Exit(1)
}

func serveDistLock(svc *distlock.Service) {
	logger.Info("start serving distlock")

	err := svc.Serve()

	if err != nil {
		logger.Errorf("distlock stopped with error: %s", err.Error())
	}

	logger.Info("distlock stopped")

	// TODO 仅简单结束了程序
	os.Exit(1)
}

func startTickEvent(tickExecutor *tickevent.Executor) {
	// TODO 可以考虑增加配置文件，配置这些任务间隔时间

	interval := 5 * 60 * 1000

	tickExecutor.Start(tickevent.NewBatchAllAgentCheckCache(), interval, tickevent.StartOption{RandomStartDelayMs: 60 * 1000})

	tickExecutor.Start(tickevent.NewBatchCheckAllPackage(), interval, tickevent.StartOption{RandomStartDelayMs: 60 * 1000})

	// tickExecutor.Start(tickevent.NewBatchCheckAllRepCount(), interval, tickevent.StartOption{RandomStartDelayMs: 60 * 1000})

	tickExecutor.Start(tickevent.NewBatchCheckAllStorage(), interval, tickevent.StartOption{RandomStartDelayMs: 60 * 1000})

	tickExecutor.Start(tickevent.NewCheckAgentState(), 5*60*1000, tickevent.StartOption{RandomStartDelayMs: 60 * 1000})

	tickExecutor.Start(tickevent.NewBatchCheckPackageRedundancy(), interval, tickevent.StartOption{RandomStartDelayMs: 20 * 60 * 1000})

	tickExecutor.Start(tickevent.NewBatchCleanPinned(), interval, tickevent.StartOption{RandomStartDelayMs: 20 * 60 * 1000})
}
