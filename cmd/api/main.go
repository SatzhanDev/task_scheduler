package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"task_scheduler/internal/httpserver"
	"time"
)

func main() {
	addr := ":8080"
	srv := httpserver.New(addr)

	go func() {
		log.Println("[MAIN] starting server on", addr)

		if err := srv.Start(); err != nil {
			log.Println("[MAIN] server stopped:", err)

		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("[MAIN] shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Println("[MAIN] shutdown error:", err)
	}
	log.Println("[MAIN] exited")
}
