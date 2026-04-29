package handler

import (
	"net/http"

	"mota/internal/model"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type AdminHandler struct{ db *gorm.DB }

func NewAdminHandler(db *gorm.DB) *AdminHandler { return &AdminHandler{db} }

// @Summary      List reports
// @Description  Returns all reports containing the post's content
// @Tags         admin
// @Produce      json
// @Success      200  {array}  object
// @Router       /admin/reports [get]
func (h *AdminHandler) ListReports(w http.ResponseWriter, r *http.Request) {
	type ReportWhitPost struct {
		model.PostReport
		Content string `json:"content"`
	}

	var reports []ReportWhitPost
	h.db.Table("post_reports").
		Select("post_reports.*, posts.content").
		Joins("JOIN posts ON posts.id = post_reports.post_id").
		Order("post_reports.created_at desc").
		Limit(100).
		Scan(&reports)

	writeJSON(w, 200, reports)
}

// @Summary      Delete post
// @Description  Soft delete a post
// @Tags         admin
// @Produce      json
// @Param        id  path  string  true  "Post ID"
// @Success      200  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /admin/posts/{id} [delete]
func (h *AdminHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	result := h.db.Delete(&model.Post{}, "id = ?", id)

	if result.RowsAffected == 0 {
		writeJSON(w, 404, map[string]string{"error": "post not found"})
		return
	}

	writeJSON(w, 200, map[string]string{"message": "post deleted!"})
}
