package api_test

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"net/http"
	"service/internal/application"
	"service/internal/domains/api/models"
	"service/internal/infrastructure/config"
	"testing"
	"time"
)

func TestModerate(t *testing.T) {
	err := godotenv.Load("../../../.env")
	if err != nil {
		t.Fatalf("failed to load .env file: %v", err)
	}

	gin.SetMode(gin.TestMode)
	router := gin.Default()

	cfg := config.NewConfig()
	cfg.Address = &config.Address{Http: ":8079"}

	app := application.NewApp(nil, nil, nil, router, cfg)

	go func() {
		app.Run()
	}()

	time.Sleep(100 * time.Millisecond)

	doRequest := func(inputText string) (map[string]interface{}, int, error) {
		reqBody := models.ModerateRequest{Text: inputText}
		jsonBytes, err := json.Marshal(reqBody)
		if err != nil {
			return nil, 0, err
		}
		req, err := http.NewRequest(http.MethodGet, "http://localhost:8079/content/moderate", bytes.NewBuffer(jsonBytes))
		if err != nil {
			return nil, 0, err
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, 0, err
		}
		defer resp.Body.Close()

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		return result, resp.StatusCode, err
	}

	resp, code, err := doRequest("иди нахуй хуеплет")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, false, resp["result"])

	resp, code, err = doRequest("привет генка")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, true, resp["result"])
}
