package event

import (
	"database/sql"

	"github.com/samber/lo"
	"gitlink.org.cn/cloudream/common/consts"
	"gitlink.org.cn/cloudream/common/utils/logger"
	"gitlink.org.cn/cloudream/db/model"
	mysql "gitlink.org.cn/cloudream/db/sql"
	agtcli "gitlink.org.cn/cloudream/rabbitmq/client/agent"
	agtmsg "gitlink.org.cn/cloudream/rabbitmq/message/agent"
	agtevt "gitlink.org.cn/cloudream/rabbitmq/message/agent/event"
	scevt "gitlink.org.cn/cloudream/rabbitmq/message/scanner/event"
	"gitlink.org.cn/cloudream/scanner/internal/config"
)

type AgentCheckStorage struct {
	scevt.AgentCheckStorage
}

func NewAgentCheckStorage(storageID int, objectIDs []int) *AgentCheckStorage {
	return &AgentCheckStorage{
		AgentCheckStorage: scevt.NewAgentCheckStorage(storageID, objectIDs),
	}
}

func (t *AgentCheckStorage) TryMerge(other Event) bool {
	event, ok := other.(*AgentCheckStorage)
	if !ok {
		return false
	}

	if t.StorageID != event.StorageID {
		return false
	}

	// ObjectIDs为nil时代表全量检查
	if event.ObjectIDs == nil {
		t.ObjectIDs = nil
	} else if t.ObjectIDs != nil {
		t.ObjectIDs = lo.Union(t.ObjectIDs, event.ObjectIDs)
	}

	return true
}

func (t *AgentCheckStorage) Execute(execCtx ExecuteContext) {
	logger.Debugf("begin agent check storage")

	stg, err := mysql.Storage.GetByID(execCtx.Args.DB.SQLCtx(), t.StorageID)
	if err != nil {
		if err != sql.ErrNoRows {
			logger.WithField("StorageID", t.StorageID).Warnf("get storage failed, err: %s", err.Error())
		}
		return
	}

	node, err := mysql.Node.GetByID(execCtx.Args.DB.SQLCtx(), stg.NodeID)
	if err != nil {
		if err != sql.ErrNoRows {
			logger.WithField("StorageID", t.StorageID).Warnf("get storage node failed, err: %s", err.Error())
		}
		return
	}

	// TODO unavailable的节点需不需要发送任务？
	if node.State != consts.NODE_STATE_NORMAL {
		return
	}

	// 获取对象信息
	var isComplete bool
	var objects []model.StorageObject
	if t.ObjectIDs == nil {
		var err error
		objects, err = mysql.StorageObject.GetAllByStorageID(execCtx.Args.DB.SQLCtx(), t.StorageID)
		if err != nil {
			logger.WithField("StorageID", t.StorageID).Warnf("get storage objects failed, err: %s", err.Error())
			return
		}
		isComplete = true
	} else {
		for _, objID := range t.ObjectIDs {
			objs, err := mysql.StorageObject.GetAllByStorageAndObjectID(execCtx.Args.DB.SQLCtx(), t.StorageID, objID)
			if err != nil {
				logger.WithField("StorageID", t.StorageID).
					WithField("ObjectID", objID).
					Warnf("get storage object failed, err: %s", err.Error())
				return
			}

			objects = append(objects, objs...)
		}
		isComplete = false
	}

	// 投递任务
	agentClient, err := agtcli.NewAgentClient(stg.NodeID, &config.Cfg().RabbitMQ)
	if err != nil {
		logger.WithField("NodeID", stg.NodeID).Warnf("create agent client failed, err: %s", err.Error())
		return
	}
	defer agentClient.Close()

	evtmsg, err := agtmsg.NewPostEventBody(
		agtevt.NewCheckStorage(stg.Directory, isComplete, objects),
		execCtx.Option.IsEmergency, // 继承本任务的执行选项
		execCtx.Option.DontMerge)
	if err != nil {
		logger.Warnf("new post event body failed, err: %s", err.Error())
		return
	}

	err = agentClient.PostEvent(evtmsg)
	if err != nil {
		logger.WithField("NodeID", stg.NodeID).Warnf("request to agent failed, err: %s", err.Error())
	}
}

func init() {
	RegisterMessageConvertor(func(msg scevt.AgentCheckStorage) Event { return NewAgentCheckStorage(msg.StorageID, msg.ObjectIDs) })
}
