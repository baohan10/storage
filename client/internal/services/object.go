package services

import (
	"fmt"
	"io"
	"time"

	cdssdk "gitlink.org.cn/cloudream/common/sdks/storage"
	mytask "gitlink.org.cn/cloudream/storage/client/internal/task"
	stgglb "gitlink.org.cn/cloudream/storage/common/globals"
	"gitlink.org.cn/cloudream/storage/common/pkgs/db/model"
	"gitlink.org.cn/cloudream/storage/common/pkgs/iterator"
	coormq "gitlink.org.cn/cloudream/storage/common/pkgs/mq/coordinator"
)

type ObjectService struct {
	*Service
}

func (svc *Service) ObjectSvc() *ObjectService {
	return &ObjectService{Service: svc}
}

func (svc *ObjectService) StartUploading(userID cdssdk.UserID, packageID cdssdk.PackageID, objIter iterator.UploadingObjectIterator, nodeAffinity *cdssdk.NodeID) (string, error) {
	tsk := svc.TaskMgr.StartNew(mytask.NewUploadObjects(userID, packageID, objIter, nodeAffinity))
	return tsk.ID(), nil
}

func (svc *ObjectService) WaitUploading(taskID string, waitTimeout time.Duration) (bool, *mytask.UploadObjectsResult, error) {
	tsk := svc.TaskMgr.FindByID(taskID)
	if tsk.WaitTimeout(waitTimeout) {
		updatePkgTask := tsk.Body().(*mytask.UploadObjects)
		return true, updatePkgTask.Result, tsk.Error()
	}
	return false, nil, nil
}

func (svc *ObjectService) Download(userID cdssdk.UserID, objectID cdssdk.ObjectID) (io.ReadCloser, error) {
	panic("not implement yet!")
}

func (svc *ObjectService) GetPackageObjects(userID cdssdk.UserID, packageID cdssdk.PackageID) ([]model.Object, error) {
	coorCli, err := stgglb.CoordinatorMQPool.Acquire()
	if err != nil {
		return nil, fmt.Errorf("new coordinator client: %w", err)
	}
	defer stgglb.CoordinatorMQPool.Release(coorCli)

	getResp, err := coorCli.GetPackageObjects(coormq.NewGetPackageObjects(userID, packageID))
	if err != nil {
		return nil, fmt.Errorf("requsting to coodinator: %w", err)
	}

	return getResp.Objects, nil
}
