package main

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) Store {
	return Store{db: db}
}

func (s Store) Add(task Task) (int, error) {
	res, err := s.db.Exec("INSERT INTO scheduler (date, title, comment, repeat) VALUES (:date, :title, :comment, :repeat)",
		sql.Named("date", task.Date),
		sql.Named("title", task.Title),
		sql.Named("comment", task.Comment),
		sql.Named("repeat", task.Repeat))
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (s Store) GetAll() ([]Task, error) {
	rows, err := s.db.Query("SELECT id, date, title, comment, repeat FROM scheduler order by date limit 50")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []Task
	for rows.Next() {
		t := Task{}

		err := rows.Scan(&t.Id, &t.Date, &t.Title, &t.Comment, &t.Repeat)
		if err != nil {
			return nil, err
		}

		res = append(res, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

func (s Store) GetByDate(date time.Time) ([]Task, error) {
	rows, err := s.db.Query("SELECT id, date, title, comment, repeat FROM scheduler WHERE date = :date limit 50",
		sql.Named("date", date.Format(DATE_FORMAT)))

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []Task
	for rows.Next() {
		t := Task{}

		err := rows.Scan(&t.Id, &t.Date, &t.Title, &t.Comment, &t.Repeat)
		if err != nil {
			return nil, err
		}

		res = append(res, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

func (s Store) GetById(id int) (Task, error) {
	t := Task{}

	row := s.db.QueryRow("SELECT id, date, title, comment, repeat FROM scheduler WHERE id = :id",
		sql.Named("id", id))

	err := row.Scan(&t.Id, &t.Date, &t.Title, &t.Comment, &t.Repeat)
	if err != nil {
		return t, err
	}

	return t, nil
}

func (s Store) GetByTitle(search string) ([]Task, error) {
	rows, err := s.db.Query("SELECT id, date, title, comment, repeat FROM scheduler WHERE UPPER(title) LIKE :search OR UPPER(comment) LIKE :search ORDER BY date LIMIT 50",
		sql.Named("search", fmt.Sprintf("%%%s%%", strings.ToUpper(search))))

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []Task
	for rows.Next() {
		t := Task{}

		err := rows.Scan(&t.Id, &t.Date, &t.Title, &t.Comment, &t.Repeat)
		if err != nil {
			return nil, err
		}

		res = append(res, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

func (s Store) Update(task Task) error {
	_, err := s.db.Exec("UPDATE scheduler SET date = :date, title = :title, comment = :comment, repeat = :repeat WHERE id = :id",
		sql.Named("id", task.Id),
		sql.Named("date", task.Date),
		sql.Named("title", task.Title),
		sql.Named("comment", task.Comment),
		sql.Named("repeat", task.Repeat))

	return err
}

func (s Store) Delete(id int) error {
	_, err := s.db.Exec("DELETE FROM scheduler WHERE id = :id",
		sql.Named("id", id))

	return err
}
