package http

import (
	"github.com/gin-gonic/gin"
	"gitlink.org.cn/cloudream/common/pkgs/logger"
	cdssdk "gitlink.org.cn/cloudream/common/sdks/storage"
	"gitlink.org.cn/cloudream/storage/client/internal/services"
)

type Server struct {
	engine     *gin.Engine
	listenAddr string
	svc        *services.Service
}

func NewServer(listenAddr string, svc *services.Service) (*Server, error) {
	engine := gin.New()

	return &Server{
		engine:     engine,
		listenAddr: listenAddr,
		svc:        svc,
	}, nil
}

func (s *Server) Serve() error {
	s.initRouters()

	logger.Infof("start serving http at: %s", s.listenAddr)
	err := s.engine.Run(s.listenAddr)

	if err != nil {
		logger.Infof("http stopped with error: %s", err.Error())
		return err
	}

	logger.Infof("http stopped")
	return nil
}

func (s *Server) initRouters() {
	s.engine.GET(cdssdk.ObjectDownloadPath, s.Object().Download)
	s.engine.POST(cdssdk.ObjectUploadPath, s.Object().Upload)
	s.engine.GET(cdssdk.ObjectGetPackageObjectsPath, s.Object().GetPackageObjects)

	s.engine.GET(cdssdk.PackageGetPath, s.Package().Get)
	s.engine.POST(cdssdk.PackageCreatePath, s.Package().Create)
	s.engine.POST("/package/delete", s.Package().Delete)
	s.engine.GET("/package/getCachedNodes", s.Package().GetCachedNodes)
	s.engine.GET("/package/getLoadedNodes", s.Package().GetLoadedNodes)

	s.engine.POST("/storage/loadPackage", s.Storage().LoadPackage)
	s.engine.POST("/storage/createPackage", s.Storage().CreatePackage)
	s.engine.GET("/storage/getInfo", s.Storage().GetInfo)

	s.engine.POST(cdssdk.CacheMovePackagePath, s.Cache().MovePackage)

	s.engine.POST(cdssdk.BucketCreatePath, s.Bucket().Create)
	s.engine.POST(cdssdk.BucketDeletePath, s.Bucket().Delete)
}
