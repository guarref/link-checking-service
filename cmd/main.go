package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/guarref/link-checking-service/internal/links"
	"github.com/guarref/link-checking-service/internal/web"
)

const (
	storageFile = "storage.json"
	ttl         = 60 * time.Second
)

func main() {
	// 1. Инициализируем хранилище
	repo := links.NewStorage(ttl)

	// 2. Загружаем данные из файла (если есть)
	if err := repo.ReadFromJSONFile(storageFile); err != nil {
		log.Printf("warning: failed to load storage: %v", err)
	} else {
		log.Println("storage loaded successfully")
	}

	// 3. Инициализируем сервис
	service := links.NewService(repo)

	// 4. Инициализируем HTTP-хендлеры
	handler := web.NewHandler(service)

	// 5. Роутер
	mux := http.NewServeMux()
	mux.HandleFunc("/getjson", handler.GetStatusToJSON)
	mux.HandleFunc("/getpdf", handler.GetStatusToPDF)

	// 6. HTTP-сервер
	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// 7. Запускаем сервер в отдельной горутине
	go func() {
		log.Println("server starting on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %v", err)
		}
	}()

	// 8. Ловим сигналы и делаем graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("server forced to shutdown: %v", err)
	}

	// 9. После завершения всех запросов сохраняем storage
	if err := repo.SaveToJSONFile(storageFile); err != nil {
		log.Printf("error saving storage: %v", err)
	} else {
		log.Println("storage saved successfully")
	}

	log.Println("server exiting")
}
