package event

import (
	"database/sql"

	"github.com/samber/lo"
	"gitlink.org.cn/cloudream/common/utils/logger"
	"gitlink.org.cn/cloudream/db/model"
	mysql "gitlink.org.cn/cloudream/db/sql"
	"gitlink.org.cn/cloudream/scanner/internal/config"

	agtcli "gitlink.org.cn/cloudream/rabbitmq/client/agent"
	agtmsg "gitlink.org.cn/cloudream/rabbitmq/message/agent"
	agtevt "gitlink.org.cn/cloudream/rabbitmq/message/agent/event"
)

type AgentCheckCache struct {
	NodeID     int
	FileHashes []string // 需要检查的FileHash列表，如果为nil（不是为空），则代表进行全量检查
}

func NewAgentCheckCache(nodeID int, fileHashes []string) *AgentCheckCache {
	return &AgentCheckCache{
		NodeID:     nodeID,
		FileHashes: fileHashes,
	}
}

func (t *AgentCheckCache) TryMerge(other Event) bool {
	event, ok := other.(*AgentCheckCache)
	if !ok {
		return false
	}

	// FileHashes为nil时代表全量检查
	if event.FileHashes == nil {
		t.FileHashes = nil
	} else if t.FileHashes != nil {
		t.FileHashes = lo.Union(t.FileHashes, event.FileHashes)
	}

	return true
}

func (t *AgentCheckCache) Execute(execCtx ExecuteContext) {
	logger.Debugf("begin agent check cache")

	var isComplete bool
	var caches []model.Cache

	// TODO unavailable的节点需不需要发送任务？

	if t.FileHashes == nil {
		var err error
		caches, err = mysql.Cache.GetNodeCaches(execCtx.Args.DB.SQLCtx(), t.NodeID)
		if err != nil {
			logger.WithField("NodeID", t.NodeID).Warnf("get node caches failed, err: %s", err.Error())
			return
		}
		isComplete = true

	} else {
		for _, hash := range t.FileHashes {
			ch, err := mysql.Cache.Get(execCtx.Args.DB.SQLCtx(), hash, t.NodeID)
			// 记录不存在则跳过
			if err == sql.ErrNoRows {
				continue
			}

			if err != nil {
				logger.WithField("FileHash", hash).WithField("NodeID", t.NodeID).Warnf("get cache failed, err: %s", err.Error())
				return
			}

			caches = append(caches, ch)
		}
		isComplete = false
	}

	// 然后向代理端发送移动文件的请求
	agentClient, err := agtcli.NewAgentClient(t.NodeID, &config.Cfg().RabbitMQ)
	if err != nil {
		logger.WithField("NodeID", t.NodeID).Warnf("create agent client failed, err: %s", err.Error())
		return
	}
	defer agentClient.Close()

	evtmsg, err := agtmsg.NewPostEventBody(
		agtevt.NewCheckCache(isComplete, caches),
		execCtx.Option.IsEmergency, // 继承本任务的执行选项
		execCtx.Option.DontMerge)
	if err != nil {
		logger.Warnf("new post event body failed, err: %s", err.Error())
		return
	}

	err = agentClient.PostEvent(evtmsg)
	if err != nil {
		logger.WithField("NodeID", t.NodeID).Warnf("request to agent failed, err: %s", err.Error())
		return
	}
}

//func init() {
//	Register(func(msg scevt.AgentCheckCache) Event { return NewAgentCheckCache(msg.NodeID, msg.FileHashes) })
//}
