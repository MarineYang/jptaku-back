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
	learningSvc "github.com/jptaku/server/internal/service/learning"
	"github.com/jptaku/server/internal/service/sentence"
	userSvc "github.com/jptaku/server/internal/service/user"
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
	Chat     chatSvc.Provider
	Feedback feedbackSvc.Provider
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
func NewDependencies(db *gorm.DB, cfg *config.Config) *Dependencies {
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
	chatService := chatSvc.NewService(repos.Chat, repos.Sentence)
	feedbackService := feedbackSvc.NewService(repos.Feedback, repos.Chat)

	services := &Services{
		Auth:     authService,
		User:     userService,
		Sentence: sentenceService,
		Learning: learningService,
		Chat:     chatService,
		Feedback: feedbackService,
		Async:    asyncService,
	}

	return &Dependencies{
		Repos:    repos,
		Services: services,
		Infra:    infra,
	}
}
