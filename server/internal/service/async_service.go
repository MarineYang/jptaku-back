package service

import (
	"context"
	"log"
	"time"

	"github.com/jptaku/server/internal/pkg"
)

// AsyncService는 비동기 작업을 관리합니다.
type AsyncService struct {
	workerPool *pkg.WorkerPool
}

// NewAsyncService는 새 AsyncService를 생성합니다.
func NewAsyncService(workers, queueSize int) *AsyncService {
	pool := pkg.NewWorkerPool(workers, queueSize)
	pool.Start()

	// 결과 처리 goroutine 시작
	go func() {
		for result := range pool.Results() {
			if result.Error != nil {
				log.Printf("Async job %s failed: %v", result.JobID, result.Error)
			} else {
				log.Printf("Async job %s completed successfully", result.JobID)
			}
		}
	}()

	return &AsyncService{workerPool: pool}
}

// Stop은 비동기 서비스를 정상 종료합니다.
func (s *AsyncService) Stop() {
	s.workerPool.StopWithTimeout(30 * time.Second)
}

// ========================================
// 백그라운드 작업 제출
// ========================================

// SubmitFeedbackCalculation은 피드백 계산을 백그라운드에서 실행합니다.
func (s *AsyncService) SubmitFeedbackCalculation(sessionID uint, calculateFn func(ctx context.Context, sessionID uint) error) {
	jobID := "feedback_calc_" + string(rune(sessionID))

	s.workerPool.SubmitFuncWithTimeout(jobID, 30*time.Second, func(ctx context.Context) error {
		return calculateFn(ctx, sessionID)
	})
}

// SubmitNotification은 알림 발송을 백그라운드에서 실행합니다.
func (s *AsyncService) SubmitNotification(userID uint, notifyFn func(ctx context.Context, userID uint) error) {
	jobID := "notification_" + string(rune(userID))

	s.workerPool.SubmitFuncWithTimeout(jobID, 10*time.Second, func(ctx context.Context) error {
		return notifyFn(ctx, userID)
	})
}

// SubmitStatsUpdate는 통계 업데이트를 백그라운드에서 실행합니다.
func (s *AsyncService) SubmitStatsUpdate(userID uint, updateFn func(ctx context.Context, userID uint) error) {
	jobID := "stats_update_" + string(rune(userID))

	s.workerPool.Submit(pkg.Job{
		ID:      jobID,
		Timeout: 1 * time.Minute,
		Handler: func(ctx context.Context) error {
			return updateFn(ctx, userID)
		},
	})
}

// SubmitEmailSend는 이메일 발송을 백그라운드에서 실행합니다 (재시도 포함).
func (s *AsyncService) SubmitEmailSend(email string, sendFn func(ctx context.Context, email string) error) {
	jobID := "email_" + email

	s.workerPool.SubmitFuncWithTimeout(jobID, 30*time.Second, func(ctx context.Context) error {
		// 재시도 로직 포함
		return pkg.WithRetryContext(ctx, pkg.RetryConfig{
			MaxRetries: 3,
			Delay:      500 * time.Millisecond,
			MaxDelay:   5 * time.Second,
			Multiplier: 2.0,
		}, func(ctx context.Context) error {
			return sendFn(ctx, email)
		})
	})
}

// ========================================
// 즉시 실행 (Fire and Forget)
// ========================================

// RunAsync는 간단한 비동기 작업을 즉시 실행합니다.
func RunAsync(fn func()) {
	pkg.Go(fn)
}

// RunAsyncWithContext는 context가 있는 비동기 작업을 실행합니다.
func RunAsyncWithContext(ctx context.Context, fn func(ctx context.Context)) {
	pkg.GoWithContext(ctx, fn)
}

// ========================================
// 병렬 실행 (결과 대기)
// ========================================

// RunParallel은 여러 작업을 병렬로 실행하고 결과를 기다립니다.
func RunParallel(fns ...func() error) []error {
	return pkg.RunParallel(fns...)
}

// RunParallelWithContext는 context와 함께 병렬 실행합니다.
func RunParallelWithContext(ctx context.Context, fns ...func(ctx context.Context) error) []error {
	return pkg.RunParallelWithContext(ctx, fns...)
}

// ========================================
// 사용 예시
// ========================================

/*
예시 1: 채팅 세션 종료 시 피드백 계산을 백그라운드로 처리

func (s *ChatService) EndSession(sessionID uint) (*model.ChatSession, error) {
    session, err := s.chatRepo.UpdateSession(sessionID, ...)
    if err != nil {
        return nil, err
    }

    // 피드백 계산은 백그라운드에서 (응답은 바로 반환)
    s.asyncService.SubmitFeedbackCalculation(sessionID, func(ctx context.Context, id uint) error {
        return s.feedbackService.CalculateFeedback(ctx, id)
    })

    return session, nil
}

예시 2: 여러 외부 API를 병렬로 호출

func (s *SomeService) GetAggregatedData(ctx context.Context) (*AggregatedData, error) {
    var userData *UserData
    var sentenceData *SentenceData

    errors := RunParallelWithContext(ctx,
        func(ctx context.Context) error {
            var err error
            userData, err = s.fetchUserData(ctx)
            return err
        },
        func(ctx context.Context) error {
            var err error
            sentenceData, err = s.fetchSentenceData(ctx)
            return err
        },
    )

    // 에러 확인
    for _, err := range errors {
        if err != nil {
            return nil, err
        }
    }

    return &AggregatedData{User: userData, Sentence: sentenceData}, nil
}

예시 3: 외부 API 호출에 재시도 로직 적용

func (s *OpenAIService) CallWithRetry(ctx context.Context, prompt string) (string, error) {
    var result string

    err := pkg.WithRetryContext(ctx, pkg.DefaultRetryConfig, func(ctx context.Context) error {
        var err error
        result, err = s.client.Call(ctx, prompt)
        return err
    })

    return result, err
}
*/
