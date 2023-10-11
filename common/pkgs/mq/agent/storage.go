package agent

import (
	"gitlink.org.cn/cloudream/common/pkgs/mq"
	stgsdk "gitlink.org.cn/cloudream/common/sdks/storage"

	"gitlink.org.cn/cloudream/storage/common/pkgs/db/model"
)

type StorageService interface {
	StartStorageLoadPackage(msg *StartStorageLoadPackage) (*StartStorageLoadPackageResp, *mq.CodeMessage)

	WaitStorageLoadPackage(msg *WaitStorageLoadPackage) (*WaitStorageLoadPackageResp, *mq.CodeMessage)

	StorageCheck(msg *StorageCheck) (*StorageCheckResp, *mq.CodeMessage)

	StartStorageCreatePackage(msg *StartStorageCreatePackage) (*StartStorageCreatePackageResp, *mq.CodeMessage)

	WaitStorageCreatePackage(msg *WaitStorageCreatePackage) (*WaitStorageCreatePackageResp, *mq.CodeMessage)
}

// 启动调度Package的任务
var _ = Register(Service.StartStorageLoadPackage)

type StartStorageLoadPackage struct {
	mq.MessageBodyBase
	UserID    int64 `json:"userID"`
	PackageID int64 `json:"packageID"`
	StorageID int64 `json:"storageID"`
}
type StartStorageLoadPackageResp struct {
	mq.MessageBodyBase
	TaskID string `json:"taskID"`
}

func NewStartStorageLoadPackage(userID int64, packageID int64, storageID int64) *StartStorageLoadPackage {
	return &StartStorageLoadPackage{
		UserID:    userID,
		PackageID: packageID,
		StorageID: storageID,
	}
}
func NewStartStorageLoadPackageResp(taskID string) *StartStorageLoadPackageResp {
	return &StartStorageLoadPackageResp{
		TaskID: taskID,
	}
}
func (client *Client) StartStorageLoadPackage(msg *StartStorageLoadPackage, opts ...mq.RequestOption) (*StartStorageLoadPackageResp, error) {
	return mq.Request(Service.StartStorageLoadPackage, client.rabbitCli, msg, opts...)
}

// 等待调度Package的任务
var _ = Register(Service.WaitStorageLoadPackage)

type WaitStorageLoadPackage struct {
	mq.MessageBodyBase
	TaskID        string `json:"taskID"`
	WaitTimeoutMs int64  `json:"waitTimeout"`
}
type WaitStorageLoadPackageResp struct {
	mq.MessageBodyBase
	IsComplete bool   `json:"isComplete"`
	Error      string `json:"error"`
	FullPath   string `json:"fullPath"`
}

func NewWaitStorageLoadPackage(taskID string, waitTimeoutMs int64) *WaitStorageLoadPackage {
	return &WaitStorageLoadPackage{
		TaskID:        taskID,
		WaitTimeoutMs: waitTimeoutMs,
	}
}
func NewWaitStorageLoadPackageResp(isComplete bool, err string, fullPath string) *WaitStorageLoadPackageResp {
	return &WaitStorageLoadPackageResp{
		IsComplete: isComplete,
		Error:      err,
		FullPath:   fullPath,
	}
}
func (client *Client) WaitStorageLoadPackage(msg *WaitStorageLoadPackage, opts ...mq.RequestOption) (*WaitStorageLoadPackageResp, error) {
	return mq.Request(Service.WaitStorageLoadPackage, client.rabbitCli, msg, opts...)
}

// 检查Storage
var _ = Register(Service.StorageCheck)

const (
	CHECK_STORAGE_RESP_OP_DELETE     = "Delete"
	CHECK_STORAGE_RESP_OP_SET_NORMAL = "SetNormal"
)

type StorageCheck struct {
	mq.MessageBodyBase
	StorageID  int64                  `json:"storageID"`
	Directory  string                 `json:"directory"`
	IsComplete bool                   `json:"isComplete"`
	Packages   []model.StoragePackage `json:"packages"`
}
type StorageCheckResp struct {
	mq.MessageBodyBase
	DirectoryState string                  `json:"directoryState"`
	Entries        []StorageCheckRespEntry `json:"entries"`
}
type StorageCheckRespEntry struct {
	PackageID int64  `json:"packageID"`
	UserID    int64  `json:"userID"`
	Operation string `json:"operation"`
}

func NewStorageCheck(storageID int64, directory string, isComplete bool, packages []model.StoragePackage) *StorageCheck {
	return &StorageCheck{
		StorageID:  storageID,
		Directory:  directory,
		IsComplete: isComplete,
		Packages:   packages,
	}
}
func NewStorageCheckResp(dirState string, entries []StorageCheckRespEntry) *StorageCheckResp {
	return &StorageCheckResp{
		DirectoryState: dirState,
		Entries:        entries,
	}
}
func NewStorageCheckRespEntry(packageID int64, userID int64, op string) StorageCheckRespEntry {
	return StorageCheckRespEntry{
		PackageID: packageID,
		UserID:    userID,
		Operation: op,
	}
}
func (client *Client) StorageCheck(msg *StorageCheck, opts ...mq.RequestOption) (*StorageCheckResp, error) {
	return mq.Request(Service.StorageCheck, client.rabbitCli, msg, opts...)
}

// 启动从Storage上传Package的任务
var _ = Register(Service.StartStorageCreatePackage)

type StartStorageCreatePackage struct {
	mq.MessageBodyBase
	UserID       int64                      `json:"userID"`
	BucketID     int64                      `json:"bucketID"`
	Name         string                     `json:"name"`
	StorageID    int64                      `json:"storageID"`
	Path         string                     `json:"path"`
	Redundancy   stgsdk.TypedRedundancyInfo `json:"redundancy"`
	NodeAffinity *int64                     `json:"nodeAffinity"`
}
type StartStorageCreatePackageResp struct {
	mq.MessageBodyBase
	TaskID string `json:"taskID"`
}

func NewStartStorageCreatePackage(userID int64, bucketID int64, name string, storageID int64, path string, redundancy stgsdk.TypedRedundancyInfo, nodeAffinity *int64) *StartStorageCreatePackage {
	return &StartStorageCreatePackage{
		UserID:       userID,
		BucketID:     bucketID,
		Name:         name,
		StorageID:    storageID,
		Path:         path,
		Redundancy:   redundancy,
		NodeAffinity: nodeAffinity,
	}
}
func NewStartStorageCreatePackageResp(taskID string) *StartStorageCreatePackageResp {
	return &StartStorageCreatePackageResp{
		TaskID: taskID,
	}
}
func (client *Client) StartStorageCreatePackage(msg *StartStorageCreatePackage, opts ...mq.RequestOption) (*StartStorageCreatePackageResp, error) {
	return mq.Request(Service.StartStorageCreatePackage, client.rabbitCli, msg, opts...)
}

// 等待从Storage上传Package的任务
var _ = Register(Service.WaitStorageCreatePackage)

type WaitStorageCreatePackage struct {
	mq.MessageBodyBase
	TaskID        string `json:"taskID"`
	WaitTimeoutMs int64  `json:"waitTimeout"`
}
type WaitStorageCreatePackageResp struct {
	mq.MessageBodyBase
	IsComplete bool   `json:"isComplete"`
	Error      string `json:"error"`
	PackageID  int64  `json:"packageID"`
}

func NewWaitStorageCreatePackage(taskID string, waitTimeoutMs int64) *WaitStorageCreatePackage {
	return &WaitStorageCreatePackage{
		TaskID:        taskID,
		WaitTimeoutMs: waitTimeoutMs,
	}
}
func NewWaitStorageCreatePackageResp(isComplete bool, err string, packageID int64) *WaitStorageCreatePackageResp {
	return &WaitStorageCreatePackageResp{
		IsComplete: isComplete,
		Error:      err,
		PackageID:  packageID,
	}
}
func (client *Client) WaitStorageCreatePackage(msg *WaitStorageCreatePackage, opts ...mq.RequestOption) (*WaitStorageCreatePackageResp, error) {
	return mq.Request(Service.WaitStorageCreatePackage, client.rabbitCli, msg, opts...)
}
