package main

import (
	"fmt"
	"os"
	"sync"

	log "gitlink.org.cn/cloudream/common/utils/logger"
	"gitlink.org.cn/cloudream/db"
	scsvr "gitlink.org.cn/cloudream/rabbitmq/server/scanner"
	"gitlink.org.cn/cloudream/scanner/internal/config"
	"gitlink.org.cn/cloudream/scanner/internal/event"
	"gitlink.org.cn/cloudream/scanner/internal/services"
	"gitlink.org.cn/cloudream/scanner/internal/tickevent"
)

func main() {
	err := config.Init()
	if err != nil {
		fmt.Printf("init config failed, err: %s", err.Error())
		os.Exit(1)
	}

	err = log.Init(&config.Cfg().Logger)
	if err != nil {
		fmt.Printf("init logger failed, err: %s", err.Error())
		os.Exit(1)
	}

	db, err := db.NewDB(&config.Cfg().DB)
	if err != nil {
		log.Fatalf("new db failed, err: %s", err.Error())
	}

	wg := sync.WaitGroup{}
	wg.Add(2)

	eventExecutor := event.NewExecutor(db)
	go serveEventExecutor(&eventExecutor, &wg)

	agtSvr, err := scsvr.NewScannerServer(services.NewService(&eventExecutor), &config.Cfg().RabbitMQ)
	if err != nil {
		log.Fatalf("new agent server failed, err: %s", err.Error())
	}
	agtSvr.OnError = func(err error) {
		log.Warnf("agent server err: %s", err.Error())
	}
	go serveScannerServer(agtSvr, &wg)

	tickExecutor := tickevent.NewExecutor(tickevent.ExecuteArgs{
		EventExecutor: &eventExecutor,
		DB:            db,
	})
	startTickEvent(&tickExecutor)

	wg.Wait()
}

func serveEventExecutor(executor *event.Executor, wg *sync.WaitGroup) {
	log.Info("start serving event executor")

	err := executor.Execute()

	if err != nil {
		log.Errorf("event executor stopped with error: %s", err.Error())
	}

	log.Info("event executor stopped")

	wg.Done()
}

func serveScannerServer(server *scsvr.ScannerServer, wg *sync.WaitGroup) {
	log.Info("start serving scanner server")

	err := server.Serve()

	if err != nil {
		log.Errorf("scanner server stopped with error: %s", err.Error())
	}

	log.Info("scanner server stopped")

	wg.Done()
}

func startTickEvent(tickExecutor *tickevent.Executor) {
	// TODO 可以考虑增加配置文件，配置这些任务间隔时间

	tickExecutor.Start(tickevent.NewBatchAllAgentCheckCache(), 5*60*100)

	tickExecutor.Start(tickevent.NewBatchCheckAllObject(), 5*60*100)

	tickExecutor.Start(tickevent.NewBatchCheckAllRepCount(), 5*60*100)

	tickExecutor.Start(tickevent.NewBatchCheckAllStorage(), 5*60*100)

	tickExecutor.Start(tickevent.NewCheckAgentState(), 5*60*100)

	tickExecutor.Start(tickevent.NewCheckUnavailableCache(), 5*60*100)
}
