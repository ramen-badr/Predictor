package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"predictor/internal/config"
	"predictor/internal/domain/models"
	"predictor/internal/storage"
	"strings"
)

type Storage struct {
	db *sql.DB
}

func New(cfgStorage config.Storage) (*Storage, error) {
	const op = "storage.postgres.New"

	db, err := sql.Open(
		"pgx",
		fmt.Sprintf("postgres://%s:%s@%s/%s", cfgStorage.User, cfgStorage.Password, cfgStorage.Address, cfgStorage.Name),
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveNationality(nationality string) (int64, error) {
	const op = "storage.postgres.SaveNationality"

	var id int64

	if err := s.db.QueryRow("SELECT id FROM nationality WHERE nationality_name = $1", nationality).Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = s.db.QueryRow(`
				INSERT INTO nationality (nationality_name)
				VALUES ($1)
				RETURNING id
			`, nationality).Scan(&id)

			if err != nil {
				return 0, fmt.Errorf("%s: %w", op, err)
			}

			return id, nil
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) SaveGender(gender string) (int64, error) {
	const op = "storage.postgres.SaveGender"

	var id int64

	if err := s.db.QueryRow("SELECT id FROM gender WHERE gender_name = $1", gender).Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = s.db.QueryRow(`
				INSERT INTO gender (gender_name)
				VALUES ($1)
				RETURNING id
			`, gender).Scan(&id)

			if err != nil {
				return 0, fmt.Errorf("%s: %w", op, err)
			}

			return id, nil
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) GetGender(gender string) (int64, error) {
	const op = "storage.postgres.GetGender"

	var id int64

	if err := s.db.QueryRow("SELECT id FROM gender WHERE gender_name = $1", gender).Scan(&id); err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) GetNationality(nationality string) (int64, error) {
	const op = "storage.postgres.GetNationality"

	var id int64

	if err := s.db.QueryRow("SELECT id FROM nationality WHERE nationality_name = $1", nationality).Scan(&id); err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) SavePeople(name, surname, patronym, gender, nationality string, age int) (int64, error) {
	const op = "storage.postgres.SavePeople"

	genderId, err := s.SaveGender(gender)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	nationalityId, err := s.SaveNationality(nationality)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	var id int64

	if err = s.db.QueryRow(`
		INSERT INTO people_info(name, surname, patronym, age, gender_id, nationality_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, name, surname, patronym, age, genderId, nationalityId).Scan(&id); err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) DeletePeople(id int64) error {
	const op = "storage.postgres.DeletePeople"

	stmt, err := s.db.Prepare("DELETE FROM people_info WHERE id = $1")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec(id)
	if errors.Is(err, sql.ErrNoRows) {
		return storage.ErrPeopleNotFound
	}
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) UpdatePeople(name, surname, patronym, gender, nationality string, age int, id int64) error {
	const op = "storage.postgres.UpdatePeople"

	query := "UPDATE people_info SET"
	var args []any

	if name != "" {
		query += fmt.Sprintf(" name = $%d,", len(args)+1)
		args = append(args, name)
	}

	if surname != "" {
		query += fmt.Sprintf(" surname = $%d,", len(args)+1)
		args = append(args, surname)
	}

	if patronym != "" {
		query += fmt.Sprintf(" patronym = $%d, =", len(args)+1)
		args = append(args, patronym)
	}

	if age != 0 {
		query += fmt.Sprintf(" age = $%d,", len(args)+1)
		args = append(args, age)
	}

	if gender != "" {
		genderId, err := s.SaveGender(gender)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		query += fmt.Sprintf(" gender_id = $%d,", len(args)+1)
		args = append(args, genderId)
	}

	if nationality != "" {
		nationalityId, err := s.SaveNationality(nationality)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		query += fmt.Sprintf(" nationality_id = $%d,", len(args)+1)
		args = append(args, nationalityId)
	}

	if args != nil {
		query = strings.TrimSuffix(query, ",")
		query += fmt.Sprintf(" WHERE id = $%d", len(args)+1)
		args = append(args, id)

		stmt, err := s.db.Prepare(query)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		_, err = stmt.Exec(args...)
		if errors.Is(err, sql.ErrNoRows) {
			return storage.ErrPeopleNotFound
		}
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	return nil
}

func (s *Storage) GetPeople(limit, offset int64, name, surname, patronym, gender, nationality string, age int) ([]models.People, int64, error) {
	const op = "storage.postgres.GetPeople"

	query := `
        SELECT name, surname, patronym, age, gender_name, nationality_name
        FROM people_info INNER JOIN gender ON gender_id = gender.id INNER JOIN nationality ON nationality_id = nationality.id
    `

	queryForTotal := "SELECT COUNT(*) FROM people_info"

	var args []any
	var cond []string

	if name != "" {
		cond = append(cond, fmt.Sprintf("name = $%d", len(args)+1))
		args = append(args, name)
	}

	if surname != "" {
		cond = append(cond, fmt.Sprintf("surname = $%d", len(args)+1))
		args = append(args, surname)
	}

	if patronym != "" {
		cond = append(cond, fmt.Sprintf("patronym = $%d", len(args)+1))
		args = append(args, patronym)
	}

	if age != 0 {
		cond = append(cond, fmt.Sprintf("age = $%d", len(args)+1))
		args = append(args, age)
	}

	if gender != "" {
		genderId, err := s.GetGender(gender)
		if err != nil {
			return nil, 0, fmt.Errorf("%s: %w", op, err)
		}

		cond = append(cond, fmt.Sprintf("gender_id = $%d", len(args)+1))
		args = append(args, genderId)
	}

	if nationality != "" {
		nationalityId, err := s.GetNationality(nationality)
		if err != nil {
			return nil, 0, fmt.Errorf("%s: %w", op, err)
		}

		cond = append(cond, fmt.Sprintf("nationality_id = $%d", len(args)+1))
		args = append(args, nationalityId)
	}

	if args != nil {
		queryForTotal += " WHERE " + strings.Join(cond, " AND ")
		query += " WHERE " + strings.Join(cond, " AND ")
	}

	var total int64

	err := s.db.QueryRow(queryForTotal, args...).Scan(&total)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, 0, storage.ErrPeopleNotFound
	}
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", op, err)
	}

	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
	args = append(args, limit, offset)

	rows, err := s.db.Query(query, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, 0, storage.ErrPeopleNotFound
	}
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var people []models.People

	for rows.Next() {
		var p models.People

		if err = rows.Scan(
			&p.Name,
			&p.Surname,
			&p.Patronymic,
			&p.Age,
			&p.Gender,
			&p.Nationality,
		); err != nil {
			return nil, 0, fmt.Errorf("%s: %w", op, err)
		}

		people = append(people, p)
	}

	return people, total, nil
}
