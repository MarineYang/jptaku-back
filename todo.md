# 일타쿠 TODO.md

> 목표: 일본어 못하는 오타쿠를 위한 **하루 5문장 + 실시간 대화** MVP 완성

---

## 0. 전체 마일스톤

- [ ] M1. 기본 환경 세팅 (Go Webserver + RN + DB/Redis)  
- [ ] M2. **하루 5문장 학습 루프** (홈 → 오늘의 5문장 → 문장 상세)  
- [ ] M3. **실전 대화 UX 목업 완성** (채팅 UI + 추천 문장 + 오늘 5문장 연동)  
- [ ] M4. **피드백 화면**을 오늘 학습/대화와 연결  
- [ ] M5. OpenAI Realtime API 연동 PoC  
- [ ] M6. 간단한 내부 테스트 & 릴리즈 준비

---

## 1. 인프라 / 아키텍처

### 1-1. 기본 인프라

- [x] Go 백엔드 프로젝트 초기화
  - [x] Go Modules 세팅 (`go mod init`)
  - [x] `server/cmd/api`, `server/internal/api`, `server/internal/service`, `server/internal/model`, `server/internal/config` 기본 구조
- [x] DB 선택 및 연결 (PostgreSQL)
  - [x] GORM + AutoMigrate 세팅
- [ ] Redis 연결 세팅 (구현 완료, 서버 연결 필요)
  - [ ] 세션 상태 / WebSocket 매핑 / 간단 캐시 용도

### 1-2. 서버 아키텍처 (논리)

- [ ] Nginx or API Gateway 초간단 설정 (개발 단계에선 생략 가능)
  - `/api/*` → Go Gin HTTP, `/ws/*` → Go Gin WebSocket
- [x] Go 내부 레이어 분리
  - [x] `internal/api/` : 핸들러 (Gin Handler)
  - [x] `internal/service/` : 도메인 로직
  - [x] `internal/repository/` : GORM DB 접근
  - [x] `internal/api/*/dto.go` : 요청/응답 DTO
- [ ] WebSocket 엔드포인트 설계
  - [ ] `/ws/conversation` 기본 핸들러
  - [ ] 연결 시 JWT 검증 & Redis에 `{connection_id -> user_id, session_id}` 저장

---

## 2. API 설계 & 스켈레톤

### 2-1. Auth / User

- [ ] `POST /api/auth/login` (소셜/이메일 토큰 받는 구조만 정의)
- [ ] `POST /api/auth/refresh`
- [ ] `POST /api/auth/logout`
- [ ] `GET  /api/user/me`
- [ ] `PUT  /api/user/profile`
- [ ] `POST /api/user/onboarding` (관심사/레벨/목적 저장)
- [ ] `GET  /api/user/settings`
- [ ] `PUT  /api/user/settings`

### 2-2. Daily Learning (5문장)

- [ ] `GET  /api/sentences/daily`
  - 유저 관심사/레벨 기반 오늘 5문장 + `daily_set_id` 반환 (처음 호출 시 생성, 이후 동일 세트 유지)
- [ ] `GET  /api/sentences/:id`
- [ ] `GET  /api/sentences/history`
- [ ] `POST /api/learning/progress`
  - 문장별: 이해/말하기/확인/암기 완료 상태 업데이트
- [ ] (옵션) `GET/POST /api/sentences/bookmarks` 즐겨찾기
- [ ] (옵션) `GET /api/meta/interests`, `GET /api/meta/levels`

### 2-3. Real-time Conversation

- [ ] `POST /api/chat/session`  (세션 생성, today_set_id 연결)
- [ ] `POST /api/chat/session/:id/end` (세션 종료)
- [ ] `GET  /api/chat/session/:id` (대화 로그, 메타 정보)
- [ ] `GET  /api/chat/sessions` (최근 세션 목록)
- [ ] `POST /api/rtc/signaling` (WebRTC SDP/ICE 교환 – 이후 단계에서 구현)

#### WebSocket 이벤트 설계 (문자열 타입만 정의)

- [ ] `audio:stream`  (실제는 WebRTC, 여기선 타입만 정의)
- [ ] `text:transcript` (실시간 STT 결과 / 자막)
- [ ] `session:update` (추천 문장, 오늘 5문장 사용 카운트 등)
- [ ] `control:interrupt` (발화 중단)
- [ ] 클라이언트 → 서버 이벤트 타입 스펙 문서화

### 2-4. Feedback & Stats

- [ ] `GET /api/feedback/:sessionId`
- [ ] `GET /api/stats/categories` (오타쿠 카테고리별 진행도)
- [ ] `GET /api/stats/weekly`
- [ ] `GET /api/stats/today` (오늘의 요약: 사용 문장 수, 학습 시간, streak 등)

---

## 3. DB 모델링

- [ ] User / UserSettings / UserOnboarding
- [ ] Sentence
  - id, jp, kr, level, tags(애니/게임/성지순례/이벤트 등)
- [ ] DailySentenceSet
  - id, user_id, date, 5문장 리스트
- [ ] LearningProgress
  - user_id, sentence_id, daily_set_id, 단계(이해/말하기/확인/암기)
- [ ] ChatSession
  - id, user_id, daily_set_id, 시작/종료 시각, today_sentence_used_count, 점수 요약
- [ ] ChatMessage
  - session_id, speaker(ai/user), jp_text, kr_text, used_today_sentence_id(옵션)
- [ ] Feedback
  - session_id, 총점, 문법/발음/자연스러움, 요약 문장들(JSON)
- [ ] (옵션) StatsAggregate (주간/월간 캐시)

---

## 4. React Native 프론트 구조

### 4-1. 네비게이션 / 공통 레이아웃

- [ ] 탭 네비게이션 구성
  - [ ] 홈
  - [ ] 대화
  - [ ] 피드백
  - [ ] 마이
- [ ] 공통 스타일: Toss 느낌의 여백/타이포/카드 컴포넌트 정리

### 4-2. 온보딩 플로우

- [ ] 일본어 레벨 선택 (Lv0~Lv5)
- [ ] 일본어를 배우는 이유 선택
- [ ] 애니/게임/성지순례/이벤트 등 관심사 선택
- [ ] 온보딩 완료 → `POST /api/user/onboarding` 연동 (초기엔 mock)

### 4-3. 홈 화면

- [ ] 상단: “오늘의 5문장 진행도 (0/5)” + Progress Bar
- [ ] 섹션: `오늘의 5문장` 리스트 (카드 5개)
  - 클릭 시 → 문장 상세 화면
- [ ] CTA: “실전 대화 시작하기” 버튼 (chat session 생성 예정)

### 4-4. 오늘의 5문장 리스트 & 상세

- [ ] **오늘의 5문장 화면**
  - 카드 5개, 각 카드에:
    - JP / KR / “외운 문장” 체크 아이콘
- [ ] **문장 상세 화면 (탭 구조)**  
  - 상단: JP, romaji(옵션), KR, 재생 버튼
  - 탭 1: 이해하기
    - 단어 풀이, 핵심 문법, 예문
  - 탭 2: 말하기
    - 속도 선택, 마이크 버튼 (음성 녹음 UX만 목업)
  - 탭 3: 확인하기
    - 간단 퀴즈 (객관식 1문제 정도 목업)
  - 하단: `이 문장 외웠어요` 버튼 (로컬 상태 + `POST /learning/progress` 예정)

---

## 5. 실전 대화 화면 (채팅 + 추천 문장)

### 5-1. 기본 채팅 UI

- [ ] 화면 상단
  - [ ] 타이틀: “실전 대화”
  - [ ] 서브: `⭐ 오늘의 문장 활용`, `0/5 사용` 배지 + Progress Bar
- [ ] 채팅 영역
  - [ ] AI 말풍선 = 왼쪽, 유저 말풍선 = 오른쪽
  - [ ] AI 말풍선 탭 시 → 아래로 KR 번역 토글 표시
  - [ ] 유저 말풍선에 오늘 5문장 사용 시 🌟 뱃지 표시 (mock 로직)
- [ ] 하단 입력부
  - [ ] 추천 문장 칩 영역 (“💡 이렇게 말해볼까요?”)
  - [ ] 텍스트 입력창 + 마이크 버튼

### 5-2. 추천 문장 동작 (목업)

- [ ] 추천 칩 데이터 mock:
  - JP + KR + `isTodaySentence` + id
- [ ] 칩 탭 시:
  - [ ] 채팅 영역 맨 아래에 “힌트 버블” 추가  
        (💡 문장 + 번역 표시, 시스템 말풍선 스타일)
  - [ ] 입력창에 해당 일본어 문장 자동 입력
  - [ ] 실제 보내기는 유저가 send/마이크 눌렀을 때만 처리
- [ ] 유저가 그 문장을 보냈다고 가정하면:
  - [ ] 상단 `0/5 사용` → `1/5 사용`으로 증가 (로컬 상태)
  - [ ] 해당 말풍선에 🌟 뱃지

### 5-3. WebSocket 연동 준비

- [ ] WebSocket 클라이언트 훅/유틸 작성 (`useConversationWS` 등)
  - 아직은 서버 없이 dummy로 이벤트 시뮬레이션
- [ ] `text:transcript`, `session:update` 이벤트를 처리할 핸들러 구조만 설계

---

## 6. 피드백 화면 개편

### 6-1. 기존 총점/문법/발음/자연스러움 유지

- [ ] 현재 UI를 그대로 가져가되, 아래에 섹션 추가

### 6-2. “오늘의 5문장 사용 결과” 섹션

- [ ] 오늘의 5문장 카드 5개
  - JP / KR / 상태(✅/🔁/⛔)
- [ ] 상단: `오늘의 5문장 중 3개를 실제 대화에서 사용했어요!` 요약 문구

### 6-3. “오늘 대화 하이라이트” 섹션

- [ ] 하이라이트 카드 2~3개 (mock 데이터)
  - 타이틀 / JP / KR / 한줄 코멘트 / “다시 들어보기” 버튼(목업)

### 6-4. “오타쿠 카테고리 진행도” 섹션

- [ ] 애니 / 가챠 / 성지순례 / 이벤트 등 아이콘 + 퍼센트 Progress UI

### 6-5. “내일을 위한 한 줄 가이드 + CTA”

- [ ] 체크리스트 2~3개
- [ ] 하단 버튼: `오늘의 5문장 다시 보러가기` 또는 `지금 한 문장 더 말해보기`

---

## 7. Realtime / OpenAI 연동 (PoC 단계)

- [ ] 서버에서 OpenAI Realtime API 예제 코드 작성 (별도 PoC 스크립트)
- [ ] WebSocket 브리지 구조 잡기
  - 클라이언트 ↔ Go Webserver ↔ OpenAI Realtime
- [ ] 간단한 “일본어로 인사만 주고 받는” PoC부터 성공시키기
- [ ] 나중에 오늘 5문장 prompt / 시스템 프롬프트 반영

---

## 8. 초기 작업 우선순위 (이번 주 시작용)

1. [x] Go 프로젝트 구조 & DB 연결 세팅 (Gin + GORM)
2. [ ] RN 쪽 네비게이션 + 기본 탭 + 간단 홈 화면 표시  
3. [x] `GET /api/sentences/daily` 구현 + RN에서 호출해 "오늘의 5문장" 리스트 보여주기  
4. [ ] 문장 상세 화면(이해하기/말하기/확인하기) UI 목업 완성  
5. [ ] 실전 대화 화면 채팅 UI + 추천 칩 / 상단 0/5 표시까지 구현 (서버 연동 없이 mock 상태)  

이 5개만 끝내면,  
"하루 5문장 받고 → 문장 상세 들어가서 공부 → 실전 대화 화면에서 대화하는 듯한 UX" 까지는  
로컬 목업 기준으로 한 바퀴 돌아갈 수 있음.

---

## 9. Go 서버 구현 현황 (완료)

### 9-1. 프레임워크 & 라이브러리
- **Gin**: 웹 프레임워크 (빠르고 간편함)
- **GORM**: ORM (PostgreSQL 연결)
- **go-redis**: Redis 클라이언트
- **golang-jwt/jwt**: JWT 인증

### 9-2. 구현된 API 엔드포인트

#### Auth
- [x] `POST /api/auth/register` - 회원가입
- [x] `POST /api/auth/login` - 로그인
- [x] `POST /api/auth/refresh` - 토큰 갱신
- [x] `POST /api/auth/logout` - 로그아웃

#### User
- [x] `GET /api/user/me` - 내 정보 조회
- [x] `PUT /api/user/profile` - 프로필 수정
- [x] `POST /api/user/onboarding` - 온보딩 정보 저장
- [x] `GET /api/user/settings` - 설정 조회
- [x] `PUT /api/user/settings` - 설정 수정

#### Sentences
- [x] `GET /api/sentences/daily` - 오늘의 5문장
- [x] `GET /api/sentences/:id` - 문장 상세 조회
- [x] `GET /api/sentences/history` - 학습 히스토리

#### Learning
- [x] `POST /api/learning/progress` - 학습 진행 상황 업데이트
- [x] `GET /api/learning/today` - 오늘의 학습 진행 상황
- [x] `GET /api/learning/history` - 학습 히스토리

#### Chat
- [x] `POST /api/chat/session` - 세션 생성
- [x] `GET /api/chat/session/:id` - 세션 조회
- [x] `POST /api/chat/session/:id/end` - 세션 종료
- [x] `GET /api/chat/sessions` - 세션 목록

#### Feedback & Stats
- [x] `GET /api/feedback/:sessionId` - 피드백 조회
- [x] `GET /api/stats/today` - 오늘의 통계
- [x] `GET /api/stats/categories` - 카테고리별 진행도
- [x] `GET /api/stats/weekly` - 주간 통계

### 9-3. 서버 실행 방법

```bash
# 의존성 설치
cd server
go mod tidy

# 환경변수 설정 (env.example 참고)
cp env.example .env

# PostgreSQL, Redis 실행 후
# 서버 실행
go run ./cmd/api/main.go

# 또는 Makefile 사용
make run

# 빌드
make build

# 샘플 데이터 시드
make seed
```



## 10. 프로젝트 구조

```
/server
 ├── cmd
 │   ├── api
 │   │    └── main.go        // REST API 실행 파일
 │   ├── ws
 │   │    └── main.go        // WebSocket 서버 (미구현)
 │   └── migrate
 │        └── main.go        // DB 마이그레이션/시드 도구
 │
 ├── internal
 │   ├── api                 // REST 핸들러 (Gin)
 │   │    ├── auth/          // 인증 (handler.go, dto.go)
 │   │    ├── user/          // 유저 (handler.go, dto.go)
 │   │    ├── sentences/     // 문장 (handler.go, dto.go)
 │   │    ├── learning/      // 학습 (handler.go, dto.go)
 │   │    ├── chat/          // 채팅 (handler.go, dto.go)
 │   │    └── feedback/      // 피드백 (handler.go, dto.go)
 │   │
 │   ├── service             // 비즈니스 로직
 │   ├── repository          // DB 접근 (GORM)
 │   ├── model               // 데이터 모델
 │   ├── cache               // Redis 클라이언트
 │   ├── config              // 설정
 │   ├── middleware          // 미들웨어 (Auth, Logger, CORS)
 │   └── pkg                 // 유틸리티 (JWT, Response, Error)
 │
 ├── Makefile
 ├── go.mod
 └── go.sum
```

### 기술 스택
- **Framework**: Gin (빠르고 경량화된 웹 프레임워크)
- **ORM**: GORM (PostgreSQL)
- **Cache**: go-redis
- **Auth**: golang-jwt/jwt   
