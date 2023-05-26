package event

import (
	"github.com/samber/lo"
	"gitlink.org.cn/cloudream/common/pkg/logger"
	scevt "gitlink.org.cn/cloudream/rabbitmq/message/scanner/event"
)

type CheckObject struct {
	scevt.CheckObject
}

func NewCheckObject(objIDs []int) *CheckObject {
	return &CheckObject{
		CheckObject: scevt.NewCheckObject(objIDs),
	}
}

func (t *CheckObject) TryMerge(other Event) bool {
	event, ok := other.(*CheckObject)
	if !ok {
		return false
	}

	t.ObjectIDs = lo.Union(t.ObjectIDs, event.ObjectIDs)
	return true
}

func (t *CheckObject) Execute(execCtx ExecuteContext) {
	log := logger.WithType[CheckObject]("Event")
	log.Debugf("begin with %v", logger.FormatStruct(t))

	for _, objID := range t.ObjectIDs {
		err := execCtx.Args.DB.Object().DeleteUnused(execCtx.Args.DB.SQLCtx(), objID)
		if err != nil {
			log.WithField("ObjectID", objID).Warnf("delete unused object failed, err: %s", err.Error())
		}
	}
}

func init() {
	RegisterMessageConvertor(func(msg scevt.CheckObject) Event { return NewCheckObject(msg.ObjectIDs) })
}
