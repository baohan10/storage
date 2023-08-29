package coordinator

import (
	"gitlink.org.cn/cloudream/common/pkgs/mq"
	"gitlink.org.cn/cloudream/storage-common/pkgs/db/model"
)

type BucketService interface {
	GetUserBuckets(msg *GetUserBuckets) (*GetUserBucketsResp, *mq.CodeMessage)

	GetBucketPackages(msg *GetBucketPackages) (*GetBucketPackagesResp, *mq.CodeMessage)

	CreateBucket(msg *CreateBucket) (*CreateBucketResp, *mq.CodeMessage)

	DeleteBucket(msg *DeleteBucket) (*DeleteBucketResp, *mq.CodeMessage)
}

// 获取用户所有的桶
var _ = Register(BucketService.GetUserBuckets)

type GetUserBuckets struct {
	UserID int64 `json:"userID"`
}
type GetUserBucketsResp struct {
	Buckets []model.Bucket `json:"buckets"`
}

func NewGetUserBuckets(userID int64) GetUserBuckets {
	return GetUserBuckets{
		UserID: userID,
	}
}
func NewGetUserBucketsResp(buckets []model.Bucket) GetUserBucketsResp {
	return GetUserBucketsResp{
		Buckets: buckets,
	}
}
func (client *Client) GetUserBuckets(msg GetUserBuckets) (*GetUserBucketsResp, error) {
	return mq.Request[GetUserBucketsResp](client.rabbitCli, msg)
}

// 获取桶中的所有Package
var _ = Register(BucketService.GetBucketPackages)

type GetBucketPackages struct {
	UserID   int64 `json:"userID"`
	BucketID int64 `json:"bucketID"`
}
type GetBucketPackagesResp struct {
	Packages []model.Package `json:"packages"`
}

func NewGetBucketPackages(userID int64, bucketID int64) GetBucketPackages {
	return GetBucketPackages{
		UserID:   userID,
		BucketID: bucketID,
	}
}
func NewGetBucketPackagesResp(packages []model.Package) GetBucketPackagesResp {
	return GetBucketPackagesResp{
		Packages: packages,
	}
}
func (client *Client) GetBucketPackages(msg GetBucketPackages) (*GetBucketPackagesResp, error) {
	return mq.Request[GetBucketPackagesResp](client.rabbitCli, msg)
}

// 创建桶
var _ = Register(BucketService.CreateBucket)

type CreateBucket struct {
	UserID     int64  `json:"userID"`
	BucketName string `json:"bucketName"`
}
type CreateBucketResp struct {
	BucketID int64 `json:"bucketID"`
}

func NewCreateBucket(userID int64, bucketName string) CreateBucket {
	return CreateBucket{
		UserID:     userID,
		BucketName: bucketName,
	}
}
func NewCreateBucketResp(bucketID int64) CreateBucketResp {
	return CreateBucketResp{
		BucketID: bucketID,
	}
}
func (client *Client) CreateBucket(msg CreateBucket) (*CreateBucketResp, error) {
	return mq.Request[CreateBucketResp](client.rabbitCli, msg)
}

// 删除桶
var _ = Register(BucketService.DeleteBucket)

type DeleteBucket struct {
	UserID   int64 `json:"userID"`
	BucketID int64 `json:"bucketID"`
}
type DeleteBucketResp struct{}

func NewDeleteBucket(userID int64, bucketID int64) DeleteBucket {
	return DeleteBucket{
		UserID:   userID,
		BucketID: bucketID,
	}
}
func NewDeleteBucketResp() DeleteBucketResp {
	return DeleteBucketResp{}
}
func (client *Client) DeleteBucket(msg DeleteBucket) (*DeleteBucketResp, error) {
	return mq.Request[DeleteBucketResp](client.rabbitCli, msg)
}