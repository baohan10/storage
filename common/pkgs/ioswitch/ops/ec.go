package ops

import (
	"fmt"
	"io"
	"sync"

	myio "gitlink.org.cn/cloudream/common/utils/io"
	stgmod "gitlink.org.cn/cloudream/storage/common/models"
	"gitlink.org.cn/cloudream/storage/common/pkgs/ec"
	"gitlink.org.cn/cloudream/storage/common/pkgs/ioswitch"
)

type ECCompute struct {
	EC                 stgmod.EC           `json:"ec"`
	InputIDs           []ioswitch.StreamID `json:"inputIDs"`
	OutputIDs          []ioswitch.StreamID `json:"outputIDs"`
	InputBlockIndexes  []int               `json:"inputBlockIndexes"`
	OutputBlockIndexes []int               `json:"outputBlockIndexes"`
}

func (o *ECCompute) Execute(sw *ioswitch.Switch, planID ioswitch.PlanID) error {
	rs, err := ec.NewRs(o.EC.K, o.EC.N, o.EC.ChunkSize)
	if err != nil {
		return fmt.Errorf("new ec: %w", err)
	}

	strs, err := sw.WaitStreams(planID, o.InputIDs...)
	if err != nil {
		return err
	}
	defer func() {
		for _, s := range strs {
			s.Stream.Close()
		}
	}()

	var inputs []io.Reader
	for _, s := range strs {
		inputs = append(inputs, s.Stream)
	}

	outputs := rs.ReconstructSome(inputs, o.InputBlockIndexes, o.OutputBlockIndexes)

	wg := sync.WaitGroup{}
	for i, id := range o.OutputIDs {
		wg.Add(1)
		sw.StreamReady(planID, ioswitch.NewStream(id, myio.AfterReadClosedOnce(outputs[i], func(closer io.ReadCloser) {
			wg.Done()
		})))
	}
	wg.Wait()

	return nil
}

type ECReconstruct struct {
	EC                stgmod.EC           `json:"ec"`
	InputIDs          []ioswitch.StreamID `json:"inputIDs"`
	OutputIDs         []ioswitch.StreamID `json:"outputIDs"`
	InputBlockIndexes []int               `json:"inputBlockIndexes"`
}

func (o *ECReconstruct) Execute(sw *ioswitch.Switch, planID ioswitch.PlanID) error {
	rs, err := ec.NewRs(o.EC.K, o.EC.N, o.EC.ChunkSize)
	if err != nil {
		return fmt.Errorf("new ec: %w", err)
	}

	strs, err := sw.WaitStreams(planID, o.InputIDs...)
	if err != nil {
		return err
	}
	defer func() {
		for _, s := range strs {
			s.Stream.Close()
		}
	}()

	var inputs []io.Reader
	for _, s := range strs {
		inputs = append(inputs, s.Stream)
	}

	outputs := rs.ReconstructData(inputs, o.InputBlockIndexes)

	wg := sync.WaitGroup{}
	for i, id := range o.OutputIDs {
		wg.Add(1)
		sw.StreamReady(planID, ioswitch.NewStream(id, myio.AfterReadClosedOnce(outputs[i], func(closer io.ReadCloser) {
			wg.Done()
		})))
	}
	wg.Wait()

	return nil
}

func init() {
	OpUnion.AddT((*ECCompute)(nil))
	OpUnion.AddT((*ECReconstruct)(nil))
}
