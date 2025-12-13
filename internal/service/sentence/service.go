package sentence

// Service 문장 관련 비즈니스 로직 구현체
type Service struct {
	sentenceRepo SentenceRepository
	userRepo     UserRepository
	learningRepo LearningRepository
}

// 컴파일 타임에 인터페이스 구현 확인
var _ Provider = (*Service)(nil)

// NewService SentenceService 생성자
func NewService(sentenceRepo SentenceRepository, userRepo UserRepository) *Service {
	return &Service{
		sentenceRepo: sentenceRepo,
		userRepo:     userRepo,
	}
}

// SetLearningRepo LearningRepository 설정 (순환 의존성 방지)
func (s *Service) SetLearningRepo(learningRepo LearningRepository) {
	s.learningRepo = learningRepo
}
