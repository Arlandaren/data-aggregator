package application

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
	"service/internal/domains/api"
	"service/internal/domains/person"
	"time"
)

type Controller struct {
	person *person.Controller
	api    *api.Controller
	Router *gin.Engine
}

func NewController(svc *Service, r *gin.Engine) *Controller {
	return &Controller{
		person: person.NewController(svc.Person),
		api:    api.NewController(svc.Api),
		Router: r,
	}
}

func (c *Controller) InitRouter() {
	c.api.Endpoints(c.Router)
	c.person.Endpoints(c.Router)
}

func (c *Controller) Run(addr string, ctx context.Context) {
	c.InitRouter()

	server := &http.Server{
		Addr:    addr,
		Handler: c.Router,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Errorf("HTTP server exited with error: %v", err)
		}
	}()
	log.Printf("HTTP server listening at %v", addr)

	<-ctx.Done()
	log.Println("Shutting down HTTP server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Errorf("HTTP server Shutdown Failed:%+v", err)
	} else {
		log.Println("HTTP server gracefully stopped")
	}
}
