# 온보딩 및 문장 시스템 데이터 명세

이 문서는 온보딩 프로세스와 문장 시스템의 데이터 구조를 정리한 파일입니다.

---

## 1. 카테고리 (OnboardingCategory)

온보딩 및 문장 시스템에서 사용하는 콘텐츠 카테고리입니다. (온보딩 시 다중 선택 가능)

| ID | Label | 설명 |
| :--- | :--- | :--- |
| `1` | 애니 | 애니메이션 관련 표현, 캐릭터 묘사, 스토리 감상 |
| `2` | 게임 | 게임 플레이, RPG, 모바일 게임, e스포츠 관련 표현 |
| `3` | 음악 | J-POP, 아이돌, 애니송, 콘서트 관련 표현 |
| `4` | 영화 | 일본 영화 감상, 리뷰, 영화관 관련 표현 |
| `5` | 드라마 | 일본 드라마 감상, 배우, 스토리 관련 표현 |

**백엔드 타입**: `pkg.OnboardingCategory`

---

## 2. 일본어 레벨 (Level)

사용자의 현재 일본어 구사 능력입니다. (단일 선택)

| ID | Label | 설명 |
| :--- | :--- | :--- |
| `5` | N5 | JLPT N5 수준 - 짧고 쉬운 표현, 기본 인사 |
| `4` | N4 | JLPT N4 수준 - 일상 회화, て형/ない형 등 기본 활용 |
| `3` | N3 | JLPT N3 수준 - 자신의 생각 표현, 조금 긴 문장 |

**백엔드 타입**: `pkg.Level`

---

## 3. 문장 키 (SentenceKey)

카테고리와 레벨의 조합으로 문장을 분류합니다.

**형식**: `{category}_{level}` (예: `1_5`, `2_4`, `3_3`)

**전체 조합**: 5 Category × 3 Level = **15개**

| SentenceKey | Category | Level |
| :--- | :--- | :--- |
| `1_5` | 애니 | N5 |
| `1_4` | 애니 | N4 |
| `1_3` | 애니 | N3 |
| `2_5` | 게임 | N5 |
| `2_4` | 게임 | N4 |
| `2_3` | 게임 | N3 |
| ... | ... | ... |

---

## API 스펙

### POST `/api/user/onboarding`

온보딩 정보를 저장합니다.

**Request Body:**
```json
{
  "categories": [1, 2, 3],  // OnboardingCategory 배열 (필수, 최소 1개)
  "level": 5                 // Level 값 (필수, 3/4/5 중 하나)
}
```

**Response:**
```json
{
  "id": 1,
  "user_id": 123,
  "categories": [1, 2, 3],
  "level": 5,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

---

## 데이터베이스 스키마

### user_onboardings 테이블

| Column | Type | Description |
| :--- | :--- | :--- |
| `id` | SERIAL PRIMARY KEY | 고유 ID |
| `user_id` | INTEGER UNIQUE NOT NULL | 사용자 ID (FK) |
| `categories` | JSONB | OnboardingCategory 배열 (1~5) |
| `level` | INTEGER NOT NULL | Level 값 (3/4/5) |
| `created_at` | TIMESTAMP | 생성 시각 |
| `updated_at` | TIMESTAMP | 수정 시각 |

### sentences 테이블

| Column | Type | Description |
| :--- | :--- | :--- |
| `id` | SERIAL PRIMARY KEY | 고유 ID |
| `sentence_key` | VARCHAR(20) NOT NULL | 조합 키 (예: "1_5") |
| `jp` | TEXT NOT NULL | 일본어 문장 |
| `kr` | TEXT NOT NULL | 한국어 번역 |
| `romaji` | TEXT | 로마지 |
| `level` | INTEGER NOT NULL | Level 값 (3/4/5) |
| `category` | INTEGER NOT NULL | OnboardingCategory 값 (1~5) |
| `audio_url` | VARCHAR(500) | 음성 파일 URL |
| `created_at` | TIMESTAMP | 생성 시각 |
| `updated_at` | TIMESTAMP | 수정 시각 |

