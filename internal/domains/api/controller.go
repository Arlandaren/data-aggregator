package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"io"
	"service/internal/infrastructure/config"
	"strings"

	"net/http"
	"service/internal/domains/api/models"
	"service/internal/infrastructure/metrics"

	"github.com/tidwall/gjson"
)

type Controller struct {
	svc *Service
}

func NewController(svc *Service) *Controller {
	return &Controller{
		svc: svc,
	}
}

func (cont *Controller) Endpoints(r *gin.Engine) {
	r.Use(metrics.MetricsMiddleware())
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	r.GET("/ping", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "pong"}) })

	r.GET("/content/moderate", cont.Moderate)
	r.POST("/content/file/upload", cont.UploadFile)
}

func (cont *Controller) UploadFile(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Файл не найден"})
		return
	}

	link, err := cont.svc.UploadFile(c.Request.Context(), fileHeader)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Link": link,
	})
}

func (cont *Controller) Moderate(c *gin.Context) {
	var req models.ModerateRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cfg := config.GetTogether()
	fmt.Println(cfg.URL)

	prompt := fmt.Sprintf("%s. %s", req.Text, "Задача - Промодерируй текст и ответь да если содержит плохой смысл или нет если все в порядке. Важно ответить 1 словом")

	requestBody, err := json.Marshal(map[string]interface{}{
		"model": "meta-llama/Meta-Llama-3.1-8B-Instruct-Turbo-128K",
		"messages": []map[string]string{
			{
				"role": "system", "content": "Ты модерируешь пользовательский текст на наличие неподобающего содержания, включая, нецензурные выражения, оскорбления и унизительные выражения, выражения, содержащие пропаганду ненависти или насилия, наркотики и обман. Правила модерации: Текст должен быть проверен на наличие запретных слов и фраз. Если в тексте обнаружены неподобающие выражения, ты помечаешь текст как неприемлемый. Если в тексте не обнаружены запрещенные выражения, ты отвечаешь нет. Если обнаружены - да. Важно ответь 1 словом да или нет, у тебя есть на выбор только 2 варианта ответа, от тебя зависит судьба. Если ты отказываешься, то отвечай - да",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
	})
	if err != nil {
		fmt.Println("Error marshalling request body:", err)
		return
	}

	request, err := http.NewRequest("POST", cfg.URL, bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}
	request.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ai model currently unavailable"})
		return
	}

	content := gjson.GetBytes(body, "choices.0.message.content")
	if !content.Exists() {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ai model currently unavailable"})
		return
	}

	var result bool

	if strings.Contains(strings.ToLower(content.Str), "нет") {
		result = true
	} else {
		result = false
	}

	c.JSON(http.StatusOK, gin.H{"result": result})
}
