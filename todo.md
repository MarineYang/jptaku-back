# 일타쿠 백엔드 TODO

> 목표: 일본어 못하는 오타쿠를 위한 **하루 5문장 + 실시간 대화** MVP 백엔드 완성

---

## 1. 인프라 / 배포

### 완료
- [x] Go 백엔드 프로젝트 초기화 (Gin + GORM)
- [x] PostgreSQL 연결 및 AutoMigrate
- [x] Docker + docker-compose 구성
- [x] Google OAuth 2.0 로그인 구현
- [x] Oracle Cloud 서버 배포 준비 (144.24.80.68)
- [x] DuckDNS 도메인 연결 (jptaku.duckdns.org)
- [x] 방화벽 포트 30001 오픈

### 진행 중
- [ ] Oracle Cloud 서버에서 Docker 빌드 및 실행
- [ ] .env 파일 프로덕션 설정 (GOOGLE_REDIRECT_URL 등)

### 향후 작업
- [ ] HTTPS 설정 (Let's Encrypt)
- [ ] Nginx 리버스 프록시 설정 (선택)
- [ ] Redis 연결 및 활용 (세션/캐시)

---

## 2. API 구현 현황

### Auth (Google OAuth) - 완료
- [x] `GET /api/auth/google` - Google OAuth URL 반환
- [x] `GET /api/auth/google/callback` - OAuth 콜백 (모바일 딥링크 지원)
- [x] `POST /api/auth/refresh` - 토큰 갱신
- [x] `POST /api/auth/logout` - 로그아웃

### User - 완료
- [x] `GET /api/user/me` - 내 정보 조회
- [x] `PUT /api/user/profile` - 프로필 수정
- [x] `POST /api/user/onboarding` - 온보딩 정보 저장
- [x] `GET /api/user/settings` - 설정 조회
- [x] `PUT /api/user/settings` - 설정 수정

### Sentences - 완료
- [x] `GET /api/sentences/today` - 오늘의 5문장 (레벨/관심사 기반 생성)
- [x] `GET /api/sentences/history` - 학습 히스토리 (페이지네이션)

### Learning - 완료
- [x] `POST /api/learning/progress` - 학습 진행 상황 업데이트
- [x] `POST /api/learning/quiz` - 퀴즈 정답 제출 및 검증
- [x] `GET /api/learning/today` - 오늘의 학습 진행 상황
- [x] `GET /api/learning/history` - 학습 히스토리

### Chat - 완료
- [x] `POST /api/chat/session` - 세션 생성
- [x] `GET /api/chat/session/:id` - 세션 조회 (메시지 포함)
- [x] `POST /api/chat/session/:id/end` - 세션 종료
- [x] `GET /api/chat/sessions` - 세션 목록

### Feedback & Stats - Mock 데이터
- [x] `GET /api/feedback/:sessionId` - 피드백 조회
- [ ] `GET /api/stats/today` - 오늘의 통계 (실제 데이터 계산 필요)
- [ ] `GET /api/stats/categories` - 카테고리별 진행도 (실제 데이터 계산 필요)
- [ ] `GET /api/stats/weekly` - 주간 통계 (실제 데이터 계산 필요)

### 미구현 API
- [ ] `GET/POST /api/sentences/bookmarks` - 즐겨찾기 (옵션)
- [ ] `GET /api/meta/interests` - 관심사 목록 (옵션)
- [ ] `GET /api/meta/levels` - 레벨 목록 (옵션)

---

## 3. 실시간 대화 (WebSocket)

### 미구현
- [ ] `/ws/conversation` WebSocket 엔드포인트
- [ ] JWT 검증 & Redis 연결 매핑
- [ ] WebSocket 이벤트 타입 정의
  - [ ] `audio:stream`
  - [ ] `text:transcript`
  - [ ] `session:update`
  - [ ] `control:interrupt`
- [ ] `POST /api/rtc/signaling` (WebRTC SDP/ICE 교환)

---

## 4. OpenAI 연동

### 완료
- [x] OpenAI API 클라이언트 초기화
- [x] 오늘의 5문장 생성 로직 (sentence_service.go)

### 미구현
- [ ] OpenAI Realtime API 연동 PoC
- [ ] WebSocket 브리지 (클라이언트 ↔ Go ↔ OpenAI)
- [ ] AI 피드백 평가 로직 (문법/발음/자연스러움 점수)
- [ ] 시스템 프롬프트에 오늘 5문장 반영

---

## 5. DB 모델 - 완료

- [x] User / UserSettings / UserOnboarding
- [x] Sentence / SentenceDetail / Quiz
- [x] DailySentenceSet / LearningProgress
- [x] ChatSession / ChatMessage
- [x] Feedback / FeedbackHighlight

---

## 6. 프로젝트 구조

```
jptaku-back/
├── cmd/
│   ├── api/main.go           # REST API 서버 진입점
│   ├── migrate/main.go       # DB 마이그레이션 도구
│   └── test/main.go          # 테스트 유틸리티
│
├── internal/
│   ├── api/                  # HTTP 핸들러
│   │   ├── auth/
│   │   ├── user/
│   │   ├── sentences/
│   │   ├── learning/
│   │   ├── chat/
│   │   └── feedback/
│   │
│   ├── service/              # 비즈니스 로직
│   ├── repository/           # DB 접근 (GORM)
│   ├── model/                # 데이터 모델
│   ├── cache/                # Redis 클라이언트
│   ├── config/               # 설정
│   ├── middleware/           # Auth, Logger, CORS
│   └── pkg/                  # 유틸리티 (JWT, OAuth, Response 등)
│
├── docs/                     # Swagger 문서
├── docker-compose.yml
├── env.example
└── go.mod, go.sum
```

---

## 7. 서버 실행 방법

```bash
# 로컬 개발
cp env.example .env
go mod tidy
go run ./cmd/api/main.go

# Docker 실행
docker compose up -d --build

# 프로덕션 서버 (jptaku.duckdns.org:30001)
# .env에 GOOGLE_REDIRECT_URL=http://jptaku.duckdns.org:30001/api/auth/google/callback 설정
docker compose up -d --build
```

---

## 8. 우선순위 작업 목록

### P0 - 즉시 (배포 완료)
- [ ] Oracle Cloud 서버 Docker 실행 완료
- [ ] 프로덕션 .env 설정
- [ ] Google OAuth 테스트

### P1 - 이번 주
- [ ] Stats API 실제 데이터 계산 구현
- [ ] Redis 세션 관리 활용

### P2 - 다음 주
- [ ] WebSocket 실시간 대화 기본 구조
- [ ] OpenAI Realtime API PoC

### P3 - 향후
- [ ] HTTPS 설정
- [ ] AI 피드백 평가 로직
- [ ] 성능 최적화 및 캐싱



같은 덕후끼리 실제로 나눌 법한 대화를 일본어로 얼마나 깊게 이어가느냐


🔹 게임

Steam API / 랭킹 페이지
	•	Top Sellers
	•	Most Played
	•	Recent Reviews 많은 게임  
자동 수집 → 중복 제거 → 상위 N개
{ "id": "elden_ring", "title": "Elden Ring", "domain": "game" }

🔹 애니

AniList GraphQL API (정답)
	•	Popularity 순
	•	이번 시즌
	•	All-time 인기
{ "id": "chainsaw_man", "title": "チェンソーマン", "domain": "anime" }

🔹 음악 (JPOP)

Spotify API
	•	Artist 단위
	•	인기 상위만
{ "id": "yoasobi", "title": "YOASOBI", "domain": "music" }

🔹 이벤트

공식 사이트 / 위키 크롤링
	•	“행사명”만
	•	날짜, 장소 ❌
{ "id": "comiket_104", "title": "コミックマーケット104", "domain": "event" }

## todo
	•	AniList API로 애니 Top 50 가져오기
	•	Steam 인기 게임 Top 50 가져오기
	•	raw_topics.json 생성

  데이터 수집 · 생성 파이프라인 기능 목록 (최종)
	1.	통합 콘텐츠 수집기 – 애니·게임·음악·영화/드라마 인기 작품 제목을 외부 API에서 자동 수집
	2.	콘텐츠 정규화 모듈 – 수집된 제목을 id / title / domain 형식으로 통일 변환
	3.	Macro Topic 생성기 – 작품 제목을 기반으로 LLM이 대화 시작 질문과 훅을 생성
	4.	Macro Topic 저장기 – 생성된 Topic을 JSON 파일로 저장해 서비스에서 바로 사용
	5.	Micro Topic 추론기 – 대화 중 유저 발화를 분석해 세부 초점과 대화 유형을 실시간 추론
	6.	Micro Topic 로그 수집기 – 추론된 Micro Topic을 분석용으로 기록(선택)