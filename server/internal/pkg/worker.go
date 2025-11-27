package pkg

import (
	"context"
	"log"
	"sync"
	"time"
)

// Job은 비동기로 처리할 작업을 나타냅니다.
type Job struct {
	ID      string
	Handler func(ctx context.Context) error
	Timeout time.Duration
}

// JobResult는 작업 처리 결과를 나타냅니다.
type JobResult struct {
	JobID string
	Error error
}

// WorkerPool은 비동기 작업을 처리하는 워커 풀입니다.
type WorkerPool struct {
	workers  int
	jobQueue chan Job
	results  chan JobResult
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	started  bool
	mu       sync.Mutex
}

// NewWorkerPool은 새 워커 풀을 생성합니다.
func NewWorkerPool(workers int, queueSize int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	return &WorkerPool{
		workers:  workers,
		jobQueue: make(chan Job, queueSize),
		results:  make(chan JobResult, queueSize),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Start는 워커 풀을 시작합니다.
func (p *WorkerPool) Start() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.started {
		return
	}

	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
	p.started = true
	log.Printf("WorkerPool started with %d workers", p.workers)
}

// worker는 개별 워커 goroutine입니다.
func (p *WorkerPool) worker(id int) {
	defer p.wg.Done()

	for {
		select {
		case <-p.ctx.Done():
			log.Printf("Worker %d shutting down", id)
			return
		case job, ok := <-p.jobQueue:
			if !ok {
				return
			}
			p.processJob(job)
		}
	}
}

// processJob은 개별 작업을 처리합니다.
func (p *WorkerPool) processJob(job Job) {
	var ctx context.Context
	var cancel context.CancelFunc

	if job.Timeout > 0 {
		ctx, cancel = context.WithTimeout(p.ctx, job.Timeout)
	} else {
		ctx, cancel = context.WithTimeout(p.ctx, 30*time.Second) // 기본 30초 타임아웃
	}
	defer cancel()

	result := JobResult{JobID: job.ID}

	// 패닉 복구
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Job %s panicked: %v", job.ID, r)
			result.Error = ErrInternalServer
		}
	}()

	result.Error = job.Handler(ctx)

	// 결과 전송 (non-blocking)
	select {
	case p.results <- result:
	default:
		// 결과 채널이 가득 찬 경우 로그만 남김
		if result.Error != nil {
			log.Printf("Job %s completed with error: %v", job.ID, result.Error)
		}
	}
}

// Submit은 작업을 큐에 추가합니다.
func (p *WorkerPool) Submit(job Job) bool {
	select {
	case p.jobQueue <- job:
		return true
	default:
		// 큐가 가득 찬 경우
		log.Printf("Job queue full, job %s rejected", job.ID)
		return false
	}
}

// SubmitFunc는 간단한 함수를 작업으로 제출합니다.
func (p *WorkerPool) SubmitFunc(id string, fn func(ctx context.Context) error) bool {
	return p.Submit(Job{
		ID:      id,
		Handler: fn,
	})
}

// SubmitFuncWithTimeout는 타임아웃이 있는 함수를 작업으로 제출합니다.
func (p *WorkerPool) SubmitFuncWithTimeout(id string, timeout time.Duration, fn func(ctx context.Context) error) bool {
	return p.Submit(Job{
		ID:      id,
		Handler: fn,
		Timeout: timeout,
	})
}

// Results는 결과 채널을 반환합니다.
func (p *WorkerPool) Results() <-chan JobResult {
	return p.results
}

// Stop은 워커 풀을 정상 종료합니다.
func (p *WorkerPool) Stop() {
	p.cancel()
	close(p.jobQueue)
	p.wg.Wait()
	close(p.results)
	log.Println("WorkerPool stopped")
}

// StopWithTimeout은 타임아웃 내에 워커 풀을 종료합니다.
func (p *WorkerPool) StopWithTimeout(timeout time.Duration) {
	p.cancel()
	close(p.jobQueue)

	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("WorkerPool stopped gracefully")
	case <-time.After(timeout):
		log.Println("WorkerPool stop timed out")
	}
	close(p.results)
}

// ========================================
// 간단한 비동기 실행 헬퍼
// ========================================

// Go는 goroutine을 실행하고 패닉을 복구합니다.
func Go(fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Goroutine panicked: %v", r)
			}
		}()
		fn()
	}()
}

// GoWithContext는 context가 있는 goroutine을 실행합니다.
func GoWithContext(ctx context.Context, fn func(ctx context.Context)) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Goroutine panicked: %v", r)
			}
		}()
		fn(ctx)
	}()
}

// RunParallel은 여러 함수를 병렬로 실행하고 모두 완료될 때까지 기다립니다.
func RunParallel(fns ...func() error) []error {
	var wg sync.WaitGroup
	errors := make([]error, len(fns))

	for i, fn := range fns {
		wg.Add(1)
		go func(idx int, f func() error) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Parallel task %d panicked: %v", idx, r)
					errors[idx] = ErrInternalServer
				}
			}()
			errors[idx] = f()
		}(i, fn)
	}

	wg.Wait()
	return errors
}

// RunParallelWithContext는 context와 함께 여러 함수를 병렬 실행합니다.
func RunParallelWithContext(ctx context.Context, fns ...func(ctx context.Context) error) []error {
	var wg sync.WaitGroup
	errors := make([]error, len(fns))

	for i, fn := range fns {
		wg.Add(1)
		go func(idx int, f func(ctx context.Context) error) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Parallel task %d panicked: %v", idx, r)
					errors[idx] = ErrInternalServer
				}
			}()

			select {
			case <-ctx.Done():
				errors[idx] = ctx.Err()
			default:
				errors[idx] = f(ctx)
			}
		}(i, fn)
	}

	wg.Wait()
	return errors
}

// ========================================
// 타임아웃/재시도 헬퍼
// ========================================

// WithTimeout은 타임아웃이 있는 함수를 실행합니다.
func WithTimeout(timeout time.Duration, fn func(ctx context.Context) error) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- fn(ctx)
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// RetryConfig는 재시도 설정입니다.
type RetryConfig struct {
	MaxRetries int
	Delay      time.Duration
	MaxDelay   time.Duration
	Multiplier float64 // 지수 백오프 승수
}

// DefaultRetryConfig는 기본 재시도 설정입니다.
var DefaultRetryConfig = RetryConfig{
	MaxRetries: 3,
	Delay:      100 * time.Millisecond,
	MaxDelay:   5 * time.Second,
	Multiplier: 2.0,
}

// WithRetry는 실패 시 재시도합니다.
func WithRetry(config RetryConfig, fn func() error) error {
	var lastErr error
	delay := config.Delay

	for i := 0; i <= config.MaxRetries; i++ {
		if err := fn(); err != nil {
			lastErr = err
			log.Printf("Attempt %d failed: %v, retrying in %v", i+1, err, delay)

			if i < config.MaxRetries {
				time.Sleep(delay)
				// 지수 백오프
				delay = time.Duration(float64(delay) * config.Multiplier)
				if delay > config.MaxDelay {
					delay = config.MaxDelay
				}
			}
		} else {
			return nil
		}
	}

	return lastErr
}

// WithRetryContext는 context와 함께 재시도합니다.
func WithRetryContext(ctx context.Context, config RetryConfig, fn func(ctx context.Context) error) error {
	var lastErr error
	delay := config.Delay

	for i := 0; i <= config.MaxRetries; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := fn(ctx); err != nil {
			lastErr = err
			log.Printf("Attempt %d failed: %v, retrying in %v", i+1, err, delay)

			if i < config.MaxRetries {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(delay):
				}
				delay = time.Duration(float64(delay) * config.Multiplier)
				if delay > config.MaxDelay {
					delay = config.MaxDelay
				}
			}
		} else {
			return nil
		}
	}

	return lastErr
}

// ========================================
// 디바운스/쓰로틀 헬퍼
// ========================================

// Debouncer는 연속 호출을 디바운스합니다.
type Debouncer struct {
	delay time.Duration
	timer *time.Timer
	mu    sync.Mutex
}

// NewDebouncer는 새 Debouncer를 생성합니다.
func NewDebouncer(delay time.Duration) *Debouncer {
	return &Debouncer{delay: delay}
}

// Debounce는 함수 호출을 디바운스합니다.
func (d *Debouncer) Debounce(fn func()) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.timer != nil {
		d.timer.Stop()
	}

	d.timer = time.AfterFunc(d.delay, fn)
}

// Throttler는 함수 호출을 쓰로틀합니다.
type Throttler struct {
	interval time.Duration
	lastRun  time.Time
	mu       sync.Mutex
}

// NewThrottler는 새 Throttler를 생성합니다.
func NewThrottler(interval time.Duration) *Throttler {
	return &Throttler{interval: interval}
}

// Throttle은 함수 호출을 쓰로틀합니다.
func (t *Throttler) Throttle(fn func()) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	if now.Sub(t.lastRun) >= t.interval {
		t.lastRun = now
		fn()
		return true
	}
	return false
}
