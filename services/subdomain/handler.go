package subdomain

import (
	"context"
	"errors"
	"fmt"
	"gitlab.example.com/zhangweijie/subdomain/global"
	"gitlab.example.com/zhangweijie/subdomain/middlerware/schemas"
	"gitlab.example.com/zhangweijie/subdomain/services/common"
	"gitlab.example.com/zhangweijie/subdomain/services/result"
	toolGlobal "gitlab.example.com/zhangweijie/tool-sdk/global"
	toolSchemas "gitlab.example.com/zhangweijie/tool-sdk/middleware/schemas"
	toolModels "gitlab.example.com/zhangweijie/tool-sdk/models"
	"strconv"
	"sync"
)

type Worker struct {
	ID         int // 任务执行者 ID
	Ctx        context.Context
	Wg         *sync.WaitGroup
	TaskChan   chan Task                // 子任务通道
	ResultChan chan result.WorkerResult // 子任务结果通道
}

type Task struct {
	WorkUUID                string   // 总任务 UUID
	TaskUUID                string   // 子任务 UUID
	TargetDomains           []string // 子任务目标
	TargetSubdomainSuffixes []string // 子任务目标
}

// NewWorker 初始化 worker
func NewWorker(ctx context.Context, wg *sync.WaitGroup, id int, taskChan chan Task, resultChan chan result.WorkerResult) *Worker {
	return &Worker{
		ID:         id,
		Ctx:        ctx,
		Wg:         wg,
		TaskChan:   taskChan,
		ResultChan: resultChan,
	}
}

// calculateChanCap 计算任务通道和结果通道的容量
func calculateChanCap(ipLength, portLength int) int {
	domainLen := (ipLength + global.DefaultDomainGroupCount) / global.DefaultDomainGroupCount
	subdomainSuffixLen := (portLength + global.DefaultSuffixGroupCount) / global.DefaultSuffixGroupCount

	return domainLen * subdomainSuffixLen
}

// GroupSubdomainWorker 真正的执行 worker, 端口扫描只需要部分属性数据，不需要获取全部信息
func (w *Worker) GroupSubdomainWorker() {
	go func() {
		defer w.Wg.Done()

		for task := range w.TaskChan {
			select {
			case <-w.Ctx.Done():
				return
			default:
				fmt.Println("------------>", task)
				workerResult := result.NewWorkerResult()
				// 向结果通道推送数据
				w.ResultChan <- *workerResult
			}
		}
	}()
}

// SubdomainMainWorker  域名爆破主程序
func SubdomainMainWorker(ctx context.Context, work *toolModels.Work, validParams *schemas.DomainParams) error {
	quit := make(chan struct{})
	errChan := make(chan error, 2)

	go func() {
		defer close(quit)
		defer close(errChan)
		validSuffixDict := common.SubdomainSuffixDict
		// 自适应计算通道的容量
		maxBufferSize := calculateChanCap(len(validParams.Domains), len(*validSuffixDict))
		onePercent := float64(100 / maxBufferSize)
		taskChan := make(chan Task, maxBufferSize)
		resultChan := make(chan result.WorkerResult, maxBufferSize)
		var wg sync.WaitGroup
		// 创建并启动多个工作者
		for i := 0; i < toolGlobal.Config.Server.Worker; i++ {
			worker := NewWorker(ctx, &wg, i, taskChan, resultChan)
			worker.GroupSubdomainWorker()
			wg.Add(1)
		}

		// 生产者向任务通道发送任务
		go func() {
			// 通知消费者所有任务已经推送完毕
			defer close(taskChan)
			count := 0
			// 超出限制端口数量，则拆分端口进行异步任务操作
			for suffixStart := 0; suffixStart < len(*validSuffixDict); suffixStart += global.DefaultSuffixGroupCount {
				suffixEnd := suffixStart + global.DefaultSuffixGroupCount
				if suffixEnd > len(*validSuffixDict) {
					suffixEnd = len(*validSuffixDict)
				}
				// 超出限制的主域名数量，则拆分主域名进行异步任务操作
				for domainStart := 0; domainStart < len(validParams.Domains); domainStart += global.DefaultDomainGroupCount {
					domainEnd := domainStart + global.DefaultDomainGroupCount
					if domainEnd > len(validParams.Domains) {
						domainEnd = len(validParams.Domains)
					}
					task := Task{
						WorkUUID:                work.UUID,
						TaskUUID:                strconv.Itoa(count),
						TargetDomains:           validParams.Domains[domainStart:domainEnd],
						TargetSubdomainSuffixes: (*validSuffixDict)[suffixStart:suffixEnd],
					}
					taskChan <- task
					count += 1
				}
			}
		}()
		// 等待所有工作者完成任务
		go func() {
			wg.Wait()
			close(resultChan)
		}()

		// 中间需要进行数据结构转换
		//tmpResult := result.NewWorkerResult()
		// 回收结果
		for workerResult := range resultChan {
			fmt.Println("------------>", workerResult)
			if work.ProgressType != "" && work.ProgressUrl != "" {
				pushProgress := &toolGlobal.Progress{WorkUUID: work.UUID, ProgressType: work.ProgressType, ProgressUrl: work.ProgressUrl, Progress: 0}
				pushProgress.Progress += onePercent
				// 回传进度
				toolGlobal.ValidProgressChan <- pushProgress
			}
		}

		baseFinalResult := make(map[string]result.WorkerResult)

		var finalResult []result.WorkerResult
		for _, domain := range validParams.Domains {
			workerResult := result.WorkerResult{Domain: domain}
			baseFinalResult[domain] = workerResult
		}

		for _, baseResult := range baseFinalResult {
			finalResult = append(finalResult, baseResult)
		}

		if work.CallbackType != "" && work.CallbackUrl != "" {
			pushResult := &toolGlobal.Result{WorkUUID: work.UUID, CallbackType: work.CallbackType, CallbackUrl: work.CallbackUrl, Result: map[string]interface{}{"result": finalResult}}
			// 回传结果
			toolGlobal.ValidResultChan <- pushResult
		}
	}()

	select {
	case <-ctx.Done():
		return errors.New(toolSchemas.WorkCancelErr)
	case <-quit:
		return nil
	case err := <-errChan:
		return err
	}
}
