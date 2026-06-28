package main

import (
	"log"
	"net/http"
	"os"

	"notes-api/internal/cache"
	"notes-api/internal/db"
	"notes-api/internal/handler"
	mw "notes-api/internal/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "notes-api/docs"
)

// @title           notes API
// @version         1.0
// @description     positive notes portal
// @host            localhost:8080
// @BasePath        /
// @securityDefinitions.apikey AdminKey
// @in header
// @name X-Admin-Key
func main() {
	godotenv.Load()

	database, err := db.Connect()
	if err != nil {
		log.Fatal("error connecting to db: ", err)
	}

	if err := cache.Connect(); err != nil {
		log.Fatal("error conecting to Redis:", err)
	}
	log.Println("Redis conected!")

	ph := handler.NewPostHandler(database)
	ah := handler.NewAdminHandler(database)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(mw.RateLimit)

	// swaggerUI
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	// public
	r.Get("/posts", ph.List)
	r.Get("/posts/{id}", ph.Get)
	r.Get("/posts/random", ph.Random)
	r.Post("/posts", ph.Create)
	r.Post("/posts/{id}/like", ph.Like)
	r.Post("/posts/{id}/report", ph.Report)

	// admin
	r.Group(func(r chi.Router) {
		r.Use(mw.AdminAuth)
		r.Get("/admin/reports", ah.ListReports)
		r.Delete("/admin/posts/{id}", ah.Delete)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("mota running at :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
