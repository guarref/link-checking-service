package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/guarref/link-checking-service/internal/links"
	"github.com/guarref/link-checking-service/internal/web"
)

const (
	storageFile = "storage.json"
	ttl         = 60 * time.Second
)

type App struct {
	server *http.Server
	repo   *links.Storage
}

func New(ctx context.Context) (*App, error) {

	repo := links.NewStorage(ttl)

	err := repo.ReadFromJSONFile(storageFile)
	if err != nil {
		log.Printf("error loading storage: %v", err)
	}

	service := links.NewService(repo)
	handler := web.NewHandler(service)

	mux := http.NewServeMux()
	mux.HandleFunc("/getjson", handler.GetStatusToJSON)
	mux.HandleFunc("/getpdf", handler.GetStatusToPDF)

	server := &http.Server{Addr: ":8080", Handler: mux}

	return &App{server: server, repo: repo}, nil
}

func (a *App) Run(ctx context.Context) error {

	serverErr := make(chan error, 1)

	go func() {
		log.Println("server is starting on port :8080")
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Println("\nshutdown was caused")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err := a.server.Shutdown(shutdownCtx)
		if err != nil {
			return fmt.Errorf("\nshutdown(server): %w", err)
		}

		errr := a.repo.SaveToJSONFile(storageFile)
		if errr != nil {
			log.Printf("error saving storage: %v", errr)
		} else {
			log.Println("storage saved successfully")
		}

		return nil

	case err := <-serverErr:
		return fmt.Errorf("server error: %w", err)
	}
}
