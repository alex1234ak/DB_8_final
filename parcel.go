package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

func (s ParcelStore) Add(p Parcel) (int, error) {
	// реализуйте добавление строки в таблицу parcel, используйте данные из переменной p
	res, err := s.db.Exec("INSERT INTO parcel (client, status, address, created_at) VALUES (:client, :status, :address,:created_at)",
		sql.Named("client", p.Client),
		sql.Named("status", p.Status),
		sql.Named("address", p.Address),
		sql.Named("created_at", p.CreatedAt))
	if err != nil {
		return 0, err
	}

	// верните идентификатор последней добавленной записи
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

// func (s ParcelStore) Exec(param1 string, arg sql.NamedArg, param3 sql.NamedArg, param4 sql.NamedArg, param5 sql.NamedArg) (any, any) {
// 	panic("unimplemented")
// }

func (s ParcelStore) Get(number int) (Parcel, error) {
	// реализуйте чтение строки по заданному number
	// здесь из таблицы должна вернуться только одна строка
	p := Parcel{}

	row := s.db.QueryRow(`SELECT number, client, status, address, created_at FROM parcel WHERE number = ?`, number)
	err := row.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	if err != nil {
		return Parcel{}, err
	}
	return p, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	// реализуйте чтение строк из таблицы parcel по заданному client
	// здесь из таблицы может вернуться несколько строк

	var res []Parcel

	rows, err := s.db.Query(`SELECT number, client, status, address, created_at FROM parcel WHERE client = ?`, client)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var p Parcel

		err := rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
		if err != nil {
			return nil, err
		}

		res = append(res, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

// func (s ParcelStore) Query(param1 string, client int) (any, any) {
// 	panic("unimplemented")
// }

func (s ParcelStore) SetStatus(number int, status string) error {
	// реализуйте обновление статуса в таблице parcel
	_, err := s.db.Exec("UPDATE parcel SET status = :status WHERE number = :number",
		sql.Named("status", status),
		sql.Named("number", number))
	if err != nil {
		fmt.Println("Ошибка при обновлении статуса:", err)
		log.Println(err)
	}
	return nil
}

func (s ParcelStore) SetAddress(number int, address string) error {
	// реализуйте обновление адреса в таблице parcel
	// менять адрес можно только если значение статуса registered
	var status string

	row := s.db.QueryRow("SELECT status FROM parcel WHERE number = :number",
		sql.Named("number", number))

	err := row.Scan(&status)
	if err != nil {
		fmt.Println("Ошибка при получении статуса:", err)
		return nil
	}

	if status != "registered" {
		fmt.Println("Изменение адреса невозможно: посылка уже отправлена")
		return nil
	}

	_, err = s.db.Exec("UPDATE parcel SET address = :address WHERE number = :number",
		sql.Named("address", address),
		sql.Named("number", number))

	if err != nil {
		fmt.Println("Ошибка при обновлении:", err)
	}

	return nil
}

func (s ParcelStore) Delete(number int) error {
	// реализуйте удаление строки из таблицы parcel
	// удалять строку можно только если значение статуса registered
	var status string

	row := s.db.QueryRow("SELECT status FROM parcel WHERE number = :number",
		sql.Named("number", number))

	err := row.Scan(&status)
	if err != nil {
		fmt.Println("Ошибка при получении статуса:", err)
		return nil
	}

	if status != "registered" {
		fmt.Println("Удаление невозможно: посылка уже отправлена")
		return nil
	}

	_, err = s.db.Exec(
		"DELETE FROM parcel WHERE number = :number",
		sql.Named("number", number))

	if err != nil {
		fmt.Println("Ошибка при удалении:", err)
	}

	return nil
}
