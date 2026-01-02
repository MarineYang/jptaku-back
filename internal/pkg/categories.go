package pkg

import "fmt"

// OnboardingCategory 온보딩 시 사용자가 선택하는 관심 카테고리 (5가지)
// 문장 시스템에서도 이 카테고리를 사용
type OnboardingCategory int

const (
	OnboardingCategoryAnime OnboardingCategory = 1 // 애니
	OnboardingCategoryGame  OnboardingCategory = 2 // 게임
	OnboardingCategoryMusic OnboardingCategory = 3 // 음악
	OnboardingCategoryMovie OnboardingCategory = 4 // 영화
	OnboardingCategoryDrama OnboardingCategory = 5 // 드라마
)

// Level 일본어 레벨 (N3/N4/N5)
type Level int

const (
	LevelN5 Level = 5 // N5 수준
	LevelN4 Level = 4 // N4 수준
	LevelN3 Level = 3 // N3 수준
)

// SentenceKey 문장 생성을 위한 조합 키 (OnboardingCategory + Level)
// 형식: "CATEGORY_LEVEL" (예: "1_5", "2_4")
type SentenceKey string

// AllOnboardingCategories 모든 OnboardingCategory 목록
var AllOnboardingCategories = []OnboardingCategory{
	OnboardingCategoryAnime,
	OnboardingCategoryGame,
	OnboardingCategoryMusic,
	OnboardingCategoryMovie,
	OnboardingCategoryDrama,
}

// AllLevels 모든 Level 목록
var AllLevels = []Level{
	LevelN5, // 5
	LevelN4, // 4
	LevelN3, // 3
}

// AllSentenceKeys 모든 문장 조합 키 (5 Category × 3 Level = 15개)
var AllSentenceKeys []SentenceKey

func init() {
	AllSentenceKeys = GenerateAllSentenceKeys()
}

// GenerateAllSentenceKeys 모든 조합 키 생성
func GenerateAllSentenceKeys() []SentenceKey {
	keys := make([]SentenceKey, 0, len(AllOnboardingCategories)*len(AllLevels))
	for _, level := range AllLevels {
		for _, category := range AllOnboardingCategories {
			keys = append(keys, NewSentenceKey(category, level))
		}
	}
	return keys
}

// NewSentenceKey SentenceKey 생성
func NewSentenceKey(category OnboardingCategory, level Level) SentenceKey {
	return SentenceKey(fmt.Sprintf("%d_%d", category, level))
}

// Parse SentenceKey를 OnboardingCategory와 Level로 분리
func (k SentenceKey) Parse() (OnboardingCategory, Level, error) {
	var category, level int
	_, err := fmt.Sscanf(string(k), "%d_%d", &category, &level)
	if err != nil {
		return 0, 0, err
	}
	return OnboardingCategory(category), Level(level), nil
}

// Category SentenceKey에서 OnboardingCategory 추출
func (k SentenceKey) Category() OnboardingCategory {
	category, _, _ := k.Parse()
	return category
}

// Level SentenceKey에서 Level 추출
func (k SentenceKey) Level() Level {
	_, level, _ := k.Parse()
	return level
}

// String SentenceKey를 문자열로 변환
func (k SentenceKey) String() string {
	return string(k)
}

// Name OnboardingCategory의 한글 이름 반환
func (c OnboardingCategory) Name() string {
	names := map[OnboardingCategory]string{
		OnboardingCategoryAnime: "애니",
		OnboardingCategoryGame:  "게임",
		OnboardingCategoryMusic: "음악",
		OnboardingCategoryMovie: "영화",
		OnboardingCategoryDrama: "드라마",
	}
	if name, ok := names[c]; ok {
		return name
	}
	return "알 수 없음"
}

// Name Level의 한글 이름 반환
func (l Level) Name() string {
	names := map[Level]string{
		LevelN5: "N5",
		LevelN4: "N4",
		LevelN3: "N3",
	}
	if name, ok := names[l]; ok {
		return name
	}
	return "알 수 없음"
}
