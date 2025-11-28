-- ============================================================================
-- 오늘의 5문장 추천 쿼리
-- ============================================================================

-- 1. 기존 일일 세트 확인 (오늘 이미 생성되었는지)
SELECT id, assigned_date, is_completed
FROM daily_sentence_sets
WHERE user_id = :user_id
  AND assigned_date = CURRENT_DATE;

-- 2. 과거에 학습한 문장 ID 목록 (제외 대상)
WITH learned_sentences AS (
    SELECT DISTINCT sentence_id
    FROM learning_progress
    WHERE user_id = :user_id
      AND is_completed = TRUE
)
SELECT sentence_id FROM learned_sentences;

-- 3. 오늘의 5문장 추천 (핵심 쿼리)
-- 조건: 사용자 레벨 + 관심 카테고리 + 과거 학습 제외 + 난이도 순
WITH user_profile AS (
    -- 사용자 레벨과 관심 카테고리 조회
    SELECT
        uo.level_id,
        ARRAY_AGG(ui.category_id ORDER BY ui.priority) AS interest_categories
    FROM user_onboarding uo
    LEFT JOIN user_interests ui ON ui.user_id = uo.user_id
    WHERE uo.user_id = :user_id
    GROUP BY uo.level_id
),
learned_sentences AS (
    -- 이미 학습 완료한 문장 제외
    SELECT DISTINCT sentence_id
    FROM learning_progress
    WHERE user_id = :user_id
      AND is_completed = TRUE
),
recently_assigned AS (
    -- 최근 7일 내 할당된 문장도 제외 (다양성 확보)
    SELECT DISTINCT dsi.sentence_id
    FROM daily_sentence_sets dss
    JOIN daily_sentence_set_items dsi ON dsi.set_id = dss.id
    WHERE dss.user_id = :user_id
      AND dss.assigned_date > CURRENT_DATE - INTERVAL '7 days'
)
SELECT
    s.id,
    s.jp_text,
    s.kr_text,
    s.romaji,
    s.level_id,
    s.category_id,
    s.difficulty_score,
    c.name_kr AS category_name,
    l.name_kr AS level_name,
    -- 관심 카테고리 우선순위 점수 (낮을수록 우선)
    CASE
        WHEN s.category_id = ANY((SELECT interest_categories FROM user_profile))
        THEN array_position((SELECT interest_categories FROM user_profile), s.category_id)
        ELSE 999
    END AS interest_priority
FROM sentences s
JOIN categories c ON c.id = s.category_id
JOIN levels l ON l.id = s.level_id
CROSS JOIN user_profile up
WHERE s.is_active = TRUE
  AND s.level_id = up.level_id
  AND s.id NOT IN (SELECT sentence_id FROM learned_sentences)
  AND s.id NOT IN (SELECT sentence_id FROM recently_assigned)
ORDER BY
    interest_priority ASC,           -- 관심 카테고리 우선
    s.difficulty_score ASC,          -- 쉬운 것부터
    s.usage_count ASC,               -- 덜 추천된 것 우선
    RANDOM()                          -- 랜덤 요소
LIMIT 5;

-- 4. 일일 세트 생성 (트랜잭션)
BEGIN;

-- 4-1. 세트 생성
INSERT INTO daily_sentence_sets (user_id, assigned_date)
VALUES (:user_id, CURRENT_DATE)
RETURNING id AS set_id;

-- 4-2. 문장 5개 할당 (위 쿼리 결과 사용)
INSERT INTO daily_sentence_set_items (set_id, sentence_id, display_order)
VALUES
    (:set_id, :sentence_id_1, 1),
    (:set_id, :sentence_id_2, 2),
    (:set_id, :sentence_id_3, 3),
    (:set_id, :sentence_id_4, 4),
    (:set_id, :sentence_id_5, 5);

-- 4-3. 문장 사용 카운트 증가
UPDATE sentences
SET usage_count = usage_count + 1
WHERE id IN (:sentence_id_1, :sentence_id_2, :sentence_id_3, :sentence_id_4, :sentence_id_5);

COMMIT;

-- 5. 오늘의 문장 세트 조회 (학습 화면용)
SELECT
    dss.id AS set_id,
    dss.assigned_date,
    dss.is_completed AS set_completed,
    dsi.display_order,
    s.id AS sentence_id,
    s.jp_text,
    s.kr_text,
    s.romaji,
    c.name_kr AS category_name,
    c.icon_url AS category_icon,
    l.name_kr AS level_name,
    -- 학습 진행 상태
    COALESCE(lp.step_understand, FALSE) AS step_understand,
    COALESCE(lp.step_speak, FALSE) AS step_speak,
    COALESCE(lp.step_confirm, FALSE) AS step_confirm,
    COALESCE(lp.step_memorize, FALSE) AS step_memorize,
    COALESCE(lp.is_completed, FALSE) AS is_completed
FROM daily_sentence_sets dss
JOIN daily_sentence_set_items dsi ON dsi.set_id = dss.id
JOIN sentences s ON s.id = dsi.sentence_id
JOIN categories c ON c.id = s.category_id
JOIN levels l ON l.id = s.level_id
LEFT JOIN learning_progress lp ON lp.user_id = dss.user_id
    AND lp.sentence_id = s.id
    AND lp.daily_set_id = dss.id
WHERE dss.user_id = :user_id
  AND dss.assigned_date = CURRENT_DATE
ORDER BY dsi.display_order;

-- 6. 학습 진행도 업데이트
UPDATE learning_progress
SET
    step_understand = :step_understand,
    step_speak = :step_speak,
    step_confirm = :step_confirm,
    step_memorize = :step_memorize,
    is_completed = (:step_understand AND :step_speak AND :step_confirm AND :step_memorize),
    completed_at = CASE
        WHEN :step_understand AND :step_speak AND :step_confirm AND :step_memorize
        THEN NOW()
        ELSE NULL
    END,
    time_spent_seconds = time_spent_seconds + :additional_seconds
WHERE user_id = :user_id
  AND sentence_id = :sentence_id
  AND daily_set_id = :daily_set_id;

-- 7. 학습 진행도 없으면 생성 (UPSERT)
INSERT INTO learning_progress (user_id, sentence_id, daily_set_id, step_understand)
VALUES (:user_id, :sentence_id, :daily_set_id, TRUE)
ON CONFLICT (user_id, sentence_id, daily_set_id)
DO UPDATE SET
    step_understand = TRUE,
    updated_at = NOW();
