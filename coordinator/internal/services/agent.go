package services

import (
	coormq "gitlink.org.cn/cloudream/storage/common/pkgs/mq/coordinator"
)

func (service *Service) TempCacheReport(msg *coormq.TempCacheReport) {
	//service.db.BatchInsertOrUpdateCache(msg.Hashes, msg.NodeID)
}

func (service *Service) AgentStatusReport(msg *coormq.AgentStatusReport) {
	//jh：根据command中的Ip，插入节点延迟表，和节点表的NodeStatus
	//根据command中的Ip，插入节点延迟表

	// TODO
	/*
		ips := utils.GetAgentIps()
		Insert_NodeDelay(msg.Body.IP, ips, msg.Body.AgentDelay)

		//从配置表里读取节点地域NodeLocation
		//插入节点表的NodeStatus
		Insert_Node(msg.Body.IP, msg.Body.IP, msg.Body.IPFSStatus, msg.Body.LocalDirStatus)
	*/
}
