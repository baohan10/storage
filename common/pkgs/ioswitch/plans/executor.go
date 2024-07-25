package plans

import (
	"context"
	"fmt"
	"io"
	"sync"

	"gitlink.org.cn/cloudream/common/pkgs/future"
	stgglb "gitlink.org.cn/cloudream/storage/common/globals"
	"gitlink.org.cn/cloudream/storage/common/pkgs/ioswitch"
)

type Executor struct {
	planID     ioswitch.PlanID
	planBlder  *PlanBuilder
	callback   *future.SetVoidFuture
	ctx        context.Context
	cancel     context.CancelFunc
	executorSw *ioswitch.Switch
}

func (e *Executor) BeginWrite(str io.ReadCloser, target *ExecutorWriteStream) {
	target.stream.Stream = str
	e.executorSw.PutVars(target.stream)
}

func (e *Executor) BeginRead(target *ExecutorReadStream) (io.ReadCloser, error) {
	err := e.executorSw.BindVars(e.ctx, target.stream)
	if err != nil {
		return nil, fmt.Errorf("bind vars: %w", err)
	}

	return target.stream.Stream, nil
}

func (e *Executor) Signal(signal *ExecutorSignalVar) {
	e.executorSw.PutVars(signal.v)
}

func (e *Executor) Wait(ctx context.Context) (map[string]any, error) {
	err := e.callback.Wait(ctx)
	if err != nil {
		return nil, err
	}

	ret := make(map[string]any)
	e.planBlder.StoreMap.Range(func(k, v any) bool {
		ret[k.(string)] = v
		return true
	})

	return ret, nil
}

func (e *Executor) execute() {
	wg := sync.WaitGroup{}

	for _, p := range e.planBlder.AgentPlans {
		wg.Add(1)

		go func(p *AgentPlanBuilder) {
			defer wg.Done()

			plan := ioswitch.Plan{
				ID:  e.planID,
				Ops: p.Ops,
			}

			cli, err := stgglb.AgentRPCPool.Acquire(stgglb.SelectGRPCAddress(&p.Node))
			if err != nil {
				e.stopWith(fmt.Errorf("new agent rpc client of node %v: %w", p.Node.NodeID, err))
				return
			}
			defer stgglb.AgentRPCPool.Release(cli)

			err = cli.ExecuteIOPlan(e.ctx, plan)
			if err != nil {
				e.stopWith(fmt.Errorf("execute plan at %v: %w", p.Node.NodeID, err))
				return
			}
		}(p)
	}

	err := e.executorSw.Run(e.ctx)
	if err != nil {
		e.stopWith(fmt.Errorf("run executor switch: %w", err))
		return
	}

	wg.Wait()

	e.callback.SetVoid()
}

func (e *Executor) stopWith(err error) {
	e.callback.SetError(err)
	e.cancel()
}

//	type ExecutorStreamVar struct {
//		blder *PlanBuilder
//		v     *ioswitch.StreamVar
//	}
type ExecutorWriteStream struct {
	stream *ioswitch.StreamVar
}

// func (b *ExecutorPlanBuilder) WillWrite(str *ExecutorWriteStream) *ExecutorStreamVar {
// 	stream := b.blder.NewStreamVar()
// 	str.stream = stream
// 	return &ExecutorStreamVar{blder: b.blder, v: stream}
// }

// func (b *ExecutorPlanBuilder) WillSignal() *ExecutorSignalVar {
// 	s := b.blder.NewSignalVar()
// 	return &ExecutorSignalVar{blder: b.blder, v: s}
// }

type ExecutorReadStream struct {
	stream *ioswitch.StreamVar
}

// func (v *ExecutorStreamVar) WillRead(str *ExecutorReadStream) {
// 	str.stream = v.v
// }
/*
func (s *ExecutorStreamVar) To(node cdssdk.Node) *AgentStreamVar {
	s.blder.ExecutorPlan.ops = append(s.blder.ExecutorPlan.ops, &ops.SendStream{Stream: s.v, Node: node})
	return &AgentStreamVar{
		owner: s.blder.AtAgent(node),
		v:     s.v,
	}
}

type ExecutorStringVar struct {
	blder *PlanBuilder
	v     *ioswitch.StringVar
}

func (s *ExecutorStringVar) Store(key string) {
	s.blder.ExecutorPlan.ops = append(s.blder.ExecutorPlan.ops, &ops.Store{
		Var:   s.v,
		Key:   key,
		Store: s.blder.StoreMap,
	})
}

type ExecutorSignalVar struct {
	blder *PlanBuilder
	v     *ioswitch.SignalVar
}

func (s *ExecutorSignalVar) To(node cdssdk.Node) *AgentSignalVar {
	s.blder.ExecutorPlan.ops = append(s.blder.ExecutorPlan.ops, &ops.SendVar{Var: s.v, Node: node})
	return &AgentSignalVar{
		owner: s.blder.AtAgent(node),
		v:     s.v,
	}
}
*/
