package pkg

import "fmt"

// type Category int

// const (
// 	CategoryAnime     Category = 1 // 애니/만화
// 	CategoryGame      Category = 2 // 게임
// 	CategoryMusic     Category = 3 // 음악
// 	CategoryVtuber    Category = 4 // 버튜버(독립)
// 	CategoryLifestyle Category = 5 // 오타쿠 라이프스타일
// 	CategorySituation Category = 6 // 실전 오타쿠 상황
// )

type Category int

const (
    CategoryImpression Category = 1 // 감상·평가 (재밌다/별로다/인상적이다)
    CategoryReaction   Category = 2 // 공감·맞장구 (나도 그래/그건 좀 의외)
    CategoryQuestion   Category = 3 // 의문·확장 (왜?/어떻게 생각해?)
    CategoryComparison Category = 4 // 비교·선호 (A보다 B/전작 vs 이번작)
    CategorySituation  Category = 5 // 상황·행동 (현장/주문/만남/실전)
)

// SubCategory: 통합된 상세 카테고리 (총 17개)
type SubCategory int

// Anime (100번대) - 3개
const (
	SubCategoryAnimeBattleFantasySF SubCategory = 101 // 배틀/판타지·SF (이세계/배틀/SF로봇 통합)
	SubCategoryAnimeSliceLoveEmo    SubCategory = 102 // 일상/러브코미·감성 (러브코미/일상/감성 통합)
	SubCategoryAnimeStoryMystery    SubCategory = 103 // 서사/추리 (미스터리/추리 중심)
)

// Game (200번대) - 3개
const (
	SubCategoryGameRpgGacha     SubCategory = 201 // RPG/가챠 (JRPG+모바일가챠)
	SubCategoryGameRhythm       SubCategory = 202 // 리듬게임
	SubCategoryGameActionVsShoo SubCategory = 203 // 액션/대전·슈터 (FPS/격투/콘솔대전 감성)
)

// Music (300번대) - 3개
const (
	SubCategoryMusicJpop  SubCategory = 301 // Jpop
	SubCategoryMusicIdol  SubCategory = 302 // 아이돌
	SubCategoryMusicAnime SubCategory = 303 // 애니송
)

// VTuber (350번대) - 1개
const (
	SubCategoryVtuber SubCategory = 351 // 버튜버 (配信/切り抜き/スパチャ/メンシ 등)
)

// Lifestyle (400번대) - 3개
const (
	SubCategoryLifePilgrimageTravel SubCategory = 401 // 성지순례/오타쿠여행
	SubCategoryLifeGoodsCollect     SubCategory = 402 // 굿즈/수집 (굿즈구매+피규어/프라모델)
	SubCategoryLifeComiketDoujin    SubCategory = 403 // 오프라인 이벤트·동인(코미케)
)

// Situation (500번대) - 4개
const (
	SubCategorySitShoppingOrder      SubCategory = 501 // 쇼핑/주문 (일본사이트·예약 포함)
	SubCategorySitOnsiteLiveGreeting SubCategory = 502 // 현장/라이브 (행사인사+콘서트/라이브)
	SubCategorySitOtakuTalk          SubCategory = 503 // 오타쿠 대화 (덕질 토크/애니 얘기)
	SubCategorySitCollabCafeGameCtr  SubCategory = 504 // 콜라보카페/게임센터
)

// Level 일본어 레벨
type Level int

const (
	LevelBeginner Level = 0 // 완전 초입문 (히라가나/가타카나 학습 중)
	LevelN5       Level = 1 // 기본 인사 가능 (N5 수준)
	LevelN4       Level = 2 // 일상 회화 조금 가능 (N4 수준)
	LevelN3       Level = 3 // 생각 표현 가능 (N3 수준)
)

// Purpose 학습 목적
type Purpose int

const (
	PurposeWatchAnime  Purpose = 1 // 자막 없이 애니·만화 즐기려고
	PurposeTalkFriends Purpose = 2 // 일본 친구와 대화하고 싶어서
	PurposeTravel      Purpose = 3 // 일본 여행에서 말하고 싶어서
	PurposeVtuber      Purpose = 4 // 버튜버 방송/콘텐츠 이해하고 싶어서
	PurposeGame        Purpose = 5 // 좋아하는 게임의 일본 서버/콘텐츠 즐기려고
	PurposeGoodsEvent  Purpose = 6 // 굿즈 구매·이벤트 참가 때문에
	PurposeOther       Purpose = 7 // 기타
)

// CategoryParent SubCategory가 속한 Category 반환
func (s SubCategory) CategoryParent() Category {
	switch {
	case s >= 100 && s < 200:
		return CategoryAnime
	case s >= 200 && s < 300:
		return CategoryGame
	case s >= 300 && s < 350:
		return CategoryMusic
	case s >= 350 && s < 400:
		return CategoryVtuber // 버튜버 추가
	case s >= 400 && s < 500:
		return CategoryLifestyle
	case s >= 500 && s < 600:
		return CategorySituation
	default:
		return 0
	}
}

// SentenceKey 문장 생성을 위한 조합 키 (SubCategory + Level)
// 형식: "SUBCATEGORY_LEVEL" (예: "101_0", "351_2")
type SentenceKey string

// AllSubCategories 모든 SubCategory 목록
var AllSubCategories = []SubCategory{
	// Anime (3개)
	SubCategoryAnimeBattleFantasySF, // 101
	SubCategoryAnimeSliceLoveEmo,    // 102
	SubCategoryAnimeStoryMystery,    // 103
	// Game (3개)
	SubCategoryGameRpgGacha,     // 201
	SubCategoryGameRhythm,       // 202
	SubCategoryGameActionVsShoo, // 203
	// Music (3개)
	SubCategoryMusicJpop,  // 301
	SubCategoryMusicIdol,  // 302
	SubCategoryMusicAnime, // 303
	// VTuber (1개)
	SubCategoryVtuber, // 351
	// Lifestyle (3개)
	SubCategoryLifePilgrimageTravel, // 401
	SubCategoryLifeGoodsCollect,     // 402
	SubCategoryLifeComiketDoujin,    // 403
	// Situation (4개)
	SubCategorySitShoppingOrder,      // 501
	SubCategorySitOnsiteLiveGreeting, // 502
	SubCategorySitOtakuTalk,          // 503
	SubCategorySitCollabCafeGameCtr,  // 504
}

// AllLevels 모든 Level 목록
var AllLevels = []Level{
	LevelBeginner, // 0
	LevelN5,       // 1
	LevelN4,       // 2
	LevelN3,       // 3
}

// AllSentenceKeys 모든 문장 조합 키 (17 SubCategory × 4 Level = 68개)
var AllSentenceKeys []SentenceKey

func init() {
	AllSentenceKeys = GenerateAllSentenceKeys()
}

// GenerateAllSentenceKeys 모든 조합 키 생성
func GenerateAllSentenceKeys() []SentenceKey {
	keys := make([]SentenceKey, 0, len(AllSubCategories)*len(AllLevels))
	for _, level := range AllLevels {
		for _, subCat := range AllSubCategories {
			keys = append(keys, NewSentenceKey(subCat, level))
		}
	}
	return keys
}

// NewSentenceKey SentenceKey 생성
func NewSentenceKey(subCategory SubCategory, level Level) SentenceKey {
	return SentenceKey(fmt.Sprintf("%d_%d", subCategory, level))
}

// Parse SentenceKey를 SubCategory와 Level로 분리
func (k SentenceKey) Parse() (SubCategory, Level, error) {
	var subCat, level int
	_, err := fmt.Sscanf(string(k), "%d_%d", &subCat, &level)
	if err != nil {
		return 0, 0, err
	}
	return SubCategory(subCat), Level(level), nil
}

// SubCategory SentenceKey에서 SubCategory 추출
func (k SentenceKey) SubCategory() SubCategory {
	subCat, _, _ := k.Parse()
	return subCat
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

// SubCategoryName SubCategory의 한글 이름 반환
func (s SubCategory) Name() string {
	names := map[SubCategory]string{
		SubCategoryAnimeBattleFantasySF: "배틀/판타지·SF",
		SubCategoryAnimeSliceLoveEmo:    "일상/러브코미·감성",
		SubCategoryAnimeStoryMystery:    "서사/추리",
		SubCategoryGameRpgGacha:         "RPG/가챠",
		SubCategoryGameRhythm:           "리듬게임",
		SubCategoryGameActionVsShoo:     "액션/대전·슈터",
		SubCategoryMusicJpop:            "J-POP",
		SubCategoryMusicIdol:            "아이돌",
		SubCategoryMusicAnime:           "애니송",
		SubCategoryVtuber:               "버튜버",
		SubCategoryLifePilgrimageTravel: "성지순례/여행",
		SubCategoryLifeGoodsCollect:     "굿즈/수집",
		SubCategoryLifeComiketDoujin:    "코미케/동인",
		SubCategorySitShoppingOrder:     "쇼핑/주문",
		SubCategorySitOnsiteLiveGreeting: "현장/라이브",
		SubCategorySitOtakuTalk:          "오타쿠 대화",
		SubCategorySitCollabCafeGameCtr:  "콜라보카페/게임센터",
	}
	if name, ok := names[s]; ok {
		return name
	}
	return "알 수 없음"
}

// LevelName Level의 한글 이름 반환
func (l Level) Name() string {
	names := map[Level]string{
		LevelBeginner: "초입문 (Lv0)",
		LevelN5:       "N5 수준 (Lv1)",
		LevelN4:       "N4 수준 (Lv2)",
		LevelN3:       "N3 수준 (Lv3)",
	}
	if name, ok := names[l]; ok {
		return name
	}
	return "알 수 없음"
}
