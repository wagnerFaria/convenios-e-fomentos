package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wagner/convenios-e-fomentos/backend/internal/config"
	"github.com/wagner/convenios-e-fomentos/backend/internal/httpserver"
)

func main() {
	cfg := config.Load()

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           httpserver.NewRouter(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("servidor ouvindo em %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("erro no servidor: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("erro ao encerrar: %v", err)
	}
	log.Println("servidor encerrado")
}
