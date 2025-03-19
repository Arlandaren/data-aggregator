package person

import (
	"bufio"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"io"
	"io/ioutil"
	"mime/multipart"
	"service/internal/domains/person/models"
	"service/internal/infrastructure/utils"

	"github.com/xuri/excelize/v2"
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

func getValueFromRecord(record []string, columnIndexes map[string]int, field string) string {
	if index, ok := columnIndexes[field]; ok && index < len(record) {
		return record[index]
	}
	return ""
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

func (s *Service) ParseAndSaveJSON(ctx context.Context, file io.Reader) error {
	var persons []models.Person
	if err := json.NewDecoder(file).Decode(&persons); err != nil {
		return fmt.Errorf("failed to decode JSON: %w", err)
	}

	for _, person := range persons {
		if err := s.repo.SavePerson(ctx, person); err != nil {
			return fmt.Errorf("failed to save person: %w", err)
		}
	}
	return nil
}

func (s *Service) ParseAndSaveXLSX(ctx context.Context, file io.Reader) error {
	// Read file data into a byte slice
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read XLSX file: %w", err)
	}

	// Save data to a temporary file because excelize requires a file path
	tempFile, err := ioutil.TempFile("", "*.xlsx")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tempFile.Close()
	defer func() {
		_ = os.Remove(tempFile.Name()) // Remove the temp file
	}()

	if _, err := tempFile.Write(data); err != nil {
		return fmt.Errorf("failed to write to temp file: %w", err)
	}

	// Open the Excel file
	f, err := excelize.OpenFile(tempFile.Name())
	if err != nil {
		return fmt.Errorf("failed to open XLSX file: %w", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	// Get the first sheet
	sheetName := f.GetSheetName(1)
	if sheetName == "" {
		return errors.New("no sheets found in Excel file")
	}

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("failed to get rows from sheet: %w", err)
	}

	if len(rows) == 0 {
		return errors.New("Excel file is empty")
	}

	// Read headers
	headers := rows[0]

	// Create a map for storing column indices
	columnIndexes := make(map[string]int)

	// Keywords to search for columns
	keywords := map[string][]string{
		"Fio":       {"фамилия имя отчество", "фамилия имя", "имя", "fio", "full name", "name", "фио"},
		"Phone":     {"телефон", "номер телефона", "phone", "phone number"},
		"Snils":     {"снилс", "страховой номер индивидуального лицевого счёта", "страховой номер", "snils"},
		"Inn":       {"инн", "идентификационный номер налогоплательщика", "inn"},
		"Passport":  {"паспорт", "документ удостоверяющий личность", "passport", "identity document"},
		"BirthDate": {"дата рождения", "день рождения", "birth date", "birthday"},
		"Address":   {"адрес", "address"},
	}

	// Search for column indices by keywords
	for i, header := range headers {
		header = strings.ToLower(strings.TrimSpace(header)) // Normalize the header
		for field, keys := range keywords {
			for _, key := range keys {
				if strings.Contains(header, key) {
					columnIndexes[field] = i
					break
				}
			}
		}
	}

	// Process each row
	for _, row := range rows[1:] {
		person := models.Person{
			Fio:       getValueFromRecord(row, columnIndexes, "Fio"),
			Phone:     getValueFromRecord(row, columnIndexes, "Phone"),
			Snils:     getValueFromRecord(row, columnIndexes, "Snils"),
			Inn:       getValueFromRecord(row, columnIndexes, "Inn"),
			Passport:  getValueFromRecord(row, columnIndexes, "Passport"),
			BirthDate: getValueFromRecord(row, columnIndexes, "Birth"),
			Address:   getValueFromRecord(row, columnIndexes, "Address"),
		}

		// Save the person to the database
		if err := s.repo.SavePerson(ctx, person); err != nil {
			return fmt.Errorf("failed to save person: %w", err)
		}
	}

	return nil
}

func (s *Service) ParseAndSaveSQL(ctx context.Context, file io.Reader) error {
	// Read all data from the file
	sqlBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read SQL file: %w", err)
	}

	// Execute the SQL statements
	sqlStatements := string(sqlBytes)
	if err := s.repo.ExecuteSQL(ctx, sqlStatements); err != nil {
		return fmt.Errorf("failed to execute SQL statements: %w", err)
	}

	return nil
}

func (s *Service) ProcessFile(ctx context.Context, file multipart.File, filename string) error {
	// Determine file type based on the extension
	if strings.HasSuffix(strings.ToLower(filename), ".csv") {
		return s.ParseAndSaveCSV(ctx, file)
	} else if strings.HasSuffix(strings.ToLower(filename), ".json") {
		return s.ParseAndSaveJSON(ctx, file)
	} else if strings.HasSuffix(strings.ToLower(filename), ".xlsx") {
		return s.ParseAndSaveXLSX(ctx, file)
	} else if strings.HasSuffix(strings.ToLower(filename), ".sql") {
		return s.ParseAndSaveSQL(ctx, file)
	} else {
		return fmt.Errorf("unsupported file type")
	}
}

func (s *Service) FindPerson(ctx context.Context, field, value string) ([]models.Person, error) {
	return s.repo.FindPerson(ctx, field, value)
}

func (s *Service) ListPersons(ctx context.Context) ([]models.Person, error) {
	return s.repo.GetAllPersons(ctx)
}
