package main

import (
	"database/sql"
	"errors"

	_ "modernc.org/sqlite"
)

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

func (s ParcelStore) Add(p Parcel) (int, error) {
	res, err := s.db.Exec(
		"INSERT INTO parcel (client, status, address, created_at) VALUES (?, ?, ?, ?)",
		p.Client, p.Status, p.Address, p.CreatedAt,
	)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (s ParcelStore) Get(number int) (Parcel, error) {
	p := Parcel{}
	row := s.db.QueryRow(
		"SELECT number, client, status, address, created_at FROM parcel WHERE number = ?",
		number,
	)
	err := row.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	if err != nil {
		return p, err
	}
	return p, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	var res []Parcel
	rows, err := s.db.Query(
		"SELECT number, client, status, address, created_at FROM parcel WHERE client = ?",
		client,
	)
	if err != nil {
		return res, err
	}
	defer rows.Close()

	for rows.Next() {
		var p Parcel
		err := rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
		if err != nil {
			return res, err
		}
		res = append(res, p)
	}

	if err = rows.Err(); err != nil {
		return res, err
	}

	return res, nil
}

func (s ParcelStore) SetStatus(number int, status string) error {
	_, err := s.db.Exec(
		"UPDATE parcel SET status = ? WHERE number = ?",
		status, number,
	)
	return err
}

func (s ParcelStore) SetAddress(number int, address string) error {
	var status string
	row := s.db.QueryRow("SELECT status FROM parcel WHERE number = ?", number)
	err := row.Scan(&status)
	if err != nil {
		return err
	}

	if status != ParcelStatusRegistered {
		return errors.New("можно менять адрес только для зарегистрированных посылок")
	}

	_, err = s.db.Exec(
		"UPDATE parcel SET address = ? WHERE number = ?",
		address, number,
	)
	return err
}

func (s ParcelStore) Delete(number int) error {
	var status string
	row := s.db.QueryRow("SELECT status FROM parcel WHERE number = ?", number)
	err := row.Scan(&status)
	if err != nil {
		return err
	}

	if status != ParcelStatusRegistered {
		return errors.New("можно удалять только зарегистрированные посылки")
	}

	_, err = s.db.Exec("DELETE FROM parcel WHERE number = ?", number)
	return err
}
