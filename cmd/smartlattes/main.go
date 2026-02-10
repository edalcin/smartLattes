package main

import (
	"context"
	_ "embed"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/edalcin/smartlattes/internal/handler"
	"github.com/edalcin/smartlattes/internal/static"
	"github.com/edalcin/smartlattes/internal/store"
)

//go:embed resumoPrompt.md
var resumoPrompt string

//go:embed analisePrompt.md
var analisePrompt string

func main() {
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Fatal("MONGODB_URI é obrigatório")
	}

	dbName := os.Getenv("MONGODB_DATABASE")
	if dbName == "" {
		dbName = "smartLattes"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	maxUploadSize := int64(10485760)
	if v := os.Getenv("MAX_UPLOAD_SIZE"); v != "" {
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
			maxUploadSize = parsed
		}
	}

	var db *store.MongoDB
	db, err := store.Connect(mongoURI, dbName)
	if err != nil {
		if db != nil {
			log.Printf("AVISO: MongoDB não acessível no momento: %v", err)
		} else {
			log.Printf("AVISO: Falha ao conectar ao MongoDB: %v", err)
		}
	} else {
		log.Println("Conectado ao MongoDB com sucesso")
	}

	handler.InitStatic(static.Files)

	mux := http.NewServeMux()

	mux.HandleFunc("/", handler.PageHandler("index.html"))
	mux.HandleFunc("/upload", handler.PageHandler("upload.html"))
	mux.HandleFunc("/resumo", handler.PageHandler("resumo.html"))
	mux.HandleFunc("/visualizar-resumo", handler.PageHandler("visualizar-resumo.html"))
	mux.HandleFunc("/analise", handler.PageHandler("analise.html"))
	mux.HandleFunc("/visualizar-relacoes", handler.PageHandler("visualizar-relacoes.html"))
	mux.Handle("/static/", http.StripPrefix("/static/", handler.StaticHandler()))

	mux.Handle("/api/upload", &handler.UploadHandler{
		Store:         db,
		MaxUploadSize: maxUploadSize,
	})
	mux.Handle("/api/health", &handler.HealthHandler{
		Store: db,
	})

	summaryHandler := &handler.SummaryHandler{
		Store:  db,
		Prompt: resumoPrompt,
	}
	mux.Handle("/api/stats", &handler.StatsHandler{Store: db})
	mux.Handle("/api/search", &handler.SearchHandler{Store: db})
	mux.Handle("/api/models", &handler.ModelsHandler{})
	mux.Handle("/api/summary", summaryHandler)
	mux.Handle("/api/summary/save", summaryHandler)
	mux.Handle("/api/download/", &handler.DownloadHandler{Store: db})

	analysisHandler := &handler.AnalysisHandler{
		Store:  db,
		Prompt: analisePrompt,
	}
	mux.Handle("/api/analysis", analysisHandler)
	mux.Handle("/api/analysis/save", analysisHandler)
	mux.Handle("/api/analysis/download/", &handler.AnalysisDownloadHandler{Store: db})
	mux.Handle("/api/summary/view/", &handler.SummaryViewHandler{Store: db})
	mux.Handle("/api/analysis/view/", &handler.AnalysisViewHandler{Store: db})

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 150 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Servidor iniciado na porta %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Erro ao iniciar servidor: %v", err)
		}
	}()

	<-done
	log.Println("Encerrando servidor...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Erro ao encerrar servidor: %v", err)
	}

	if db != nil {
		if err := db.Disconnect(ctx); err != nil {
			log.Printf("Erro ao desconectar do MongoDB: %v", err)
		}
	}

	log.Println("Servidor encerrado")
}
