package pkg

import "fmt"

// Domain 콘텐츠 도메인
type Domain string

const (
	DomainAnime Domain = "anime"
	DomainGame  Domain = "game"
	DomainMusic Domain = "music"
	DomainMovie Domain = "movie"
	DomainDrama Domain = "drama"
)

// AllDomains 모든 도메인 목록
var AllDomains = []Domain{
	DomainAnime,
	DomainGame,
	DomainMusic,
	DomainMovie,
	DomainDrama,
}

// DomainFromCategory 온보딩 카테고리 int → Domain 변환
func DomainFromCategory(category int) Domain {
	switch category {
	case 1:
		return DomainAnime
	case 2:
		return DomainGame
	case 3:
		return DomainMusic
	case 4:
		return DomainMovie
	case 5:
		return DomainDrama
	default:
		return DomainAnime
	}
}

// DomainsFromCategories 온보딩 카테고리 배열 → Domain 배열 변환
func DomainsFromCategories(categories []int) []string {
	domains := make([]string, 0, len(categories))
	for _, c := range categories {
		domains = append(domains, string(DomainFromCategory(c)))
	}
	return domains
}

// LevelsForUser 온보딩 레벨 → 쿼리용 Level int 목록 반환
// DB와 온보딩 모두 5=N5, 4=N4, 3=N3 통일
// 유저 레벨 이하의 모든 레벨을 반환 (예: N4 유저 → [5, 4])
func LevelsForUser(onboardingLevel int) []int {
	switch onboardingLevel {
	case 3:
		return []int{5, 4, 3} // N5 + N4 + N3
	case 4:
		return []int{5, 4} // N5 + N4
	case 5:
		return []int{5} // N5
	default:
		return []int{5} // N5
	}
}

// LevelNameFromInt Level int → 표시용 문자열
func LevelNameFromInt(level int) string {
	switch level {
	case 5:
		return "N5"
	case 4:
		return "N4"
	case 3:
		return "N3"
	default:
		return fmt.Sprintf("N%d", level)
	}
}
