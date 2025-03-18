package models

type Person struct {
	Id        int    `db:"id" json:"id"`
	Fio       string `db:"fio" json:"fio"`
	Phone     string `db:"phone" json:"phone"`
	Snils     string `db:"snils" json:"snils"`
	Inn       string `db:"inn" json:"inn"`
	Passport  string `db:"passport" json:"passport"`
	BirthDate string `db:"birth_date" json:"birth_date"`
	Address   string `db:"address" json:"address"`
}
