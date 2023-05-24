package tickevent

import (
	"github.com/samber/lo"
	"gitlink.org.cn/cloudream/common/utils/logger"
	"gitlink.org.cn/cloudream/db/model"
	mysql "gitlink.org.cn/cloudream/db/sql"
	"gitlink.org.cn/cloudream/scanner/internal/event"
)

const AGENT_CHECK_CACHE_BATCH_SIZE = 2

type BatchAgentCheckCache struct {
	nodeIDs []int
}

func (e *BatchAgentCheckCache) Execute(ctx ExecuteContext) {
	if e.nodeIDs == nil || len(e.nodeIDs) == 0 {
		nodes, err := mysql.Node.GetAllNodes(ctx.Args.DB.SQLCtx())
		if err != nil {
			logger.Warnf("get all nodes failed, err: %s", err.Error())
			return
		}

		e.nodeIDs = lo.Map(nodes, func(node model.Node, index int) int { return node.NodeID })

		logger.Debugf("new check start, get all nodes")
	}

	for i := 0; i < len(e.nodeIDs) && i < AGENT_CHECK_CACHE_BATCH_SIZE; i++ {
		// nil代表进行全量检查
		ctx.Args.EventExecutor.Post(event.NewAgentCheckCache(e.nodeIDs[i], nil))
	}
	e.nodeIDs = e.nodeIDs[AGENT_CHECK_CACHE_BATCH_SIZE:]
}
