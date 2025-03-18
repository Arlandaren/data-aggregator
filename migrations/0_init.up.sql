CREATE TABLE persons (
                        id SERIAL PRIMARY KEY,
                        fio TEXT,
                        phone TEXT,
                        snils TEXT,
                        inn TEXT,
                        passport TEXT,
                        birth_date DATE,
                        address TEXT
);
ALTER TABLE persons ADD CONSTRAINT unique_person UNIQUE (fio, phone, snils, inn, passport, birth_date, address);