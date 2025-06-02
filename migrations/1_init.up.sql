CREATE TABLE IF NOT EXISTS gender
(
    id SERIAL PRIMARY KEY,
    gender_name TEXT UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS nationality
(
    id SERIAL PRIMARY KEY,
    nationality_name TEXT UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS people_info
(
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    surname TEXT NOT NULL,
    patronym TEXT,
    age INTEGER NOT NULL,
    gender_id INTEGER NOT NULL,
    nationality_id INTEGER NOT NULL,
    FOREIGN KEY (gender_id) REFERENCES gender(id) ON DELETE RESTRICT,
    FOREIGN KEY (nationality_id) REFERENCES nationality(id) ON DELETE RESTRICT
);
CREATE INDEX IF NOT EXISTS idx_name_surname ON people_info (name, surname);