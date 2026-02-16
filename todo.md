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
- [ ] 시스템 프롬프트에 오타쿠 캐릭터 페르소나 적용
- [ ] 토픽별 대화 스타일 분기 (애니/게임/음악 등)
- [ ] 유저 레벨에 맞는 일본어 난이도 조절

#### 2. 데이터 동기화 아키텍처 (Cron 레포 연동)
- [ ] `POST /api/admin/sync/sentences` — NCP JSON → sentences 테이블 upsert
- [ ] `POST /api/admin/sync/topics` — NCP JSON → topics 테이블 upsert
- [ ] topics 테이블 모델 + 마이그레이션 (macro/micro topic)
- [ ] 관리자 인증 미들웨어 (Cron 서버만 호출 가능하도록)

> 데이터 흐름: Cron 레포(별도)에서 문장/토픽 생성 → NCP JSON 업로드 → 백엔드 sync API 호출 → DB 저장

#### 3. 실제 Sentence 데이터 구축
- [ ] mock 데이터 → Cron 레포에서 생성된 실데이터로 교체
- [ ] 카테고리별 문장 확보 (애니/게임/음악/영화/드라마)
- [ ] JLPT N5~N3 레벨별 문장 분류

#### 4. Sentence 1일차부터 순차 제공
- [ ] 학습 일차 기반 문장 제공 로직 구현
- [ ] 난이도 점진적 상승 커리큘럼 설계

#### 5. Sentence 오디오 데이터 제공
- [ ] 문장별 일본어 음성 생성 (VoiceVox 또는 외부 TTS)
- [ ] NCP Object Storage에 업로드
- [ ] Sentence.AudioURL 실데이터 연결

#### 6. 학습 N일차 실제 카운팅
- [ ] UserStudyStats 모델 서비스 레이어 구현
- [ ] 일일 학습 완료 판정 로직 (5문장 중 N개 완료 시)
- [ ] 연속 학습일(스트릭) 계산
- [ ] Stats API 실제 데이터 반영 (`GET /api/stats/today`)

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
