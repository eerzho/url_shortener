package async

import (
	"context"
	"errors"
	"sync"
)

var (
	ErrJobNil      = errors.New("job cannot be nil")
	ErrPoolStopped = errors.New("worker pool not started")
	ErrQueueFull   = errors.New("job queue is full")
)

type Job func(ctx context.Context, workerId int)

type WorkerPool struct {
	started     bool
	workerCount int
	jobsChan    chan Job
	mu          sync.Mutex
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewWorkerPool(workerCount int, jobQueueSize int) *WorkerPool {
	if workerCount <= 0 {
		panic("workerCount must be positive")
	}
	if jobQueueSize <= 0 {
		panic("jobQueueSize must be positive")
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &WorkerPool{
		workerCount: workerCount,
		jobsChan:    make(chan Job, jobQueueSize),
		ctx:         ctx,
		cancel:      cancel,
	}
}

func (w *WorkerPool) Start() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.started {
		return
	}

	for id := range w.workerCount {
		w.wg.Add(1)
		go w.worker(id)
	}

	w.started = true
}

func (w *WorkerPool) Submit(job Job) error {
	if job == nil {
		return ErrJobNil
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.started {
		return ErrPoolStopped
	}

	select {
	case w.jobsChan <- job:
		return nil
	case <-w.ctx.Done():
		return w.ctx.Err()
	default:
		return ErrQueueFull
	}
}

func (w *WorkerPool) Shutdown() {
	w.cancel()

	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.started {
		return
	}

	close(w.jobsChan)
	w.wg.Wait()
	w.started = false
}

func (w *WorkerPool) worker(id int) {
	defer w.wg.Done()

	for {
		select {
		case job, ok := <-w.jobsChan:
			if !ok {
				return
			}
			job(w.ctx, id)
		case <-w.ctx.Done():
			return
		}
	}
}
