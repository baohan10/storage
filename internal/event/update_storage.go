package event

import (
	tskcst "gitlink.org.cn/cloudream/common/consts/event"
	"gitlink.org.cn/cloudream/common/utils/logger"
	mysql "gitlink.org.cn/cloudream/db/sql"
	scevt "gitlink.org.cn/cloudream/rabbitmq/message/scanner/event"
)

type UpdateStorageEntry = scevt.UpdateStorageEntry
type UpdateStorage struct {
	scevt.UpdateStorage
}

func NewUpdateStorage(dirState string, entries []UpdateStorageEntry) UpdateStorage {
	return UpdateStorage{
		UpdateStorage: scevt.NewUpdateStorage(dirState, entries),
	}
}

func (t *UpdateStorage) TryMerge(other Event) bool {
	event, ok := other.(*UpdateStorage)
	if !ok {
		return false
	}
	if event.StorageID != t.StorageID {
		return false
	}

	// 后投递的任务的状态更新一些
	t.DirectoryState = event.DirectoryState
	// TODO 可以考虑合并同FileHash和NodeID的记录
	t.Entries = append(t.Entries, event.Entries...)
	return true
}

func (t *UpdateStorage) Execute(execCtx ExecuteContext) {

	err := mysql.Storage.ChangeState(execCtx.Args.DB.SQLCtx(), t.StorageID, t.DirectoryState)
	if err != nil {
		logger.WithField("StorageID", t.StorageID).Warnf("change storage state failed, err: %s", err.Error())
	}

	for _, entry := range t.Entries {
		switch entry.Operation {
		case tskcst.UPDATE_STORAGE_DELETE:
			err := mysql.StorageObject.Delete(execCtx.Args.DB.SQLCtx(), t.StorageID, entry.ObjectID, entry.UserID)
			if err != nil {
				logger.WithField("StorageID", t.StorageID).
					WithField("ObjectID", entry.ObjectID).
					Warnf("delete storage object failed, err: %s", err.Error())
			}

		case tskcst.UPDATE_STORAGE_SET_NORMAL:
			err := mysql.StorageObject.SetStateNormal(execCtx.Args.DB.SQLCtx(), t.StorageID, entry.ObjectID, entry.UserID)
			if err != nil {
				logger.WithField("StorageID", t.StorageID).
					WithField("ObjectID", entry.ObjectID).
					Warnf("change storage object state failed, err: %s", err.Error())
			}
		}
	}
}
