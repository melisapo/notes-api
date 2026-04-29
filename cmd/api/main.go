package main

import (
	"log"
	"net/http"
	"os"

	"mota/internal/db"
	"mota/internal/handler"
	mw "mota/internal/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "mota/docs"
)

// @title           mota API
// @version         1.0
// @description     Send anonymous notes for everyone to read.
// @host            localhost:8080
// @BasePath        /
func main() {
	godotenv.Load()

	database, err := db.Connect()
	if err != nil {
		log.Fatal("error connecting to db: ", err)
	}

	ph := handler.NewPostHandler(database)
	ah := handler.NewAdminHandler(database)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)

	// swaggerUI
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	// public
	r.Get("/posts", ph.List)
	r.Get("/posts/{id}", ph.Get)
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
