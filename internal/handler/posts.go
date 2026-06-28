package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"notes-api/internal/cache"
	"notes-api/internal/model"
	"notes-api/internal/moderation"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type PostHandler struct{ db *gorm.DB }

func NewPostHandler(db *gorm.DB) *PostHandler { return &PostHandler{db} }

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Comtent-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// @Summary      List posts
// @Description  Returns paginated, approved posts.
// @Tags         posts
// @Produce      json
// @Param        page  query  int  false  "Page"
// @Success      200  {array}   model.Post
// @Router       /posts [get]
func (h *PostHandler) List(w http.ResponseWriter, r *http.Request) {
	page := 1

	if p := r.URL.Query().Get("page"); p != "" {
		json.Unmarshal([]byte(p), &page)
	}
	if page < 1 {
		page = 1
	}

	cacheKey := fmt.Sprintf("posts:list:%d", page)

	if data, err := cache.Get(cacheKey); err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
		return
	}

	var posts []model.Post
	h.db.Where("status = ?", "approved").
		Order("created_at desc").
		Limit(20).
		Offset((page - 1) * 20).
		Find(&posts)

	data, _ := json.Marshal(posts)
	cache.Set(cacheKey, data, 60*time.Second)

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

// @Summary      Get post
// @Description  Returns post by id
// @Tags         posts
// @Produce      json
// @Param        id  path  string  true  "Post ID"
// @Success      200  {object}  model.Post
// @Failure      404  {object}  map[string]string
// @Router       /posts/{id} [get]
func (h *PostHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var post model.Post

	result := h.db.Where("id = ? AND status = ?", id, "approved").First(&post)
	if result.Error != nil {
		writeJSON(w, 404, map[string]string{"error": "post not found"})
		return
	}

	writeJSON(w, 200, post)
}

// @Summary      Post random
// @Description  Returns a random post
// @Tags         posts
// @Produce      json
// @Success      200  {object}  model.Post
// @Failure      404  {object}  map[string]string
// @Router       /posts/random [get]
func (h *PostHandler) Random(w http.ResponseWriter, r *http.Request) {
	cacheKey := "posts:random"

	if data, err := cache.Get(cacheKey); err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
		return
	}

	var post model.Post

	result := h.db.Raw(`
		SELECT * FROM posts TABLESAMPLE BERNOULLI(10)
		WHERE status = 'approved' AND deleted_at IS NULL
		LIMIT 1
	`).Scan(&post)

	if result.RowsAffected == 0 {
		h.db.Where("status = ? AND deleted_at IS NULL", "approved").
			Order("RANDOM()").
			First(&post)
	}

	if post.ID == "" {
		writeJSON(w, 404, map[string]string{"error": "posts not found"})
		return
	}

	data, _ := json.Marshal(post)
	cache.Set(cacheKey, data, 60*time.Second)

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

// @Summary      Create post
// @Description  Creates a post
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        body  body  object{content=string,drawing=string,bg=string,fg=string}  true  "Content"
// @Success      201  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Router       /posts [post]
func (h *PostHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Content string  `json:"content"`
		Drawing *string `json:"drawing"`
		BgColor string  `json:"bg"`
		FgColor string  `json:"fg"`
	}

	hexColorRegex := regexp.MustCompile(`^#([A-Fa-f0-9]{6})$`)

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid JSON"})
		return
	}
	if body.Content == "" {
		writeJSON(w, 400, map[string]string{"error": "content is required"})
		return
	}
	if len(body.Content) > 250 {
		writeJSON(w, 400, map[string]string{"error": "content too long"})
		return
	}
	if !hexColorRegex.MatchString(body.BgColor) {
		body.BgColor = "#f7f19e"
	}
	if !hexColorRegex.MatchString(body.FgColor) {
		body.FgColor = "#1e1f24"
	}

	result := moderation.Moderate(body.Content)

	status := "pending"
	if result.Approved {
		status = "approved"
	} else if !result.Approved && result.Score < 0.3 {
		status = "rejected"
	}

	post := model.Post{
		Content: body.Content,
		Drawing: body.Drawing,
		BgColor: body.BgColor,
		FgColor: body.FgColor,
		Status:  status,
	}

	if err := h.db.Create(&post).Error; err != nil {
		writeJSON(w, 500, map[string]string{"error": "error while saving"})
		return
	}
	cache.Del("posts:random")
	cache.Del("posts:list:1")

	msg := "your note is pending review"
	if status == "approved" {
		msg = "your note has been published!"
	} else if status == "rejected" {
		msg = "your note could not be published: " + result.Reason
	}

	writeJSON(w, 200, map[string]any{
		"id":      post.ID,
		"status":  status,
		"message": msg,
	})

}

// @Summary      Like
// @Description  Registers like by IP
// @Tags         posts
// @Produce      json
// @Param        id  path  string  true  "Post ID"
// @Success      200  {object}  map[string]string
// @Router       /posts/{id}/like [post]
func (h *PostHandler) Like(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.RemoteAddr
	}

	like := model.PostLike{PostID: id, IPAddress: ip}
	result := h.db.Where(model.PostLike{PostID: id, IPAddress: ip}).FirstOrCreate(&like)

	if result.RowsAffected == 0 {
		writeJSON(w, 200, map[string]string{"message": "you already liked this note"})
		return
	}

	h.db.Model(&model.Post{}).Where("id = ?", id).UpdateColumn("likes", gorm.Expr("likes + 1"))
	cache.Del("posts:list:1")

	writeJSON(w, 200, map[string]string{"message": "like registered!"})
}

// @Summary      Report post
// @Description  Sends a report about a post
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        id    path  string               true  "Post ID"
// @Param        body  body  object{reason=string}  false  "Reason"
// @Success      200  {object}  map[string]string
// @Router       /posts/{id}/report [post]
func (h *PostHandler) Report(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var body struct {
		Reason *string `json:"reason"`
	}
	json.NewDecoder(r.Body).Decode(&body)

	report := model.PostReport{
		PostID: id,
		Reason: body.Reason,
	}

	if err := h.db.Create(&report).Error; err != nil {
		writeJSON(w, 500, map[string]string{"error": "error while reporting"})
		return
	}

	writeJSON(w, 200, map[string]string{"message": "report send, thank you"})
}
