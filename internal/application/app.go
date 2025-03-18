package application

import (
	"context"
	"github.com/Arlandaren/pgxWrappy/pkg/postgres"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"service/internal/infrastructure/config"
	"service/internal/infrastructure/storage/minio"
	"service/internal/infrastructure/storage/redis"
	"sync"
	"syscall"
)

type App struct {
	Controller *Controller
	wg         *sync.WaitGroup
	cfg        *config.Config
}

func NewApp(db *postgres.Wrapper, rdb *redis.RDB, s3 *minio.Minio, r *gin.Engine, cfg *config.Config) *App {
	repo := NewRepository(db, rdb, s3)
	svc := NewService(repo)
	controller := NewController(svc, r)
	return &App{
		Controller: controller,
		wg:         &sync.WaitGroup{},
		cfg:        cfg,
	}
}

func (app *App) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	// Start the HTTP server
	wg.Add(1)
	go func() {
		defer wg.Done()
		app.Controller.Run(app.cfg.Address.Http, ctx)
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Shutdown signal received")

	cancel()

	wg.Wait()
	log.Println("Application gracefully stopped.")
}
