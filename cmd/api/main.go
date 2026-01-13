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
	if err := usersqlite.Migrate(db); err != nil {
		log.Fatal("[MAIN] migrate users:", err)
	}

	//jwt токен
	cfg := config.Load()
	jwtManager := auth.NewJWTManager(
		cfg.JWTSecret,
		cfg.JWTTTL,
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
