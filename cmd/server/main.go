package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/employee-management-rest-api/internal/employee"
	"github.com/employee-management-rest-api/internal/middleware"
)

func main() {
	addr := ":8080"
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	}

	mux := http.NewServeMux()
	employee.NewHandler(employee.NewService(employee.NewStore())).Register(mux)

	handler := middleware.Chain(mux,
		middleware.RequestID,
		middleware.Logger,
		middleware.Recover,
	)

	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("employee API listening on %s", addr)
	log.Fatal(srv.ListenAndServe())
}
