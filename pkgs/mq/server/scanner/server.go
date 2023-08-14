package scanner

import (
	"gitlink.org.cn/cloudream/common/pkg/mq"
	mymq "gitlink.org.cn/cloudream/storage-common/pkgs/mq"
	"gitlink.org.cn/cloudream/storage-common/pkgs/mq/config"
)

// Service 协调端接口
type Service interface {
	EventService
}
type Server struct {
	service   Service
	rabbitSvr mq.RabbitMQServer

	OnError func(err error)
}

func NewServer(svc Service, cfg *config.Config) (*Server, error) {
	srv := &Server{
		service: svc,
	}

	rabbitSvr, err := mq.NewRabbitMQServer(
		cfg.MakeConnectingURL(),
		mymq.SCANNER_QUEUE_NAME,
		func(msg *mq.Message) (*mq.Message, error) {
			return msgDispatcher.Handle(srv.service, msg)
		},
	)
	if err != nil {
		return nil, err
	}

	srv.rabbitSvr = *rabbitSvr

	return srv, nil
}

func (s *Server) Stop() {
	s.rabbitSvr.Close()
}

func (s *Server) Serve() error {
	return s.rabbitSvr.Serve()
}

var msgDispatcher mq.MessageDispatcher = mq.NewMessageDispatcher()

// Register 将Service中的一个接口函数作为指定类型消息的处理函数
// TODO 需要约束：Service实现了TSvc接口
func Register[TSvc any, TReq any, TResp any](svcFn func(svc TSvc, msg *TReq) (*TResp, *mq.CodeMessage)) {
	mq.AddServiceFn(&msgDispatcher, svcFn)
}

// RegisterNoReply 将Service中的一个*没有返回值的*接口函数作为指定类型消息的处理函数
// TODO 需要约束：Service实现了TSvc接口
func RegisterNoReply[TSvc any, TReq any](svcFn func(svc TSvc, msg *TReq)) {
	mq.AddNoRespServiceFn(&msgDispatcher, svcFn)
}
