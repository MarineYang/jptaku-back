# JPTaku Backend Server

일본어 학습 서비스 **JPTaku**의 백엔드 API 서버입니다.

## 기술 스택

| 분류 | 기술 |
|------|------|
| **Language** | Go 1.24 |
| **Framework** | Gin |
| **Database** | PostgreSQL |
| **Cache** | Redis |
| **ORM** | GORM |
| **Auth** | JWT + Google OAuth 2.0 (ID Token) |
| **AI** | OpenAI GPT |
| **Storage** | NCP Object Storage (S3 호환) |
| **TTS** | VoiceVox |

## 프로젝트 구조

```
jptaku-back/
├── cmd/
│   ├── api/
│   │   └── main.go              # API 서버 진입점
│   ├── cron/
│   │   └── main.go              # 문장 생성 크론 작업
│   └── seed-topics/
│       └── main.go              # AI 대화 토픽 시드 데이터 등록
│
├── internal/
│   ├── app/                     # 애플리케이션 초기화
│   │   ├── app.go              # App 구조체, Run(), Shutdown()
│   │   ├── database.go         # DB 초기화, 마이그레이션
│   │   ├── router.go           # Gin 라우터 설정
│   │   └── services.go         # 의존성 조립 (repos, services, infra)
│   │
│   ├── api/                     # API 핸들러 (Controller)
│   │   ├── auth/
│   │   ├── audio/
│   │   ├── chat/
│   │   ├── feedback/
│   │   ├── flash/
│   │   ├── learning/
│   │   ├── sentences/
│   │   └── user/
│   │
│   ├── config/                  # 설정 관리
│   │   ├── config.go
│   │   └── database.go
│   │
│   ├── middleware/              # HTTP 미들웨어
│   │   ├── auth.go
│   │   ├── cors.go
│   │   └── logger.go
│   │
│   ├── model/                   # 데이터 모델 (Entity)
│   │   ├── user.go
│   │   ├── sentence.go
│   │   ├── learning.go
│   │   ├── chat.go
│   │   ├── persona.go
│   │   └── feedback.go
│   │
│   ├── pkg/                     # 유틸리티
│   │   ├── jwt.go
│   │   ├── google_oauth.go
│   │   ├── response.go
│   │   ├── error.go
│   │   └── categories.go
│   │
│   ├── repository/              # 데이터 접근 레이어
│   │   ├── db_manager.go
│   │   ├── user_repo.go
│   │   ├── sentence_repo.go
│   │   ├── learning_repo.go
│   │   ├── chat_repo.go
│   │   └── feedback_repo.go
│   │
│   └── service/                 # 비즈니스 로직 레이어
│       ├── auth/
│       │   ├── interface.go    # Provider 인터페이스
│       │   ├── dto.go          # 입출력 DTO
│       │   └── service.go      # 비즈니스 로직
│       ├── user/
│       ├── sentence/
│       ├── learning/
│       ├── flash/
│       ├── chat/
│       ├── topic/
│       └── feedback/
│
├── docs/                        # Swagger 문서
├── docker-compose.yml
├── Dockerfile
└── Makefile
```

## 아키텍처

```
┌─────────────────────────────────────────────────────────────┐
│                      HTTP Request                           │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   Middleware Layer                          │
│                (Logger → CORS → Auth)                       │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Handler Layer                            │
│     (auth, user, sentences, learning, chat, flash ...)      │
│               - 요청 파싱 / 응답 포맷팅                        │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Service Layer                            │
│          - 비즈니스 로직 (인터페이스 기반)                      │
│          - interface.go / dto.go / service.go               │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   Repository Layer                          │
│               - 데이터베이스 접근                              │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│        PostgreSQL / Redis / NCP Object Storage              │
└─────────────────────────────────────────────────────────────┘
```

## API 엔드포인트

### Health Check
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | 서버 상태 확인 |

### Auth - `/api/auth`
| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| POST | `/guest` | 비회원 토큰 발급 | - |
| POST | `/google/token` | Google ID Token 로그인 (네이티브 앱) | - |
| POST | `/refresh` | 토큰 갱신 (회원 전용) | - |
| POST | `/logout` | 로그아웃 | - |

### User - `/api/user`
| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| GET | `/me` | 내 정보 조회 | O |
| PUT | `/profile` | 프로필 수정 | O |
| POST | `/onboarding` | 온보딩 저장 | O |
| GET | `/settings` | 설정 조회 | O |
| PUT | `/settings` | 설정 수정 | O |

### Sentences - `/api/sentences`
| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| GET | `/today` | 오늘의 문장 조회 | 회원 |
| GET | `/guest` | 비회원 오늘의 문장 (N5 랜덤 5문장) | 비회원 |
| GET | `/history` | 학습 히스토리 조회 | 회원 |

### Flash Cards - `/api/flash`
| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| GET | `/` | 플래시카드 조회 | 회원 |
| GET | `/guest` | 비회원 플래시카드 (N5 랜덤 5문장) | 비회원 |

### Chat - `/api/chat`
| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| POST | `/session` | 대화 세션 생성 | 회원 |
| GET | `/session/:id` | 세션 상세 조회 | 회원 |
| POST | `/session/:id/message` | 메시지 전송 (SSE 스트리밍) | 회원 |
| POST | `/session/:id/end` | 세션 종료 | 회원 |
| GET | `/sessions` | 세션 목록 조회 | 회원 |
| POST | `/guest/start` | 비회원 대화 세션 시작 | 비회원 |
| POST | `/guest/message` | 비회원 메시지 전송 (SSE 스트리밍) | 비회원 |

### Audio - `/api/audio`
| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| GET | `/:filename` | 음성 파일 스트리밍 | - |

### Feedback & Stats
| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| GET | `/feedback/:sessionId` | 세션별 피드백 조회 | O |
| GET | `/stats/today` | 오늘의 통계 | O |
| GET | `/stats/categories` | 카테고리별 진행도 | O |
| GET | `/stats/weekly` | 주간 통계 | O |

## 시작하기

### 요구사항
- Go 1.24+
- PostgreSQL 15+
- Redis 7+
- VoiceVox (TTS 서버, 로컬 실행)
- OpenAI API 키
- NCP Object Storage 버킷
- Google Cloud 프로젝트 (OAuth Client ID)

### 환경 설정

```bash
cp .env.example .env
# .env 파일을 열어 실제 값으로 수정
```

### 환경 변수

| 변수 | 설명 | 예시 |
|------|------|------|
| `SERVER_PORT` | 서버 포트 | `30001` |
| `GIN_MODE` | Gin 모드 | `debug` / `release` |
| `DB_HOST` | PostgreSQL 호스트 | `localhost` |
| `DB_PORT` | PostgreSQL 포트 | `5432` |
| `DB_USER` | DB 사용자명 | `postgres` |
| `DB_PASSWORD` | DB 비밀번호 | `postgres` |
| `DB_NAME` | DB 이름 | `jptaku` |
| `REDIS_HOST` | Redis 호스트 | `localhost` |
| `REDIS_PORT` | Redis 포트 | `6379` |
| `REDIS_PASSWORD` | Redis 비밀번호 | (없으면 빈 값) |
| `JWT_SECRET` | JWT 서명 키 | 임의의 긴 문자열 |
| `JWT_EXPIRATION_HOURS` | Access Token 만료 시간 | `24` |
| `OPEN_AI_API_KEY` | OpenAI API 키 | `sk-proj-...` |
| `GOOGLE_CLIENT_ID` | Google OAuth Client ID | `xxx.apps.googleusercontent.com` |
| `GOOGLE_CLIENT_SECRET` | Google OAuth Client Secret | `GOCSPX-...` |
| `VOICEVOX_URL` | VoiceVox 서버 주소 | `http://localhost:50021` |
| `NCP_ENDPOINT` | NCP 스토리지 엔드포인트 | `https://kr.object.ncloudstorage.com` |
| `NCP_ACCESS_KEY` | NCP IAM Access Key | `ncp_iam_...` |
| `NCP_SECRET_KEY` | NCP IAM Secret Key | `ncp_iam_...` |
| `NCP_BUCKET_NAME` | NCP 버킷 이름 | `jptaku` |

### 실행

```bash
# 의존성 설치
go mod tidy

# 개발 서버 실행
go run cmd/api/main.go

# 빌드
go build -o server cmd/api/main.go

# Docker
docker-compose up -d

# AI 대화 토픽 시드 데이터 등록 (최초 1회)
go run cmd/seed-topics/main.go
```

## 인증 방식

JWT 기반 인증. 로그인은 **Google OAuth (네이티브 앱 ID Token)** 또는 **비회원 토큰** 발급 방식을 사용합니다.

### 비회원 로그인
```
POST /api/auth/guest
→ { "access_token": "...", "is_guest": true }
```
- 유효 기간: 24시간
- Refresh Token 없음 (만료 시 재발급)
- 데이터 저장 없음

### 회원 로그인 (Google - 네이티브 앱)
```
POST /api/auth/google/token
Body: { "id_token": "<Google SDK에서 받은 ID Token>" }
→ { "access_token": "...", "refresh_token": "...", "user": { ... } }
```

### 토큰 갱신 (회원 전용)
```
POST /api/auth/refresh
Body: { "refresh_token": "..." }
→ { "access_token": "...", "refresh_token": "..." }
```

### API 인증 헤더
```
Authorization: Bearer <access_token>
```

## 비회원 vs 회원 기능 비교

| 기능 | 비회원 | 회원 |
|------|--------|------|
| 오늘의 문장 | N5 랜덤 5문장 | 레벨별 맞춤 5문장 (매일 고정) |
| 플래시카드 | N5 랜덤 5문장 | 학습 히스토리 기반 |
| AI 대화 | 세션 저장 없음 (무제한 도메인) | 세션/메시지 저장, 피드백 제공 |
| 학습 기록 | X | O |
| 통계 | X | O |

## 응답 형식

### 성공
```json
{
  "success": true,
  "data": { ... }
}
```

### 페이지네이션
```json
{
  "success": true,
  "data": [ ... ],
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 100
  }
}
```

### 에러
```json
{
  "success": false,
  "error": {
    "message": "에러 메시지"
  }
}
```

## License

License by MarineYang
