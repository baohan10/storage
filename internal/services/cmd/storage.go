package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"gitlink.org.cn/cloudream/agent/internal/config"
	"gitlink.org.cn/cloudream/ec"
	"gitlink.org.cn/cloudream/utils"
	log "gitlink.org.cn/cloudream/utils/logger"

	coorcli "gitlink.org.cn/cloudream/rabbitmq/client/coordinator"
	ramsg "gitlink.org.cn/cloudream/rabbitmq/message"
	agtmsg "gitlink.org.cn/cloudream/rabbitmq/message/agent"
	coormsg "gitlink.org.cn/cloudream/rabbitmq/message/coordinator"
	"gitlink.org.cn/cloudream/utils/consts/errorcode"
)

func (service *Service) RepMove(msg *agtmsg.RepMoveCommand) *agtmsg.AgentMoveResp {
	outFileName := utils.MakeMoveOperationFileName(msg.Body.ObjectID, msg.Body.UserID)
	outFileDir := filepath.Join(config.Cfg().StorageBaseDir, msg.Body.Directory)
	outFilePath := filepath.Join(outFileDir, outFileName)

	err := os.MkdirAll(outFileDir, 0644)
	if err != nil {
		log.Warnf("create file directory %s failed, err: %s", outFileDir, err.Error())
		return ramsg.ReplyFailed[agtmsg.AgentMoveResp](errorcode.OPERATION_FAILED, "create local file directory failed")
	}

	outFile, err := os.Create(outFilePath)
	if err != nil {
		log.Warnf("create file %s failed, err: %s", outFilePath, err.Error())
		return ramsg.ReplyFailed[agtmsg.AgentMoveResp](errorcode.OPERATION_FAILED, "create local file failed")
	}
	defer outFile.Close()

	hashs := msg.Body.Hashs
	fileHash := hashs[0]
	ipfsRd, err := service.ipfs.OpenRead(fileHash)
	if err != nil {
		log.Warnf("read ipfs file %s failed, err: %s", fileHash, err.Error())
		return ramsg.ReplyFailed[agtmsg.AgentMoveResp](errorcode.OPERATION_FAILED, "read ipfs file failed")
	}
	defer ipfsRd.Close()

	_, err = io.Copy(outFile, ipfsRd)
	if err != nil {
		log.WithField("FileHash", fileHash).
			Warnf("copy ipfs file data to local file failed, err: %s", err.Error())
		return ramsg.ReplyFailed[agtmsg.AgentMoveResp](errorcode.OPERATION_FAILED, "copy ipfs file data to local file failed")
	}

	//向coor报告临时缓存hash
	coorClient, err := coorcli.NewCoordinatorClient(&config.Cfg().RabbitMQ)
	if err != nil {
		log.Warnf("new coordinator client failed, err: %s", err.Error())
		return ramsg.ReplyFailed[agtmsg.AgentMoveResp](errorcode.OPERATION_FAILED, "new coordinator client failed")
	}
	defer coorClient.Close()

	// TODO 这里更新失败残留下的文件是否要删除？
	// TODO 参考数据修复功能需求里描述的流程进行上报
	coorClient.TempCacheReport(coormsg.NewTempCacheReportBody(config.Cfg().ID, hashs))

	return ramsg.ReplyOK(agtmsg.NewAgentMoveRespBody())
}

func (service *Service) ECMove(msg *agtmsg.ECMoveCommand) *agtmsg.AgentMoveResp {
	panic("not implement yet!")
	/*
		wg := sync.WaitGroup{}
		fmt.Println("EcMove")
		fmt.Println(msg.Hashs)
		hashs := msg.Hashs
		fileSizeInBytes := msg.FileSizeInBytes
		blockIds := msg.IDs
		ecName := msg.ECName
		goalName := msg.BucketName + ":" + msg.ObjectName + ":" + strconv.Itoa(msg.UserID)
		ecPolicies := *utils.GetEcPolicy()
		ecPolicy := ecPolicies[ecName]
		ecK := ecPolicy.GetK()
		ecN := ecPolicy.GetN()
		numPacket := (fileSizeInBytes + int64(ecK)*int64(config.Cfg().GRCPPacketSize) - 1) / (int64(ecK) * int64(config.Cfg().GRCPPacketSize))

		getBufs := make([]chan []byte, ecN)
		decodeBufs := make([]chan []byte, ecK)
		for i := 0; i < ecN; i++ {
			getBufs[i] = make(chan []byte)
		}
		for i := 0; i < ecK; i++ {
			decodeBufs[i] = make(chan []byte)
		}

		wg.Add(1)

		//执行调度操作
		// TODO 这一块需要改写以适配IPFS流式读取
		for i := 0; i < len(blockIds); i++ {
			go service.get(hashs[i], getBufs[blockIds[i]], numPacket)
		}
		go decode(getBufs[:], decodeBufs[:], blockIds, ecK, numPacket)
		// TODO 写入的文件路径需要带上msg中的Directory字段，参考RepMove
		go persist(decodeBufs[:], numPacket, goalName, &wg)
		wg.Wait()

		//向coor报告临时缓存hash
		coorClient, err := racli.NewCoordinatorClient()
		if err != nil {
			// TODO 日志
			return ramsg.NewAgentMoveRespFailed(errorcode.OPERATION_FAILED, fmt.Sprintf("create coordinator client failed"))
		}
		defer coorClient.Close()
		coorClient.TempCacheReport(NodeID, hashs)

		return ramsg.NewAgentMoveRespOK()
	*/
}

func decode(inBufs []chan []byte, outBufs []chan []byte, blockSeq []int, ecK int, numPacket int64) {
	fmt.Println("decode ")
	var tmpIn [][]byte
	var zeroPkt []byte
	tmpIn = make([][]byte, len(inBufs))
	hasBlock := map[int]bool{}
	for j := 0; j < len(blockSeq); j++ {
		hasBlock[blockSeq[j]] = true
	}
	needRepair := false //检测是否传入了所有数据块
	for j := 0; j < len(outBufs); j++ {
		if blockSeq[j] != j {
			needRepair = true
		}
	}
	enc := ec.NewRsEnc(ecK, len(inBufs))
	for i := 0; int64(i) < numPacket; i++ {
		for j := 0; j < len(inBufs); j++ { //3
			if hasBlock[j] {
				tmpIn[j] = <-inBufs[j]
			} else {
				tmpIn[j] = zeroPkt
			}
		}
		fmt.Printf("%v", tmpIn)
		if needRepair {
			err := enc.Repair(tmpIn)
			print("&&&&&")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Decode Repair Error: %s", err.Error())
			}
		}
		//fmt.Printf("%v",tmpIn)

		for j := 0; j < len(outBufs); j++ { //1,2,3//示意，需要调用纠删码编解码引擎：  tmp[k] = tmp[k]+(tmpIn[w][k]*coefs[w][j])
			outBufs[j] <- tmpIn[j]
		}
	}
	fmt.Println("decode over")
	for i := 0; i < len(outBufs); i++ {
		close(outBufs[i])
	}
}

func (service *Service) get(blockHash string, getBuf chan []byte, numPacket int64) {
	/*
		data := CatIPFS(blockHash)
		for i := 0; int64(i) < numPacket; i++ {
			buf := []byte(data[i*config.Cfg().GRCPPacketSize : i*config.Cfg().GRCPPacketSize+config.Cfg().GRCPPacketSize])
			getBuf <- buf
		}
		close(getBuf)
	*/
}

func persist(inBuf []chan []byte, numPacket int64, localFilePath string, wg *sync.WaitGroup) {
	//这里的localFilePath应该是要写入的filename
	fDir, err := os.Executable()
	if err != nil {
		panic(err)
	}
	fURL := filepath.Join(filepath.Dir(fDir), "assets3")
	_, err = os.Stat(fURL)
	if os.IsNotExist(err) {
		os.MkdirAll(fURL, os.ModePerm)
	}

	file, err := os.Create(filepath.Join(fURL, localFilePath))
	if err != nil {
		return
	}

	for i := 0; int64(i) < numPacket; i++ {
		for j := 0; j < len(inBuf); j++ {
			tmp := <-inBuf[j]
			fmt.Println(tmp)
			file.Write(tmp)
		}
	}
	file.Close()
	wg.Done()
}
