package services

import (
	"fmt"
	"io"
	"time"

	cdssdk "gitlink.org.cn/cloudream/common/sdks/storage"
	mytask "gitlink.org.cn/cloudream/storage/client/internal/task"
	stgglb "gitlink.org.cn/cloudream/storage/common/globals"
	stgmod "gitlink.org.cn/cloudream/storage/common/models"
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

func (svc *ObjectService) UpdateInfo(userID cdssdk.UserID, updatings []cdssdk.UpdatingObject) error {
	coorCli, err := stgglb.CoordinatorMQPool.Acquire()
	if err != nil {
		return fmt.Errorf("new coordinator client: %w", err)
	}
	defer stgglb.CoordinatorMQPool.Release(coorCli)

	_, err = coorCli.UpdateObjectInfos(coormq.ReqUpdateObjectInfos(userID, updatings))
	if err != nil {
		return fmt.Errorf("requsting to coodinator: %w", err)
	}

	return nil
}

func (svc *ObjectService) Download(userID cdssdk.UserID, objectID cdssdk.ObjectID) (io.ReadCloser, error) {
	coorCli, err := stgglb.CoordinatorMQPool.Acquire()
	if err != nil {
		return nil, fmt.Errorf("new coordinator client: %w", err)
	}
	defer stgglb.CoordinatorMQPool.Release(coorCli)

	resp, err := coorCli.GetObjectDetails(coormq.ReqGetObjectDetails([]cdssdk.ObjectID{objectID}))
	if err != nil {
		return nil, fmt.Errorf("requesting to coordinator")
	}

	if resp.Objects[0] == nil {
		return nil, fmt.Errorf("object not found")
	}

	iter := iterator.NewDownloadObjectIterator([]stgmod.ObjectDetail{*resp.Objects[0]}, &iterator.DownloadContext{
		Distlock: svc.DistLock,
	})
	defer iter.Close()

	downloading, err := iter.MoveNext()
	if err != nil {
		return nil, err
	}

	return downloading.File, nil
}

func (svc *ObjectService) Delete(userID cdssdk.UserID, objectIDs []cdssdk.ObjectID) error {
	coorCli, err := stgglb.CoordinatorMQPool.Acquire()
	if err != nil {
		return fmt.Errorf("new coordinator client: %w", err)
	}
	defer stgglb.CoordinatorMQPool.Release(coorCli)

	_, err = coorCli.DeleteObjects(coormq.ReqDeleteObjects(userID, objectIDs))
	if err != nil {
		return fmt.Errorf("requsting to coodinator: %w", err)
	}

	return nil
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
