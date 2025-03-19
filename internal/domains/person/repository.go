package person

import (
	"context"
	"fmt"
	"github.com/Arlandaren/pgxWrappy/pkg/postgres"
	"log"
	"service/internal/domains/person/models"
	"service/internal/infrastructure/storage/redis"
	"strings"
)

type Repository struct {
	db  *postgres.Wrapper
	rdb *redis.RDB
}

func NewRepository(db *postgres.Wrapper, rdb *redis.RDB) *Repository {
	return &Repository{
		db:  db,
		rdb: rdb,
	}
}

func (r *Repository) SavePerson(ctx context.Context, person models.Person) error {
	query := `
        INSERT INTO persons (fio, phone, snils, inn, passport, birth_date,address)
        VALUES ($1, $2, $3, $4, $5, $6,$7)
        ON CONFLICT (fio, phone, snils, inn, passport, birth_date,address) DO NOTHING`

	var birthDate interface{} = person.BirthDate
	if person.BirthDate == "" {
		birthDate = nil
	}

	_, err := r.db.Exec(ctx, query, person.Fio, person.Phone, person.Snils, person.Inn, person.Passport, birthDate, person.Address)
	if err != nil {
		return fmt.Errorf("failed to save person: %w", err)
	}
	return nil
}

func (r *Repository) FindPerson(ctx context.Context, field, value string) ([]models.Person, error) {
	var persons []models.Person

	// Очищаем значение от лишних кавычек и пробелов
	value = strings.Trim(value, `"' `)

	// Формируем шаблон поиска
	searchPattern := "%" + value + "%"
	log.Printf("Search pattern: '%s'", searchPattern)

	// Допустимые поля для поиска
	validFields := map[string]bool{
		"fio":        true,
		"phone":      true,
		"snils":      true,
		"inn":        true,
		"passport":   true,
		"address":    true,
		"birth_date": true,
	}

	var query string
	var args []interface{}

	if field != "" {
		// Проверяем, что указано допустимое поле
		log.Println(field)
		if !validFields[field] {
			return nil, fmt.Errorf("invalid field: %s", field)
		}

		// Особая обработка для поля birth_date
		if field == "birth_date" {
			query = `
                SELECT
                    id,
                    fio,
                    phone,
                    snils,
                    inn,
                    passport,
                    birth_date,
                    address
                FROM persons
                WHERE birth_date ILIKE $1`
		} else {
			query = fmt.Sprintf(`
                SELECT
                    id,
                    fio,
                    phone,
                    snils,
                    inn,
                    passport,
                    birth_date,
                    address
                FROM persons
                WHERE %s ILIKE $1`, field)
		}
		args = []interface{}{searchPattern}
	} else {
		// Поиск по всем полям
		conditions := []string{
			"fio ILIKE $1",
			"phone ILIKE $1",
			"snils ILIKE $1",
			"inn ILIKE $1",
			"passport ILIKE $1",
			"address ILIKE $1",
			"birth_date ILIKE $1",
		}

		query = fmt.Sprintf(`
            SELECT
                id,
                fio,
                phone,
                snils,
                inn,
                passport,
			 	COALESCE(birth_date, '') as birth_date,
                address
            FROM persons
            WHERE %s`, strings.Join(conditions, " OR "))

		args = []interface{}{searchPattern}
	}

	log.Printf("Executing query:\n%s\nWith params: %v", query, args)

	// Выполняем запрос
	if err := r.db.Select(ctx, &persons, query, args...); err != nil {
		log.Printf("Query error: %v", err)
		return nil, fmt.Errorf("query error: %w", err)
	}

	log.Printf("Found %d records", len(persons))
	for _, p := range persons {
		log.Printf("Person: %+v", p)
	}

	return persons, nil
}

func (r *Repository) GetAllPersons(ctx context.Context) ([]models.Person, error) {
	var persons []models.Person
	if err := r.db.Select(ctx, &persons, "SELECT id, fio, phone, snils, inn, passport, COALESCE(birth_date, '') as birth_date , address FROM persons"); err != nil {
		return nil, fmt.Errorf("failed to query persons: %w", err)
	}
	return persons, nil
}

func (r *Repository) ExecuteSQL(ctx context.Context, sqlStatements string) error {
	_, err := r.db.Pool.Exec(ctx, sqlStatements)
	return err
}
