package person

import (
	"bufio"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"service/internal/domains/person/models"
	"service/internal/infrastructure/utils"
	"strings"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// detectDelimiter определяет разделитель в CSV-файле
func detectDelimiter(firstLine string) rune {
	delimiters := []rune{',', ';', '\t', '|'} // Возможные разделители
	maxCount := 0
	detectedDelimiter := ','

	for _, delimiter := range delimiters {
		count := strings.Count(firstLine, string(delimiter))
		if count > maxCount {
			maxCount = count
			detectedDelimiter = delimiter
		}
	}

	return detectedDelimiter
}

func (s *Service) ParseAndSaveCSV(ctx context.Context, file io.Reader) error {
	// Use bufio.Reader to read the file
	bufReader := bufio.NewReader(file)

	// Peek the first 4096 bytes without advancing the reader
	peekBytes, err := bufReader.Peek(4096)
	if err != nil && err != io.EOF {
		return fmt.Errorf("failed to peek into the CSV file: %w", err)
	}

	// Convert peeked bytes to string and split by newline to get the first line
	peekStr := string(peekBytes)
	lines := strings.SplitN(peekStr, "\n", 2)
	if len(lines) == 0 {
		return fmt.Errorf("empty CSV file")
	}
	firstLine := lines[0]

	// Detect the delimiter
	delimiter := detectDelimiter(firstLine)
	fmt.Printf("Detected delimiter: %q\n", delimiter)

	// Create CSV reader with the detected delimiter
	reader := csv.NewReader(bufReader)
	reader.Comma = delimiter
	reader.FieldsPerRecord = -1 // Allow variable number of fields per record

	// Читаем заголовки
	headers, err := reader.Read()
	if err != nil {
		return fmt.Errorf("failed to read headers: %w", err)
	}

	// Создаем мапу для хранения индексов столбцов
	columnIndexes := make(map[string]int)

	// Ключевые слова для поиска столбцов
	keywords := map[string][]string{
		"ID":       {"номер заявки", "номер заявления", "идентификатор", "id", "application number", "identifier"},
		"Fio":      {"фамилия имя отчество", "фамилия имя", "имя", "fio", "full name", "name"},
		"Phone":    {"телефон", "номер телефона", "phone", "phone number"},
		"Snils":    {"снилс", "страховой номер индивидуального лицевого счёта", "индивидуальный лицевой счет", "страховой номер", "snils", "insurance number"},
		"Inn":      {"инн", "идентификационный номер налогоплательщика", "inn", "taxpayer identification number"},
		"Passport": {"паспорт", "документ удостоверяющий личность", "passport", "identity document"},
		"Birth":    {"дата рождения", "день рождения", "birth date", "birthday"},
		"Address":  {"адрес", "address"},
	}

	// Ищем индексы столбцов по ключевым словам
	for i, header := range headers {
		header = strings.ToLower(header) // Приводим к нижнему регистру для унификации
		for field, keys := range keywords {
			for _, key := range keys {
				if strings.Contains(header, key) {
					columnIndexes[field] = i
					break
				}
			}
		}
	}

	// Читаем остальные строки
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read records: %w", err)
	}

	// Обрабатываем каждую запись
	for _, record := range records {
		fmt.Println(record)
		person := models.Person{
			Fio:       getValueFromRecord(record, columnIndexes, "Fio"),
			Phone:     getValueFromRecord(record, columnIndexes, "Phone"),
			Snils:     getValueFromRecord(record, columnIndexes, "Snils"),
			Inn:       getValueFromRecord(record, columnIndexes, "Inn"),
			Passport:  getValueFromRecord(record, columnIndexes, "Passport"),
			BirthDate: getValueFromRecord(record, columnIndexes, "Birth"),
			Address:   getValueFromRecord(record, columnIndexes, "Address"),
		}

		// Сохраняем запись в базу данных
		if err := s.repo.SavePerson(ctx, person); err != nil {
			return fmt.Errorf("failed to save person: %w", err)
		}
	}

	return nil
}

func (s *Service) ParseAndSaveCSVWithAi(ctx context.Context, file io.Reader) error {
	reader := csv.NewReader(file)
	reader.Comma = ','
	reader.FieldsPerRecord = -1 // Разрешаем разное количество полей в строках

	// Читаем заголовки
	headers, err := reader.Read()
	if err != nil {
		return fmt.Errorf("failed to read headers: %w", err)
	}

	// Используем функцию CheckFields для определения соответствий
	result, err := utils.CheckFields(headers)
	if err != nil {
		return fmt.Errorf("failed to check fields: %w", err)
	}

	// Создаем мапу для хранения индексов столбцов
	columnIndexes := make(map[string]int)

	// Используем мапу result для сопоставления полей
	for i, header := range headers {
		header = strings.ToLower(header) // Приводим к нижнему регистру для унификации
		for dbField, csvField := range result {
			if strings.Contains(header, strings.ToLower(csvField)) {
				columnIndexes[dbField] = i
				break
			}
		}
	}

	// Читаем остальные строки
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read records: %w", err)
	}

	// Обрабатываем каждую запись
	for _, record := range records {
		person := models.Person{
			Fio:       getValueFromRecord(record, columnIndexes, "Fio"),
			Phone:     getValueFromRecord(record, columnIndexes, "Phone"),
			Snils:     getValueFromRecord(record, columnIndexes, "Snils"),
			Inn:       getValueFromRecord(record, columnIndexes, "Inn"),
			Passport:  getValueFromRecord(record, columnIndexes, "Passport"),
			BirthDate: getValueFromRecord(record, columnIndexes, "Birth"),
			Address:   getValueFromRecord(record, columnIndexes, "Address"),
		}

		// Сохраняем запись в базу данных
		if err := s.repo.SavePerson(ctx, person); err != nil {
			return fmt.Errorf("failed to save person: %w", err)
		}
	}

	return nil
}

func getValueFromRecord(record []string, columnIndexes map[string]int, field string) string {
	if index, ok := columnIndexes[field]; ok && index < len(record) {
		return record[index]
	}
	return ""
}

func (s *Service) ParseAndSaveJSON(ctx context.Context, file io.Reader) error {
	var persons []models.Person
	if err := json.NewDecoder(file).Decode(&persons); err != nil {
		return err
	}

	for _, person := range persons {
		if err := s.repo.SavePerson(ctx, person); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) FindPerson(ctx context.Context, field, value string) ([]models.Person, error) {
	return s.repo.FindPerson(ctx, field, value)
}

func (s *Service) ListPersons(ctx context.Context) ([]models.Person, error) {
	return s.repo.GetAllPersons(ctx)
}
