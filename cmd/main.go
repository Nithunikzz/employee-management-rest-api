package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"github.com/employee-management-rest-api/handlers"
	"github.com/employee-management-rest-api/middleware"
	"github.com/employee-management-rest-api/repository"
	"github.com/employee-management-rest-api/services"
)

func main() {
	addr := ":8080"
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	}

	binding.EnableDecoderDisallowUnknownFields = true

	repo := repository.NewInMemoryEmployeeRepository()
	svc := services.NewEmployeeService(repo)

	router := gin.New()
	router.Use(middleware.RequestID(), middleware.Logger(), middleware.Recovery())
	handlers.NewEmployeeHandler(svc).RegisterRoutes(router)

	srv := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("employee API listening on %s", addr)
	log.Fatal(srv.ListenAndServe())
}
