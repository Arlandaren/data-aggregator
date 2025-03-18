package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
	"io"
	"net/http"
	"service/internal/infrastructure/config"
	"strings"
)

func CheckFields(fields []string) (map[string]string, error) {
	cfg := config.GetTogether()

	prompt := fmt.Sprintf("%s", strings.Join(fields, ","))

	requestBody, err := json.Marshal(map[string]interface{}{
		"model": "meta-llama/Meta-Llama-3.1-8B-Instruct-Turbo-128K",
		"messages": []map[string]string{
			{
				"role": "system", "content": "Ты работник для программы унификации данных, твоя задача определять какие поля присутствуют в базе данных их заданного массива. В базе данных есть поля - ID,Fio,Phone,Snils,Inn,Passport,Birth,Address. Тебе на вход будут подавать ключевые слова из импортируемых данных, через запятую. Твоя задача проверить подходят ли эти поля под колонки в базе данных. В ответ верни список разделенный запятыми ключевых слов которе подойдут под нашу бд, ВАЖНО!!! без \n переносов и прочего форматировани. Только требуемый ответ, как из примера, указывай какому полю из бд соответствует заданное ключевое слово. Если ключевое слово не подходит ни под одно поле бд, пропускай его. Например Запрос- name,adres,date_of_birth,passport,number Правильным ответом будет - phone:number,passport:passport,fio:name,birth:date_of_birth,address:adres.",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
	})
	if err != nil {
		fmt.Println("Error marshalling request body:", err)
		return nil, err
	}

	request, err := http.NewRequest("POST", cfg.URL, bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		fmt.Println("Error making request:", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("ai model currently unavailable")
	}

	content := gjson.GetBytes(body, "choices.0.message.content")
	if !content.Exists() {
		return nil, errors.New("ai model currently unavailable")
	}
	fmt.Println(prompt)
	fmt.Println(content.Str)
	input := content.Str

	input = strings.ReplaceAll(input, "\n", "")      // Убираем \n
	input = strings.TrimSpace(input)                 // Убираем пробелы в начале и конце
	input = strings.Join(strings.Fields(input), " ") // Убираем лишние пробелы внутри строки

	pairs := strings.Split(input, ",")

	result := make(map[string]string)

	// Обрабатываем каждую пару
	for _, pair := range pairs {
		// Пропускаем пустые пары
		if pair == "" {
			continue
		}

		// Разделяем пару по двоеточию
		parts := strings.Split(pair, ":")
		if len(parts) == 2 { // Убедимся, что пара корректна
			key := strings.TrimSpace(parts[0])   // Ключ (например, "phone")
			value := strings.TrimSpace(parts[1]) // Значение (например, "number")

			// Пропускаем дубликаты ключей
			if _, exists := result[key]; exists {
				fmt.Printf("Дубликат ключа: %s\n", key)
				continue
			}

			// Сохраняем пару в мапу
			result[key] = value
		}
	}
	fmt.Println(result)
	return result, nil
}
