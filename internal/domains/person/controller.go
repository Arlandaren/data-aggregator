package person

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Controller struct {
	svc *Service
}

func NewController(svc *Service) *Controller {
	return &Controller{
		svc: svc,
	}
}

func (c *Controller) Endpoints(r *gin.Engine) {
	r.POST("/person/upload", c.UploadFile)
	r.GET("/person/find", c.FindPerson)
	r.POST("/person/upload/ai/csv", c.UploadCSVWithAi)
	r.GET("/persons", c.ListPersons)
}

func (c *Controller) UploadFile(ctx *gin.Context) {
	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "File not found"})
		return
	}
	defer file.Close()

	filename := header.Filename

	// Process the file based on its type
	if err := c.svc.ProcessFile(ctx.Request.Context(), file, filename); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "File successfully uploaded and processed"})
}

func (c *Controller) UploadCSVWithAi(ctx *gin.Context) {
	file, _, err := ctx.Request.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Файл не найден"})
		return
	}
	defer file.Close()

	if err := c.svc.ParseAndSaveCSVWithAi(ctx.Request.Context(), file); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "CSV файл успешно загружен и обработан"})
}

func (c *Controller) FindPerson(ctx *gin.Context) {
	field := ctx.Query("field")
	value := ctx.Query("value")

	if value == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Значение обязательно к указанию"})
		return
	}

	persons, err := c.svc.FindPerson(ctx.Request.Context(), field, value)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"persons": persons})
}

func (c *Controller) ListPersons(ctx *gin.Context) {
	persons, err := c.svc.ListPersons(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"persons": persons})
}
