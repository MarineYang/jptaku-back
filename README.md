# JPTaku Backend Server

일본어 학습 서비스 **JPTaku**의 백엔드 API 서버입니다.

## 기술 스택

| 분류 | 기술 |
|------|------|
| **Language** | Go 1.24 |
| **Framework** | Gin |
| **Database** | PostgreSQL |
| **ORM** | GORM |
| **Auth** | JWT + Google OAuth 2.0 |
| **Storage** | NCP Object Storage (S3 호환) |
| **TTS** | VoiceVox |

## 프로젝트 구조

```
jptaku-back/
├── cmd/
│   ├── api/
│   │   └── main.go              # API 서버 진입점
│   └── cron/
│       └── main.go              # 문장 생성 크론 작업
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
│       ├── async_service.go
│       ├── auth/
│       │   ├── interface.go    # Provider 인터페이스
│       │   ├── dto.go          # 입출력 DTO
│       │   └── service.go      # 비즈니스 로직
│       ├── user/
│       ├── sentence/
│       ├── learning/
│       ├── chat/
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
│                      HTTP Request                            │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   Middleware Layer                           │
│                (Logger → CORS → Auth)                        │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Handler Layer                             │
│       (auth, user, sentences, learning, chat, feedback)      │
│               - 요청 파싱 / 응답 포맷팅                        │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Service Layer                             │
│          - 비즈니스 로직 (인터페이스 기반)                     │
│          - interface.go / dto.go / service.go                │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   Repository Layer                           │
│               - 데이터베이스 접근                             │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│             PostgreSQL / NCP Object Storage                  │
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
| GET | `/google` | Google OAuth URL 반환 | - |
| GET | `/google/callback` | Google OAuth 콜백 | - |
| POST | `/refresh` | 토큰 갱신 | - |
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
| GET | `/today` | 오늘의 5문장 조회 | O |
| GET | `/history` | 학습 히스토리 조회 | O |

### Learning - `/api/learning`
| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| POST | `/progress` | 학습 진행 상황 업데이트 | O |
| POST | `/quiz` | 퀴즈 정답 제출 | O |
| GET | `/today` | 오늘의 학습 진행 상황 | O |
| GET | `/history` | 학습 히스토리 | O |

### Chat - `/api/chat`
| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| POST | `/session` | 대화 세션 생성 | O |
| GET | `/session/:id` | 세션 상세 조회 | O |
| POST | `/session/:id/end` | 세션 종료 | O |
| GET | `/sessions` | 세션 목록 조회 | O |

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
- Docker (선택)

### 환경 설정

```bash
cp env.example .env
```

### 환경 변수

```env
# Server
SERVER_PORT=30001
GIN_MODE=debug

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=jptaku
DB_SSLMODE=disable

# JWT
JWT_SECRET=your-secret-key
JWT_EXPIRATION_HOURS=24

# Google OAuth
GOOGLE_CLIENT_ID=xxx
GOOGLE_CLIENT_SECRET=xxx
GOOGLE_REDIRECT_URL=http://localhost:30001/api/auth/google/callback

# NCP Object Storage
NCP_ACCESS_KEY=xxx
NCP_SECRET_KEY=xxx
NCP_ENDPOINT=https://kr.object.ncloudstorage.com
NCP_BUCKET=jptaku

# VoiceVox (for cron)
VOICEVOX_URL=http://localhost:50021
```

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
```

## 인증 방식

Google OAuth 2.0 + JWT 기반 인증

### 로그인 플로우
1. `GET /api/auth/google?state=mobile` 호출
2. 반환된 URL로 Google 로그인
3. 콜백 처리
   - **웹**: JSON으로 토큰 반환
   - **모바일**: `jptaku://auth/callback?access_token=xxx` 딥링크

### API 인증
```
Authorization: Bearer <access_token>
```

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
