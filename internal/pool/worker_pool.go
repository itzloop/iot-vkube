package pool

import (
	"context"
	"errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"sync/atomic"
	"time"
)

type Stats struct {
	TotalWorkers        int
	TotalJobs           int64
	TotalFailedJobs     int64
	JobsPerWorker       map[int]int64
	FailedJobsPerWorker map[int]int64
}

type Job interface {
	Run(ctx context.Context) error
}

type jobWrapper struct {
	currentBatchCount *atomic.Int64
	Job               Job
	closed            *atomic.Bool
	errChan           chan error
}

type WorkerPool struct {
	ctx    context.Context
	cancel context.CancelFunc
	jobs   chan jobWrapper
	errGrp *errgroup.Group

	started *atomic.Bool
	closed  *atomic.Bool

	totalJobs           *atomic.Int64
	totalFailedJobs     *atomic.Int64
	jobsPerWorker       map[int]*atomic.Int64
	failedJobsPerWorker map[int]*atomic.Int64

	workerSize     int
	jobsBufferSize int
}

func NewWorkerPool(workers, jobsBufferSize int) *WorkerPool {
	wp := &WorkerPool{
		jobs:                make(chan jobWrapper, jobsBufferSize),
		workerSize:          workers,
		jobsBufferSize:      jobsBufferSize,
		started:             &atomic.Bool{},
		closed:              &atomic.Bool{},
		totalJobs:           &atomic.Int64{},
		totalFailedJobs:     &atomic.Int64{},
		jobsPerWorker:       map[int]*atomic.Int64{},
		failedJobsPerWorker: map[int]*atomic.Int64{},
	}

	for i := 0; i < workers; i++ {
		wp.jobsPerWorker[i] = &atomic.Int64{}
		wp.failedJobsPerWorker[i] = &atomic.Int64{}
	}

	return wp
}

func (wp *WorkerPool) GetStats() Stats {
	jobsPerWorker := map[int]int64{}
	for k, v := range wp.jobsPerWorker {
		jobsPerWorker[k] = v.Load()
	}

	failedJobsPerWorker := map[int]int64{}
	for k, v := range wp.failedJobsPerWorker {
		failedJobsPerWorker[k] = v.Load()
	}

	return Stats{
		TotalWorkers:        wp.workerSize,
		TotalJobs:           wp.totalJobs.Load(),
		TotalFailedJobs:     wp.totalFailedJobs.Load(),
		JobsPerWorker:       jobsPerWorker,
		FailedJobsPerWorker: failedJobsPerWorker,
	}
}

func (wp *WorkerPool) Start(ctx context.Context) {
	if wp.started.Swap(true) {
		return
	}

	ctx, cancel := context.WithCancel(ctx)
	wp.ctx = ctx
	wp.cancel = cancel

	group, ctx := errgroup.WithContext(ctx)
	wp.errGrp = group

	for i := 0; i < wp.workerSize; i++ {
		i := i
		wp.errGrp.Go(func() error {
			return wp.worker(i)
		})
	}
}

func (wp *WorkerPool) Close() error {
	if wp.started.Swap(true) {
		return nil
	}

	wp.cancel()
	close(wp.jobs)

	if err := wp.errGrp.Wait(); err != nil {
		return err
	}
	return nil
}

func (wp *WorkerPool) ExecuteBatch(jobs <-chan Job) (<-chan error, error) {
	if !wp.started.Load() {
		return nil, errors.New("called ExecuteBatch on WorkerPool that has not started yet")
	}

	if wp.closed.Load() {
		return nil, errors.New("called ExecuteBatch on a closed WorkerPool")
	}

	errChan := make(chan error)
	go func() {

		var (
			maxBatchCount     = int64(0)
			currentBatchCount = atomic.Int64{}
			closed            = atomic.Bool{}
		)
		defer close(errChan)
		defer closed.Swap(true)
		for job := range jobs {
			maxBatchCount++
			select {
			case <-wp.ctx.Done():
				if wp.ctx.Err() == context.Canceled {
					return
				}

				errChan <- wp.ctx.Err()
				return
			default:
				wp.jobs <- jobWrapper{
					currentBatchCount: &currentBatchCount,
					Job:               job,
					closed:            &closed,
					errChan:           errChan,
				}
			}
		}

		// wait for all jobs to complete
		ticker := time.NewTicker(time.Millisecond * 10)
		defer ticker.Stop()
		for {
			select {
			case <-wp.ctx.Done():
				if wp.ctx.Err() == context.Canceled {
					return
				}

				errChan <- wp.ctx.Err()
				return
			case <-ticker.C:
				if currentBatchCount.Load() == maxBatchCount {
					return
				}
			}
		}
	}()

	return errChan, nil
}

func (wp *WorkerPool) worker(id int) (err error) {
	spot := "pool/pool.go/WorkerPool.worker"
	entry := logrus.WithFields(logrus.Fields{
		"spot":      spot,
		"worker_id": id,
	})

	jobsCount := wp.jobsPerWorker[id]
	failedJobsCount := wp.failedJobsPerWorker[id]

	entry.Info("running worker")
	defer func() {
		entry.WithError(err).Info("worker done")
	}()

	for {
		select {
		case job, ok := <-wp.jobs:
			if !ok {
				return nil
			}

			err = job.Job.Run(wp.ctx)
			wp.updateStats(jobsCount, failedJobsCount, err)
			if err != nil {
				select {
				case <-wp.ctx.Done():
					job.currentBatchCount.Add(1)
					if wp.ctx.Err() == context.Canceled {
						return nil
					}

					return wp.ctx.Err()
				default:
					if job.closed.Load() {
						job.currentBatchCount.Add(1)
						return
					}

					job.errChan <- err
					job.currentBatchCount.Add(1)
				}
			} else {
				job.currentBatchCount.Add(1)
			}

		case <-wp.ctx.Done():
			if wp.ctx.Err() == context.Canceled {
				return nil
			}

			return wp.ctx.Err()
		}
	}
}

func (wp *WorkerPool) updateStats(jobsCount, failedJobsCount *atomic.Int64, err error) {
	wp.totalJobs.Add(1)
	jobsCount.Add(1)

	if err != nil {
		wp.totalFailedJobs.Add(1)
		failedJobsCount.Add(1)
	}
}
