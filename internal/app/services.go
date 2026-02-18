package app

import (
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jptaku/server/internal/config"
	"github.com/jptaku/server/internal/pkg"
	"github.com/jptaku/server/internal/repository"
	"github.com/jptaku/server/internal/service"
	authSvc "github.com/jptaku/server/internal/service/auth"
	chatSvc "github.com/jptaku/server/internal/service/chat"
	feedbackSvc "github.com/jptaku/server/internal/service/feedback"
	flashSvc "github.com/jptaku/server/internal/service/flash"
	learningSvc "github.com/jptaku/server/internal/service/learning"
	"github.com/jptaku/server/internal/service/sentence"
	topicSvc "github.com/jptaku/server/internal/service/topic"
	userSvc "github.com/jptaku/server/internal/service/user"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Repositories 모든 저장소
type Repositories struct {
	DBManager    *repository.DBManager
	User         *repository.UserRepository
	Sentence     *repository.SentenceRepository
	Learning     *repository.LearningRepository
	Chat         *repository.ChatRepository
	Feedback     *repository.FeedbackRepository
}

// Services 모든 서비스
type Services struct {
	Auth     authSvc.Provider
	User     userSvc.Provider
	Sentence sentence.Provider
	Learning learningSvc.Provider
	Flash    flashSvc.Provider
	Chat     chatSvc.Provider
	Feedback feedbackSvc.Provider
	Topic    *topicSvc.Service
	Async    *service.AsyncService
}

// Infra 인프라 의존성
type Infra struct {
	JWTManager *pkg.JWTManager
	S3Client   *s3.Client
	BucketName string
}

// Dependencies 모든 의존성
type Dependencies struct {
	Repos    *Repositories
	Services *Services
	Infra    *Infra
}

// NewDependencies 모든 의존성 초기화
func NewDependencies(db *gorm.DB, cfg *config.Config, redisClient *redis.Client) *Dependencies {
	// Repositories
	repos := &Repositories{
		DBManager: repository.NewDBManager(db),
		User:      repository.NewUserRepository(db),
		Sentence:  repository.NewSentenceRepository(db),
		Learning:  repository.NewLearningRepository(db),
		Chat:      repository.NewChatRepository(db),
		Feedback:  repository.NewFeedbackRepository(db),
	}

	// Infrastructure
	jwtManager := pkg.NewJWTManager(cfg.JWT.Secret, cfg.JWT.ExpirationHours)

	s3Client := s3.New(s3.Options{
		Region:       "kr-standard",
		BaseEndpoint: aws.String(cfg.NCP_Storage.Endpoint),
		Credentials:  credentials.NewStaticCredentialsProvider(cfg.NCP_Storage.AccessKey, cfg.NCP_Storage.SecretKey, ""),
	})

	infra := &Infra{
		JWTManager: jwtManager,
		S3Client:   s3Client,
		BucketName: cfg.NCP_Storage.BucketName,
	}

	// Services
	asyncService := service.NewAsyncService(4, 100)

	authService := authSvc.NewService(repos.DBManager, repos.User, jwtManager)
	if cfg.Google.ClientID != "" && cfg.Google.ClientSecret != "" {
		googleOAuth := pkg.NewGoogleOAuthManager(
			cfg.Google.ClientID,
			cfg.Google.ClientSecret,
			cfg.Google.RedirectURL,
		)
		authService.SetGoogleOAuth(googleOAuth)
		log.Println("Google OAuth initialized")
	} else {
		log.Println("Warning: Google OAuth not configured")
	}

	sentenceService := sentence.NewService(repos.Sentence, repos.User)
	userService := userSvc.NewService(repos.User, sentenceService)
	learningService := learningSvc.NewService(repos.Learning, repos.Sentence)
	chatService := chatSvc.NewService(repos.Chat, repos.Sentence, repos.User, cfg.OpenAI.APIKey, cfg.OpenAI.Model)

	// VoiceVox TTS 설정
	if cfg.VoiceVox.VoiceVoxURL != "" {
		chatService.SetVoiceVox(cfg.VoiceVox.VoiceVoxURL)
	}

	feedbackService := feedbackSvc.NewService(repos.Feedback, repos.Chat, repos.Sentence)

	// Flash Service (DB 조회만 수행, OpenAI 호출 없음)
	flashService := flashSvc.NewService(repos.Sentence, repos.Learning)

	// Topic Service (Redis 기반 토픽 데이터 로더)
	var topicService *topicSvc.Service
	if redisClient != nil {
		topicService = topicSvc.NewService(redisClient)
		chatService.SetTopicService(topicService)
		log.Println("Topic service initialized with Redis")
	}

	services := &Services{
		Auth:     authService,
		User:     userService,
		Sentence: sentenceService,
		Learning: learningService,
		Flash:    flashService,
		Chat:     chatService,
		Feedback: feedbackService,
		Topic:    topicService,
		Async:    asyncService,
	}

	return &Dependencies{
		Repos:    repos,
		Services: services,
		Infra:    infra,
	}
}
