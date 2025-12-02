package pkg

// Category 관심사 대분류
type Category int

const (
	CategoryAnime     Category = 1 // 애니/만화
	CategoryGame      Category = 2 // 게임
	CategoryMusic     Category = 3 // 음악/Jpop/버튜버
	CategoryLifestyle Category = 4 // 오타쿠 라이프스타일
	CategorySituation Category = 5 // 실전 오타쿠 상황
)

// SubCategory 상세 장르
type SubCategory int

// Anime 장르 (100번대)
const (
	SubCategoryIsekai      SubCategory = 101 // 이세계/판타지
	SubCategoryLoveComedy  SubCategory = 102 // 러브코미디
	SubCategorySliceOfLife SubCategory = 103 // 일상물
	SubCategoryBattle      SubCategory = 104 // 배틀/액션
	SubCategorySports      SubCategory = 105 // 스포츠물
	SubCategorySFRobot     SubCategory = 106 // SF/로봇
	SubCategoryMusicIdol   SubCategory = 107 // 음악/아이돌물
	SubCategoryMystery     SubCategory = 108 // 미스터리/추리
)

// Game 장르 (200번대)
const (
	SubCategoryJRPG     SubCategory = 201 // JRPG
	SubCategoryGacha    SubCategory = 202 // 모바일 가챠게임
	SubCategoryRhythm   SubCategory = 203 // 리듬게임
	SubCategoryFPS      SubCategory = 204 // FPS
	SubCategoryNintendo SubCategory = 205 // 닌텐도 게임
	SubCategoryFighting SubCategory = 206 // 격투 게임
)

// Music 장르 (300번대)
const (
	SubCategoryJpop     SubCategory = 301 // Jpop
	SubCategoryVocaloid SubCategory = 302 // Vocaloid
	SubCategoryAnisong  SubCategory = 303 // 애니송
	SubCategoryIdol     SubCategory = 304 // 아이돌
	SubCategoryVtuber   SubCategory = 305 // 버튜버
)

// Lifestyle 장르 (400번대)
const (
	SubCategoryPilgrimage SubCategory = 401 // 성지순례
	SubCategoryGoods      SubCategory = 402 // 굿즈 구매
	SubCategoryFigure     SubCategory = 403 // 피규어/프라모델
	SubCategoryComiket    SubCategory = 404 // 코미케/행사 참가
	SubCategoryAnimeCafe  SubCategory = 405 // 애니카페 방문
	SubCategoryGameCenter SubCategory = 406 // 게임센터 방문
)

// Situation 장르 (500번대)
const (
	SubCategoryGoodsReserve   SubCategory = 501 // 굿즈 예약하기
	SubCategoryEventGreeting  SubCategory = 502 // 행사에서 인사하기
	SubCategoryAnimeTalk      SubCategory = 503 // 친구와 애니 얘기하기
	SubCategoryJapanSiteOrder SubCategory = 504 // 일본 사이트 주문하기
	SubCategoryOtakuTravel    SubCategory = 505 // 일본 여행 오타쿠 코스
	SubCategoryConcert        SubCategory = 506 // 콘서트/라이브 관람
)

// Level 일본어 레벨
type Level int

const (
	LevelBeginner Level = 0 // 완전 초입문 (히라가나/가타카나 학습 중)
	LevelN5       Level = 1 // 기본 인사 가능 (N5 수준)
	LevelN4       Level = 2 // 일상 회화 조금 가능 (N4 수준)
	LevelN3       Level = 3 // 생각 표현 가능 (N3 수준)
	LevelN2       Level = 4 // 능숙 (N2 수준)
	LevelN1       Level = 5 // 거의 원어민 수준 (N1 수준)
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
	case s >= 300 && s < 400:
		return CategoryMusic
	case s >= 400 && s < 500:
		return CategoryLifestyle
	case s >= 500 && s < 600:
		return CategorySituation
	default:
		return 0
	}
}
