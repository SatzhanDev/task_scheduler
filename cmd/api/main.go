package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"
	"task_scheduler/internal/auth"
	"task_scheduler/internal/config"
	"task_scheduler/internal/httpserver"
	"task_scheduler/internal/task"
	"task_scheduler/internal/user"
	"time"

	_ "modernc.org/sqlite"

	tasksqlite "task_scheduler/internal/task/sqlite"
	usersqlite "task_scheduler/internal/user/sqlite"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("[MAIN] load config:", err)
	}
	addr := cfg.HTTP.Addr
	// --- SQLite init ---
	_ = os.MkdirAll("data", 0o755)

	db, err := sql.Open("sqlite", cfg.DB.Path)
	if err != nil {
		log.Fatal("[MAIN] open db:", err)
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		log.Fatal("[MAIN] ping db:", err)
	}

	if err := tasksqlite.Migrate(db); err != nil {
		_ = db.Close()
		log.Fatal("[MAIN] migrate db:", err)
	}
	if err := usersqlite.Migrate(db); err != nil {
		_ = db.Close()
		log.Fatal("[MAIN] migrate users:", err)
	}

	//jwt токен
	jwtManager := auth.NewJWTManager(
		cfg.JWT.Secret,
		cfg.JWT.TTL,
	)
	//repos
	taskRepo := tasksqlite.New(db)
	userRepo := usersqlite.New(db)

	//services
	taskSvc := task.NewService(taskRepo)
	userSvc := user.NewService(userRepo)

	//servers
	srv := httpserver.New(addr, taskSvc, userSvc, jwtManager)

	go func() {
		log.Println("[MAIN] starting server on", addr)

		if err := srv.Start(); err != nil {
			log.Println("[MAIN] server stopped:", err)

		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	log.Println("[MAIN] shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Println("[MAIN] shutdown error:", err)
	}
	if err := db.Close(); err != nil {
		log.Println("[MAIN] db close error:", err)
	}
	log.Println("[MAIN] exited")
}
