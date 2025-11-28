-- ============================================================================
-- 일타쿠(JPTAKU) Database Schema
-- PostgreSQL 15+
-- ============================================================================

-- ============================================================================
-- 1. MASTER TABLES (마스터 테이블)
-- ============================================================================

-- 레벨 마스터
CREATE TABLE levels (
    id SERIAL PRIMARY KEY,
    code VARCHAR(20) NOT NULL UNIQUE,          -- 'N5', 'N4', 'N3', 'beginner', etc.
    name_kr VARCHAR(50) NOT NULL,              -- '초급', '중급'
    description TEXT,
    sort_order SMALLINT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 카테고리 마스터 (오타쿠 카테고리)
CREATE TABLE categories (
    id SERIAL PRIMARY KEY,
    code VARCHAR(30) NOT NULL UNIQUE,          -- 'anime', 'game', 'gacha', 'seichi', etc.
    name_kr VARCHAR(50) NOT NULL,              -- '애니메이션', '게임'
    sort_order SMALLINT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 태그 마스터
CREATE TABLE tags (
    id SERIAL PRIMARY KEY,
    code VARCHAR(30) NOT NULL UNIQUE,          -- 'greeting', 'shopping', 'emotion', etc.
    name_kr VARCHAR(50) NOT NULL,
    sort_order SMALLINT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 퀴즈 유형 마스터
CREATE TABLE quiz_types (
    id SERIAL PRIMARY KEY,
    code VARCHAR(30) NOT NULL UNIQUE,          -- 'meaning', 'fill_blank', 'listening', 'ordering'
    name_kr VARCHAR(50) NOT NULL,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- 2. USER TABLES (사용자 테이블)
-- ============================================================================

-- 사용자
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(100) NOT NULL,
    profile_image_url VARCHAR(500),
    status VARCHAR(20) NOT NULL DEFAULT 'active',  -- 'active', 'inactive', 'suspended', 'deleted'
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 사용자 설정
CREATE TABLE user_settings (
    user_id BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
    notification_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    daily_reminder_time TIME DEFAULT '09:00:00',
    preferred_voice_speed DECIMAL(2,1) NOT NULL DEFAULT 1.0 CHECK (preferred_voice_speed BETWEEN 0.5 AND 2.0),
    show_romaji BOOLEAN NOT NULL DEFAULT TRUE,
    show_translation BOOLEAN NOT NULL DEFAULT TRUE,
    theme VARCHAR(20) NOT NULL DEFAULT 'light',     -- 'light', 'dark', 'system'
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 사용자 온보딩 정보
CREATE TABLE user_onboarding (
    user_id BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
    level_id INTEGER REFERENCES levels(id) ON DELETE SET NULL ON UPDATE CASCADE,
    purposes JSONB NOT NULL DEFAULT '[]',           -- ['travel', 'work', 'hobby']
    is_completed BOOLEAN NOT NULL DEFAULT FALSE,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 사용자 관심 카테고리 (N:N)
CREATE TABLE user_interests (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
    category_id INTEGER NOT NULL REFERENCES categories(id) ON DELETE CASCADE ON UPDATE CASCADE,
    priority SMALLINT NOT NULL DEFAULT 0,           -- 관심도 우선순위
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, category_id)
);

-- ============================================================================
-- 3. CONTENT TABLES (콘텐츠 테이블)
-- ============================================================================

-- 문장 Seed Pool
CREATE TABLE sentences (
    id BIGSERIAL PRIMARY KEY,
    jp_text TEXT NOT NULL,                          -- 일본어 원문
    kr_text TEXT NOT NULL,                          -- 한국어 번역
    romaji TEXT,                                    -- 로마자 표기
    pronunciation_guide TEXT,                       -- 발음 가이드
    level_id INTEGER NOT NULL REFERENCES levels(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    category_id INTEGER NOT NULL REFERENCES categories(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    difficulty_score SMALLINT DEFAULT 50 CHECK (difficulty_score BETWEEN 1 AND 100),
    usage_count INTEGER NOT NULL DEFAULT 0,         -- 추천 횟수 (통계용)
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 문장-태그 매핑 (N:N)
CREATE TABLE sentence_tags (
    sentence_id BIGINT NOT NULL REFERENCES sentences(id) ON DELETE CASCADE ON UPDATE CASCADE,
    tag_id INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE ON UPDATE CASCADE,
    PRIMARY KEY (sentence_id, tag_id)
);

-- 문장별 퀴즈 (LLM 생성 캐시)
CREATE TABLE sentence_quizzes (
    id BIGSERIAL PRIMARY KEY,
    sentence_id BIGINT NOT NULL REFERENCES sentences(id) ON DELETE CASCADE ON UPDATE CASCADE,
    quiz_type_id INTEGER NOT NULL REFERENCES quiz_types(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    question_jp TEXT NOT NULL,                      -- 일본어 문제
    question_kr TEXT,                               -- 한국어 문제 (optional)
    options JSONB NOT NULL,                         -- ["선택지1", "선택지2", "선택지3", "선택지4"]
    correct_answer VARCHAR(255) NOT NULL,           -- 정답
    explanation_jp TEXT,                            -- 일본어 해설
    explanation_kr TEXT,                            -- 한국어 해설
    generated_by VARCHAR(50) DEFAULT 'llm',         -- 'llm', 'manual'
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (sentence_id, quiz_type_id)              -- 문장당 퀴즈 유형별 1개
);

-- 문장 상세 정보 (문법, 어휘 분석 캐시)
CREATE TABLE sentence_details (
    sentence_id BIGINT PRIMARY KEY REFERENCES sentences(id) ON DELETE CASCADE ON UPDATE CASCADE,
    grammar_points JSONB DEFAULT '[]',              -- [{"pattern": "~ている", "meaning": "진행/상태"}]
    vocabulary JSONB DEFAULT '[]',                  -- [{"word": "食べる", "reading": "たべる", "meaning": "먹다"}]
    cultural_notes TEXT,                            -- 문화적 배경 설명
    usage_examples JSONB DEFAULT '[]',              -- 추가 예문
    generated_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- 4. LEARNING TABLES (학습 테이블)
-- ============================================================================

-- 일일 문장 세트
CREATE TABLE daily_sentence_sets (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
    assigned_date DATE NOT NULL,
    is_completed BOOLEAN NOT NULL DEFAULT FALSE,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, assigned_date)                 -- 유저별 일일 1세트
);

-- 일일 문장 세트 아이템 (5문장)
CREATE TABLE daily_sentence_set_items (
    set_id BIGINT NOT NULL REFERENCES daily_sentence_sets(id) ON DELETE CASCADE ON UPDATE CASCADE,
    sentence_id BIGINT NOT NULL REFERENCES sentences(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    display_order SMALLINT NOT NULL CHECK (display_order BETWEEN 1 AND 5),
    PRIMARY KEY (set_id, sentence_id),
    UNIQUE (set_id, display_order)                  -- 순서 중복 방지
);

-- 학습 진행도 (문장별)
CREATE TABLE learning_progress (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
    sentence_id BIGINT NOT NULL REFERENCES sentences(id) ON DELETE CASCADE ON UPDATE CASCADE,
    daily_set_id BIGINT REFERENCES daily_sentence_sets(id) ON DELETE SET NULL ON UPDATE CASCADE,

    -- 4단계 학습 상태
    step_understand BOOLEAN NOT NULL DEFAULT FALSE,  -- 이해하기
    step_speak BOOLEAN NOT NULL DEFAULT FALSE,       -- 말하기
    step_confirm BOOLEAN NOT NULL DEFAULT FALSE,     -- 확인하기
    step_memorize BOOLEAN NOT NULL DEFAULT FALSE,    -- 암기 완료

    -- 학습 통계
    time_spent_seconds INTEGER DEFAULT 0,            -- 총 학습 시간
    repeat_count SMALLINT DEFAULT 0,                 -- 반복 학습 횟수
    comprehension_level SMALLINT DEFAULT 0 CHECK (comprehension_level BETWEEN 0 AND 5),
    notes TEXT,                                      -- 개인 메모

    is_completed BOOLEAN NOT NULL DEFAULT FALSE,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, sentence_id, daily_set_id)     -- 세트별 유니크
);

-- 퀴즈 시도 기록
CREATE TABLE quiz_attempts (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
    quiz_id BIGINT NOT NULL REFERENCES sentence_quizzes(id) ON DELETE CASCADE ON UPDATE CASCADE,
    selected_answer VARCHAR(255) NOT NULL,
    is_correct BOOLEAN NOT NULL,
    time_spent_seconds SMALLINT,                    -- 풀이 소요 시간
    attempted_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- 5. CHAT TABLES (채팅 테이블)
-- ============================================================================

-- 채팅 세션
CREATE TABLE chat_sessions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
    daily_set_id BIGINT REFERENCES daily_sentence_sets(id) ON DELETE SET NULL ON UPDATE CASCADE,
    title VARCHAR(255),                              -- 세션 제목 (자동 생성)
    status VARCHAR(20) NOT NULL DEFAULT 'active',    -- 'active', 'completed', 'abandoned'
    sentences_used_count SMALLINT NOT NULL DEFAULT 0 CHECK (sentences_used_count BETWEEN 0 AND 5),
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 채팅 메시지
CREATE TABLE chat_messages (
    id BIGSERIAL PRIMARY KEY,
    session_id BIGINT NOT NULL REFERENCES chat_sessions(id) ON DELETE CASCADE ON UPDATE CASCADE,
    speaker VARCHAR(10) NOT NULL CHECK (speaker IN ('user', 'ai', 'system')),
    jp_text TEXT,                                    -- 일본어 텍스트
    kr_text TEXT,                                    -- 한국어 텍스트
    sentence_id BIGINT REFERENCES sentences(id) ON DELETE SET NULL ON UPDATE CASCADE,  -- 사용된 문장 (nullable)
    metadata JSONB DEFAULT '{}',                     -- 추가 메타데이터 (발음 점수 등)
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 세션 피드백
CREATE TABLE session_feedbacks (
    id BIGSERIAL PRIMARY KEY,
    session_id BIGINT NOT NULL UNIQUE REFERENCES chat_sessions(id) ON DELETE CASCADE ON UPDATE CASCADE,

    -- 점수 (0-100)
    grammar_score SMALLINT CHECK (grammar_score BETWEEN 0 AND 100),
    pronunciation_score SMALLINT CHECK (pronunciation_score BETWEEN 0 AND 100),
    naturalness_score SMALLINT CHECK (naturalness_score BETWEEN 0 AND 100),
    overall_score SMALLINT CHECK (overall_score BETWEEN 0 AND 100),

    -- 상세 피드백
    summary JSONB DEFAULT '{}',                      -- {"strengths": [], "improvements": [], "tips": []}
    category_scores JSONB DEFAULT '{}',              -- {"anime": 85, "game": 70} 카테고리별 점수
    detailed_feedback TEXT,                          -- LLM 생성 상세 피드백

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- 6. STATISTICS TABLES (통계 테이블)
-- ============================================================================

-- 일일 학습 통계
CREATE TABLE daily_learning_stats (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
    stat_date DATE NOT NULL,
    sentences_learned INTEGER NOT NULL DEFAULT 0,
    quizzes_completed INTEGER NOT NULL DEFAULT 0,
    quizzes_correct INTEGER NOT NULL DEFAULT 0,
    chat_sessions_count INTEGER NOT NULL DEFAULT 0,
    total_study_minutes INTEGER NOT NULL DEFAULT 0,
    streak_days INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, stat_date)
);

-- ============================================================================
-- 7. INDEXES
-- ============================================================================

-- Master Tables
CREATE INDEX idx_levels_sort ON levels(sort_order) WHERE is_active = TRUE;
CREATE INDEX idx_categories_sort ON categories(sort_order) WHERE is_active = TRUE;
CREATE INDEX idx_tags_sort ON tags(sort_order) WHERE is_active = TRUE;

-- Users
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_status ON users(status) WHERE status = 'active';
CREATE INDEX idx_user_interests_category ON user_interests(category_id);
CREATE INDEX idx_user_onboarding_level ON user_onboarding(level_id);

-- Sentences
CREATE INDEX idx_sentences_level ON sentences(level_id) WHERE is_active = TRUE;
CREATE INDEX idx_sentences_category ON sentences(category_id) WHERE is_active = TRUE;
CREATE INDEX idx_sentences_level_category ON sentences(level_id, category_id) WHERE is_active = TRUE;
CREATE INDEX idx_sentences_difficulty ON sentences(difficulty_score) WHERE is_active = TRUE;
CREATE INDEX idx_sentence_tags_tag ON sentence_tags(tag_id);
CREATE INDEX idx_sentence_quizzes_sentence ON sentence_quizzes(sentence_id);
CREATE INDEX idx_sentence_quizzes_type ON sentence_quizzes(quiz_type_id);

-- Daily Sets
CREATE INDEX idx_daily_sets_user_date ON daily_sentence_sets(user_id, assigned_date DESC);
CREATE INDEX idx_daily_sets_date ON daily_sentence_sets(assigned_date);
CREATE INDEX idx_daily_set_items_sentence ON daily_sentence_set_items(sentence_id);

-- Learning Progress
CREATE INDEX idx_learning_progress_user ON learning_progress(user_id);
CREATE INDEX idx_learning_progress_sentence ON learning_progress(sentence_id);
CREATE INDEX idx_learning_progress_user_completed ON learning_progress(user_id, is_completed);
CREATE INDEX idx_learning_progress_daily_set ON learning_progress(daily_set_id);

-- Quiz Attempts
CREATE INDEX idx_quiz_attempts_user ON quiz_attempts(user_id);
CREATE INDEX idx_quiz_attempts_quiz ON quiz_attempts(quiz_id);
CREATE INDEX idx_quiz_attempts_user_date ON quiz_attempts(user_id, attempted_at DESC);

-- Chat
CREATE INDEX idx_chat_sessions_user ON chat_sessions(user_id);
CREATE INDEX idx_chat_sessions_user_date ON chat_sessions(user_id, started_at DESC);
CREATE INDEX idx_chat_sessions_status ON chat_sessions(status) WHERE status = 'active';
CREATE INDEX idx_chat_messages_session ON chat_messages(session_id);
CREATE INDEX idx_chat_messages_sentence ON chat_messages(sentence_id) WHERE sentence_id IS NOT NULL;

-- Statistics
CREATE INDEX idx_daily_stats_user_date ON daily_learning_stats(user_id, stat_date DESC);

-- ============================================================================
-- 8. TRIGGERS
-- ============================================================================

-- updated_at 자동 갱신 함수
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- updated_at 트리거 적용
CREATE TRIGGER update_levels_updated_at BEFORE UPDATE ON levels FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_categories_updated_at BEFORE UPDATE ON categories FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_tags_updated_at BEFORE UPDATE ON tags FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_quiz_types_updated_at BEFORE UPDATE ON quiz_types FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_user_settings_updated_at BEFORE UPDATE ON user_settings FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_user_onboarding_updated_at BEFORE UPDATE ON user_onboarding FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_sentences_updated_at BEFORE UPDATE ON sentences FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_sentence_details_updated_at BEFORE UPDATE ON sentence_details FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_sentence_quizzes_updated_at BEFORE UPDATE ON sentence_quizzes FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_learning_progress_updated_at BEFORE UPDATE ON learning_progress FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_chat_sessions_updated_at BEFORE UPDATE ON chat_sessions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_session_feedbacks_updated_at BEFORE UPDATE ON session_feedbacks FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_daily_learning_stats_updated_at BEFORE UPDATE ON daily_learning_stats FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- 9. INITIAL SEED DATA
-- ============================================================================

-- 레벨 초기 데이터
INSERT INTO levels (code, name_jp, name_kr, sort_order) VALUES
('N5', '初級1', '초급1 (N5)', 1),
('N4', '初級2', '초급2 (N4)', 2),
('N3', '中級1', '중급1 (N3)', 3),
('N2', '中級2', '중급2 (N2)', 4),
('N1', '上級', '상급 (N1)', 5);

-- 카테고리 초기 데이터 (오타쿠 카테고리)
INSERT INTO categories (code, name_jp, name_kr, icon_url, color_hex, sort_order) VALUES
('anime', 'アニメ', '애니메이션', '/icons/anime.svg', '#FF6B6B', 1),
('manga', 'マンガ', '만화', '/icons/manga.svg', '#4ECDC4', 2),
('game', 'ゲーム', '게임', '/icons/game.svg', '#45B7D1', 3),
('jpop', 'J-POP', 'J-POP/음악', '/icons/jpop.svg', '#96CEB4', 4),
('gacha', 'ガチャ', '가챠/굿즈', '/icons/gacha.svg', '#FFEAA7', 5),
('seichi', '聖地巡礼', '성지순례', '/icons/seichi.svg', '#DDA0DD', 6),
('idol', 'アイドル', '아이돌', '/icons/idol.svg', '#FFB6C1', 7),
('vtuber', 'VTuber', 'VTuber', '/icons/vtuber.svg', '#87CEEB', 8),
('cosplay', 'コスプレ', '코스프레', '/icons/cosplay.svg', '#F0E68C', 9),
('daily', '日常', '일상회화', '/icons/daily.svg', '#B8B8B8', 10);

-- 태그 초기 데이터
INSERT INTO tags (code, name_jp, name_kr, sort_order) VALUES
('greeting', '挨拶', '인사', 1),
('shopping', '買い物', '쇼핑', 2),
('restaurant', 'レストラン', '식당', 3),
('travel', '旅行', '여행', 4),
('emotion', '感情', '감정표현', 5),
('request', 'お願い', '부탁/요청', 6),
('question', '質問', '질문', 7),
('direction', '道案内', '길찾기', 8),
('hobby', '趣味', '취미', 9),
('slang', 'スラング', '슬랭/유행어', 10);

-- 퀴즈 유형 초기 데이터
INSERT INTO quiz_types (code, name_jp, name_kr, description) VALUES
('meaning', '意味', '의미 선택', '문장의 올바른 의미를 선택하세요'),
('fill_blank', '穴埋め', '빈칸 채우기', '빈칸에 들어갈 알맞은 단어를 선택하세요'),
('listening', 'リスニング', '듣기', '음성을 듣고 올바른 문장을 선택하세요'),
('ordering', '並び替え', '어순 배열', '단어를 올바른 순서로 배열하세요'),
('translation', '翻訳', '번역', '올바른 번역을 선택하세요');
