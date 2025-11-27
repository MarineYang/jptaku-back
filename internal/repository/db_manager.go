package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// DBManager는 데이터베이스 세션 및 트랜잭션을 관리합니다.
type DBManager struct {
	db *gorm.DB
}

// NewDBManager는 새 DBManager 인스턴스를 생성합니다.
func NewDBManager(db *gorm.DB) *DBManager {
	return &DBManager{db: db}
}

// DB는 기본 GORM DB 인스턴스를 반환합니다.
func (m *DBManager) DB() *gorm.DB {
	return m.db
}

// WithContext는 context가 적용된 DB 세션을 반환합니다.
func (m *DBManager) WithContext(ctx context.Context) *gorm.DB {
	return m.db.WithContext(ctx)
}

// WithTimeout은 timeout이 적용된 context로 DB 세션을 반환합니다.
func (m *DBManager) WithTimeout(timeout time.Duration) (*gorm.DB, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	return m.db.WithContext(ctx), cancel
}

// Transaction은 트랜잭션 내에서 함수를 실행합니다.
// 에러가 발생하면 자동으로 롤백됩니다.
func (m *DBManager) Transaction(fn func(tx *gorm.DB) error) error {
	return m.db.Transaction(fn)
}

// TransactionWithContext는 context가 적용된 트랜잭션을 실행합니다.
func (m *DBManager) TransactionWithContext(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return m.db.WithContext(ctx).Transaction(fn)
}

// ========================================
// 쿼리 결과 래퍼
// ========================================

// QueryResult는 단일 쿼리 결과를 래핑합니다.
type QueryResult[T any] struct {
	Data  T
	Error error
	Found bool
}

// QueryListResult는 리스트 쿼리 결과를 래핑합니다.
type QueryListResult[T any] struct {
	Data  []T
	Error error
	Total int64
}

// PagedResult는 페이지네이션된 결과를 래핑합니다.
type PagedResult[T any] struct {
	Data       []T   `json:"data"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	TotalPages int   `json:"total_pages"`
}

// ========================================
// 쿼리 헬퍼 함수
// ========================================

// FindOne은 단일 레코드를 조회합니다.
func FindOne[T any](db *gorm.DB, conditions ...interface{}) QueryResult[T] {
	var result T
	var query *gorm.DB

	if len(conditions) > 0 {
		query = db.Where(conditions[0], conditions[1:]...).First(&result)
	} else {
		query = db.First(&result)
	}

	if query.Error != nil {
		if errors.Is(query.Error, gorm.ErrRecordNotFound) {
			return QueryResult[T]{Found: false}
		}
		return QueryResult[T]{Error: query.Error}
	}

	return QueryResult[T]{Data: result, Found: true}
}

// FindByID는 ID로 레코드를 조회합니다.
func FindByID[T any](db *gorm.DB, id uint) QueryResult[T] {
	var result T
	if err := db.First(&result, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return QueryResult[T]{Found: false}
		}
		return QueryResult[T]{Error: err}
	}
	return QueryResult[T]{Data: result, Found: true}
}

// FindAll은 조건에 맞는 모든 레코드를 조회합니다.
func FindAll[T any](db *gorm.DB, conditions ...interface{}) QueryListResult[T] {
	var results []T
	var query *gorm.DB

	if len(conditions) > 0 {
		query = db.Where(conditions[0], conditions[1:]...)
	} else {
		query = db
	}

	if err := query.Find(&results).Error; err != nil {
		return QueryListResult[T]{Error: err}
	}

	return QueryListResult[T]{Data: results, Total: int64(len(results))}
}

// FindPaged는 페이지네이션된 결과를 반환합니다.
func FindPaged[T any](db *gorm.DB, page, perPage int, conditions ...interface{}) PagedResult[T] {
	var results []T
	var total int64
	var query *gorm.DB

	// 기본값 설정
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	offset := (page - 1) * perPage

	// 조건 적용
	if len(conditions) > 0 {
		query = db.Model(new(T)).Where(conditions[0], conditions[1:]...)
	} else {
		query = db.Model(new(T))
	}

	// 총 개수 조회
	query.Count(&total)

	// 페이지네이션 적용하여 데이터 조회
	if len(conditions) > 0 {
		db.Where(conditions[0], conditions[1:]...).Offset(offset).Limit(perPage).Find(&results)
	} else {
		db.Offset(offset).Limit(perPage).Find(&results)
	}

	// 총 페이지 수 계산
	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	return PagedResult[T]{
		Data:       results,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}
}

// ========================================
// Create/Update/Delete 헬퍼
// ========================================

// Create는 새 레코드를 생성합니다.
func Create[T any](db *gorm.DB, model *T) error {
	return db.Create(model).Error
}

// Update는 레코드를 업데이트합니다.
func Update[T any](db *gorm.DB, model *T) error {
	return db.Save(model).Error
}

// UpdateFields는 특정 필드만 업데이트합니다.
func UpdateFields[T any](db *gorm.DB, model *T, fields map[string]interface{}) error {
	return db.Model(model).Updates(fields).Error
}

// Delete는 레코드를 삭제합니다 (소프트 삭제 지원).
func Delete[T any](db *gorm.DB, id uint) error {
	return db.Delete(new(T), id).Error
}

// HardDelete는 레코드를 완전히 삭제합니다.
func HardDelete[T any](db *gorm.DB, id uint) error {
	return db.Unscoped().Delete(new(T), id).Error
}

// ========================================
// 존재 확인 헬퍼
// ========================================

// Exists는 조건에 맞는 레코드가 존재하는지 확인합니다.
func Exists[T any](db *gorm.DB, conditions ...interface{}) (bool, error) {
	var count int64
	var query *gorm.DB

	if len(conditions) > 0 {
		query = db.Model(new(T)).Where(conditions[0], conditions[1:]...)
	} else {
		query = db.Model(new(T))
	}

	if err := query.Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

// ExistsByID는 해당 ID의 레코드가 존재하는지 확인합니다.
func ExistsByID[T any](db *gorm.DB, id uint) (bool, error) {
	var count int64
	if err := db.Model(new(T)).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// ========================================
// 에러 헬퍼
// ========================================

// IsNotFound는 에러가 레코드 없음 에러인지 확인합니다.
func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

// IsDuplicateError는 중복 키 에러인지 확인합니다 (PostgreSQL).
func IsDuplicateError(err error) bool {
	if err == nil {
		return false
	}
	// PostgreSQL 중복 키 에러 코드: 23505
	return errors.Is(err, gorm.ErrDuplicatedKey) ||
		containsString(err.Error(), "23505") ||
		containsString(err.Error(), "duplicate key")
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ========================================
// 쿼리 빌더 헬퍼
// ========================================

// QueryBuilder는 체이닝 방식의 쿼리 빌더입니다.
type QueryBuilder[T any] struct {
	db *gorm.DB
}

// NewQueryBuilder는 새 QueryBuilder를 생성합니다.
func NewQueryBuilder[T any](db *gorm.DB) *QueryBuilder[T] {
	return &QueryBuilder[T]{db: db.Model(new(T))}
}

// Where는 조건을 추가합니다.
func (q *QueryBuilder[T]) Where(query interface{}, args ...interface{}) *QueryBuilder[T] {
	q.db = q.db.Where(query, args...)
	return q
}

// OrderBy는 정렬 조건을 추가합니다.
func (q *QueryBuilder[T]) OrderBy(order string) *QueryBuilder[T] {
	q.db = q.db.Order(order)
	return q
}

// Preload는 관계를 미리 로드합니다.
func (q *QueryBuilder[T]) Preload(relation string, args ...interface{}) *QueryBuilder[T] {
	q.db = q.db.Preload(relation, args...)
	return q
}

// Limit는 결과 개수를 제한합니다.
func (q *QueryBuilder[T]) Limit(limit int) *QueryBuilder[T] {
	q.db = q.db.Limit(limit)
	return q
}

// Offset은 시작 위치를 설정합니다.
func (q *QueryBuilder[T]) Offset(offset int) *QueryBuilder[T] {
	q.db = q.db.Offset(offset)
	return q
}

// First는 첫 번째 결과를 반환합니다.
func (q *QueryBuilder[T]) First() QueryResult[T] {
	var result T
	if err := q.db.First(&result).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return QueryResult[T]{Found: false}
		}
		return QueryResult[T]{Error: err}
	}
	return QueryResult[T]{Data: result, Found: true}
}

// All은 모든 결과를 반환합니다.
func (q *QueryBuilder[T]) All() QueryListResult[T] {
	var results []T
	if err := q.db.Find(&results).Error; err != nil {
		return QueryListResult[T]{Error: err}
	}
	return QueryListResult[T]{Data: results, Total: int64(len(results))}
}

// Paged는 페이지네이션된 결과를 반환합니다.
func (q *QueryBuilder[T]) Paged(page, perPage int) PagedResult[T] {
	var results []T
	var total int64

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	// 총 개수 조회 (별도 쿼리로)
	countDB := q.db.Session(&gorm.Session{})
	countDB.Count(&total)

	// 데이터 조회
	offset := (page - 1) * perPage
	q.db.Offset(offset).Limit(perPage).Find(&results)

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	return PagedResult[T]{
		Data:       results,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}
}

// Count는 조건에 맞는 레코드 수를 반환합니다.
func (q *QueryBuilder[T]) Count() (int64, error) {
	var count int64
	err := q.db.Count(&count).Error
	return count, err
}

// ========================================
// 배치 작업 헬퍼
// ========================================

// BatchCreate는 여러 레코드를 배치로 생성합니다.
func BatchCreate[T any](db *gorm.DB, models []T, batchSize int) error {
	if batchSize <= 0 {
		batchSize = 100
	}
	return db.CreateInBatches(models, batchSize).Error
}

// BatchUpdate는 조건에 맞는 모든 레코드를 업데이트합니다.
func BatchUpdate[T any](db *gorm.DB, updates map[string]interface{}, conditions ...interface{}) (int64, error) {
	var query *gorm.DB
	if len(conditions) > 0 {
		query = db.Model(new(T)).Where(conditions[0], conditions[1:]...)
	} else {
		query = db.Model(new(T))
	}

	result := query.Updates(updates)
	return result.RowsAffected, result.Error
}

// ========================================
// 디버그 헬퍼
// ========================================

// Debug는 다음 쿼리의 SQL을 출력합니다.
func (m *DBManager) Debug() *gorm.DB {
	return m.db.Debug()
}

// Explain은 쿼리 실행 계획을 반환합니다.
func Explain[T any](db *gorm.DB, conditions ...interface{}) (string, error) {
	var result T
	var query *gorm.DB

	if len(conditions) > 0 {
		query = db.Model(&result).Where(conditions[0], conditions[1:]...)
	} else {
		query = db.Model(&result)
	}

	var explanation string
	row := query.Raw(fmt.Sprintf("EXPLAIN %s", query.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Find(&result)
	}))).Row()

	if err := row.Scan(&explanation); err != nil {
		return "", err
	}
	return explanation, nil
}
