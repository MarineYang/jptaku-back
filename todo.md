# 일타쿠 백엔드 TODO

> 목표: 일본어 못하는 오타쿠를 위한 **하루 5문장 + AI 대화** MVP 완성

---

## 완료된 기능

- [x] Go 백엔드 (Gin + GORM + PostgreSQL + Redis)
- [x] Docker Compose 배포 환경
- [x] Google OAuth 로그인 (웹 + 모바일 ID Token)
- [x] JWT 인증 + Refresh Token
- [x] 유저 온보딩 (레벨, 관심 카테고리)
- [x] 오늘의 5문장 선별 (레벨/카테고리 기반, 학습 이력 반영)
- [x] 문장 학습 진행도 (이해/말하기/확인/암기)
- [x] 퀴즈 검증
- [x] 플래시카드 SRS (bad→10분, mid→1시간, good→24시간)
- [x] AI 대화 세션 (OpenAI 스트리밍 + SSE)
- [x] AI 대화 중 한국어 번역 + 제안 3개
- [x] VoiceVox TTS (AI 응답 음성 생성)
- [x] NCP Object Storage 오디오 서빙
- [x] Swagger 문서

---

## 남은 작업

### P0 - 핵심 기능 (MVP 완성)

#### 1. AI 대화 페르소나 입히기
- [x] 시스템 프롬프트에 오타쿠 캐릭터 페르소나 적용
- [x] 토픽별 대화 스타일 분기 (애니/게임/음악 등)
- [x] 유저 레벨에 맞는 일본어 난이도 조절

#### 2. 데이터 동기화 아키텍처 (Cron 레포 연동)
- [ ] `POST /api/admin/sync/sentences` — NCP JSON → sentences 테이블 upsert
- [ ] `POST /api/admin/sync/topics` — NCP JSON → topics 테이블 upsert
- [ ] topics 테이블 모델 + 마이그레이션 (macro/micro topic)
- [ ] 관리자 인증 미들웨어 (Cron 서버만 호출 가능하도록)

> 데이터 흐름: Cron 레포(별도)에서 문장/토픽 생성 → NCP JSON 업로드 → 백엔드 sync API 호출 → DB 저장

#### 3. 실제 Sentence 데이터 구축
- [x] mock 데이터 → Cron 레포에서 생성된 실데이터로 교체
- [x] 카테고리별 문장 확보 (애니/게임/음악/영화/드라마)
- [x] JLPT N5~N3 레벨별 문장 분류

#### 4. Sentence 1일차부터 순차 제공
- [x] 학습 일차 기반 문장 제공 로직 구현
- [ ] 난이도 점진적 상승 커리큘럼 설계

#### 5. Sentence 오디오 데이터 제공 (실제 문장 시드 확보 후)
- [ ] 실시간 생성 + 캐싱 방식 구현 (요청 시 VoiceVox 생성 → NCP 업로드 → URL 캐싱)
- [ ] Sentence.AudioURL 실데이터 연결

#### 6. 학습 N일차 실제 카운팅
- [x] UserStudyStats 모델 서비스 레이어 구현
- [x] 일일 학습 완료 판정 로직 (5문장 중 N개 완료 시)
- [ ] 연속 학습일(스트릭) 계산
- [x] Stats API 실제 데이터 반영 (`GET /api/stats/today`)

---

### P1 - 대화 품질 강화

#### 7. AI 대화 STT 지원
- [ ] STT 서비스 선정 (Google Cloud Speech / Whisper 등)
- [ ] 음성 입력 엔드포인트 구현
- [ ] 클라이언트 음성 → 텍스트 변환 → 기존 메시지 흐름에 합류

#### 8. AI 대화 피드백 기능
- [ ] 대화 종료 시 OpenAI로 피드백 자동 생성
- [ ] 문법/자연스러움 점수 산출
- [ ] 잘한 표현 하이라이트 추출
- [ ] 피드백 저장 + 조회 API 연결

#### 9. 피드백 데이터 제공 방식 기획
- [ ] 대화 직후 요약 피드백 vs 상세 리포트
- [ ] 점수 시각화 데이터 구조 확정
- [ ] 주간/월간 성장 추이 제공 여부

---

### P2 - 복습 & 확장

#### 10. 대화 데이터 + 문장 복습 기획
- [ ] 이전 대화에서 사용한 표현 복습
- [ ] 틀렸던 문장 재출제 로직
- [ ] 플래시카드와 대화 데이터 연계

#### 10. 콘텐츠 데이터 수집 파이프라인
- [x] AniList API → 애니 인기작 Top 50
- [x] Steam API → 인기 게임 Top 50
- [x] Spotify API → JPOP 아티스트
- [x] 수집 데이터 → 대화 토픽 자동 생성

#### 11. 인프라
- [ ] HTTPS 설정 (Let's Encrypt)
- [ ] 프로덕션 배포 자동화
- [ ] 주간 통계 API 구현 (`GET /api/stats/weekly`)

---

## 2/19 작업 목록

### 1순위: 학습 완료 판정
> 완료 기준: 5문장 모두 Memorized = true
> 미완료 이월: Memorized = false인 문장은 다음날 새 5문장 아래에 추가 제공

- [x] DailySentenceSet에 완료 여부 필드 추가 또는 LearningProgress 기반 판정
- [x] 오늘의 문장 조회 시 미완료 이월 문장 포함 로직
- [x] 학습 완료 API 또는 자동 판정 (5문장 Memorized 체크)

### 2순위: 문장 순차 제공
> 현재 랜덤 → 레벨 오름차순 + 학습한 문장 제외로 변경

- [x] FindRandom → FindSequential로 변경 (N5 → N4 → N3 순서)
- [x] 이미 Memorized된 문장 제외 로직

### 3순위: AI 대화 페르소나
- [x] Redis에서 도메인별 topic JSON 로드 로직 (TopicService)
- [x] 랜덤 캐릭터 생성 (이름/성별/말투) Go 이식 (persona.go)
- [x] ChatSession 모델에 Persona/ContentID 필드 추가
- [x] CreateSession → 도메인 선택 → Redis 작품 랜덤 → 페르소나 생성
- [x] 시스템 프롬프트에 페르소나+작품정보+오늘의 문장 주입
- [x] suggestion에 오늘의 문장 마킹 포함

### 추후 작업 (데이터 확보 후)
- [ ] Day 1~Day X 커리큘럼 설계
- [ ] 목표: 총 5000~6000문장으로 일본어 완성 (현재 3743문장, 추가 확보 필요)
- [ ] 문장 TTS 오디오: 실시간 생성 + 캐싱 (VoiceVox → NCP → URL)

---

## 2/18 개발 내용 (AI 대화 품질 개선)

### 완료된 작업

#### 시나리오 시스템 구축 (persona.go)
- [x] `ScenarioEntry` struct 도입 — `Text`(일본어), `TextKr`(한국어), `UserRole`(유저 역할) 3필드
- [x] 도메인별(anime/game/drama/movie/music) 4~5개 구체적 상황 시나리오 풀 정의
  - 코미케 부스, 애니 숍, 방과 후 교실, 팬 이벤트, 애니 카페 등
  - 각 시나리오마다 AI 입장 + 유저 입장 명시
- [x] `selectScenario()` — `sessionID % 시나리오수`로 결정 (같은 세션 = 항상 같은 시나리오)
- [x] `buildPersonaSystemPrompt()`에 `【今回の状況】` + `【ユーザーの立場】` 섹션 추가
- [x] 시나리오 한국어 번역 — `scenario_text_kr`로 API 응답에 포함 (클라이언트 화면 상단 표시용)

#### 유저 레벨 연동 (service.go + persona.go)
- [x] `UserRepository` 인터페이스 추가 — `GetOnboarding()` 으로 레벨 조회
- [x] `getUserLevel()` 헬퍼 (5=N5 초급, 4=N4 중급, 3=N3 중상급, 기본값 5)
- [x] 레벨별 언어 설정:
  - N5: 1문장 15자 이내, 어려운 한자 히라가나 병기, 기본 어휘만
  - N4: 1~2문장, 기본 한자 OK, ~ている 등 기본 문형
  - N3: 2~3문장, 자연스러운 일본어, 다양한 문형

#### 오늘의 학습 문장 연동 개선
- [x] AI 자신은 오늘의 문장 직접 사용 금지 → 유저가 쓰고 싶어지도록 유도만
- [x] 제안 3개 중 문맥에 맞을 때만 오늘의 문장 포함 (강제 아님)
- [x] `is_today_sentence: true` 플래그로 클라이언트에 오늘의 문장 여부 전달

#### 제안 생성 품질 개선 (service.go)
- [x] 대화 히스토리 최근 6개를 컨텍스트로 포함
- [x] AI 질문 유형에 맞는 고유명사 직접 사용 지시:
  - 작품명 질문 → 실제 작품명 (チェンソーマンが好き！)
  - 캐릭터 질문 → 실제 캐릭터명 (デンジが好き！)
  - 곡 질문 → 실제 곡명
- [x] 제안 유형: 직접 답변 / 의견·감정 / 되묻기 3가지로 구분
- [x] `levelSuggestionHint()`로 레벨별 제안 길이·난이도 조정

#### API 응답 필드 추가 (dto.go)
- [x] `CreateSessionResponse.ScenarioTextKr` — 대화 상황 한국어 설명 (화면 상단 표시용)

#### 인사 프롬프트 개선 (service.go)
- [x] 첫 인사를 시나리오 상황 속 자연스러운 한 마디로 생성
- [x] "〜について話しましょう" 같은 기계적 표현 금지 명시
- [x] 상황에서 자연스럽게 나오는 첫 마디 생성 지시 (예: "あ、それ…！")

---

### 미해결 / 다음 작업

#### AI 대화 자연스러움 개선 (진행 중)
- [ ] **대화가 아직 자연스럽지 않음** — 원인 분석 및 프롬프트 추가 개선 필요
  - 현상: AI 응답이 상황에 완전히 몰입하지 못하고 어색한 경우 있음
  - 제안이 AI 질문과 맥락이 완전히 맞지 않는 경우 있음
  - 추가 검토 필요: 시나리오 다양성, 페르소나 일관성, 대화 흐름

#### Redis 토픽 데이터 문제
- [ ] `contentTitle`이 "anime" 등 도메인명으로 폴백되는 현상
  - 원인: Redis에 토픽 데이터가 로드되지 않은 경우 폴백
  - 해결: `go run ./cmd/seed-topics/main.go` 실행으로 Redis에 토픽 시드 필요
  - 토픽 데이터 없을 경우 기본 처리 방식 개선 검토

#### 대화 내용 저장
- [ ] 대화 종료 후 대화 내용 저장 방식 확정 필요
  - 현재: 메시지는 `chat_messages` 테이블에 턴마다 저장됨
  - 검토: 복습용 대화 요약 저장 / 좋은 표현 추출 저장 / 학습 이력 연계
