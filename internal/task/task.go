package task

import (
	distsvc "gitlink.org.cn/cloudream/common/pkg/distlock/service"
	"gitlink.org.cn/cloudream/common/pkg/task"
	"gitlink.org.cn/cloudream/common/utils/ipfs"
	coorcli "gitlink.org.cn/cloudream/rabbitmq/client/coordinator"
)

type TaskContext struct {
	IPFS        *ipfs.IPFS
	Coordinator *coorcli.Client
	DistLock    *distsvc.Service
}

// 需要在Task结束后主动调用，completing函数将在Manager加锁期间被调用，
// 因此适合进行执行结果的设置
type CompleteFn = task.CompleteFn

type Manager = task.Manager[TaskContext]

type TaskBody = task.TaskBody[TaskContext]

type Task = task.Task[TaskContext]

type CompleteOption = task.CompleteOption

func NewManager(ipfs *ipfs.IPFS, coorCli *coorcli.Client, distLock *distsvc.Service) Manager {
	return task.NewManager(TaskContext{
		IPFS:        ipfs,
		Coordinator: coorCli,
		DistLock:    distLock,
	})
}
