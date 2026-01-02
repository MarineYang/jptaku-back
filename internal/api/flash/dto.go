package flash

// UpdateFlashProgressRequest Flash 진행 상황 업데이트 요청
type UpdateFlashProgressRequest struct {
	SentenceID uint   `json:"sentence_id" binding:"required"`
	Grade      string `json:"grade" binding:"required,oneof=bad mid good"`
}
