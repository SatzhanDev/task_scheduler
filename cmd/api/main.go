package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"
	"task_scheduler/internal/httpserver"
	"task_scheduler/internal/task"
	"time"

	_ "modernc.org/sqlite"

	tasksqlite "task_scheduler/internal/task/sqlite"
)

func main() {
	addr := ":8080"
	// --- SQLite init ---
	_ = os.MkdirAll("data", 0o755)

	db, err := sql.Open("sqlite", "data/tasks.db")
	if err != nil {
		log.Fatal("[MAIN] open db:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("[MAIN] ping db:", err)
	}

	if err := tasksqlite.Migrate(db); err != nil {
		log.Fatal("[MAIN] migrate db:", err)
	}
	// --- /SQLite init ---

	svc := task.NewService()
	srv := httpserver.New(addr, svc)

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
