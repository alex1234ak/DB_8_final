package main

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

const (
	ParcelStatusRegistered = "registered"
	ParcelStatusSent       = "sent"
	ParcelStatusDelivered  = "delivered"
)

type Parcel struct {
	Number    int
	Client    int
	Status    string
	Address   string
	CreatedAt string
}

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

type ParcelService struct {
	store ParcelStore
}

func NewParcelService(store ParcelStore) ParcelService {
	return ParcelService{store: store}
}

func (s ParcelService) Register(client int, address string) (Parcel, error) {
	parcel := Parcel{
		Client:    client,
		Status:    ParcelStatusRegistered,
		Address:   address,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	id, err := s.store.Add(parcel)
	if err != nil {
		return parcel, err
	}

	parcel.Number = id

	fmt.Printf("Новая посылка № %d на адрес %s от клиента с идентификатором %d зарегистрирована %s\n",
		parcel.Number, parcel.Address, parcel.Client, parcel.CreatedAt)

	return parcel, nil
}

func (s ParcelService) PrintClientParcels(client int) error {
	parcels, err := s.store.GetByClient(client)
	if err != nil {
		return err
	}

	fmt.Printf("Посылки клиента %d:\n", client)

	for _, parcel := range parcels {
		fmt.Printf("Посылка № %d на адрес %s от клиента с идентификатором %d зарегистрирована %s, статус %s\n",
			parcel.Number, parcel.Address, parcel.Client, parcel.CreatedAt, parcel.Status)
	}
	fmt.Println()

	return nil
}

func (s ParcelService) NextStatus(number int) error {
	parcel, err := s.store.Get(number)
	if err != nil {
		return err
	}

	var nextStatus string
	switch parcel.Status {
	case ParcelStatusRegistered:
		nextStatus = ParcelStatusSent
	case ParcelStatusSent:
		nextStatus = ParcelStatusDelivered
	case ParcelStatusDelivered:
		return nil
	}

	fmt.Printf("У посылки № %d новый статус: %s\n", number, nextStatus)

	return s.store.SetStatus(number, nextStatus)
}

func (s ParcelService) ChangeAddress(number int, address string) error {
	return s.store.SetAddress(number, address)
}

func (s ParcelService) Delete(number int) error {
	return s.store.Delete(number)
}

func main() {
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		fmt.Println("Ошибка подключения к БД:", err)
		return
	}
	defer db.Close()

	store := NewParcelStore(db)
	service := NewParcelService(store)

	client := 1
	address := "Псков, д. Пушкина, ул. Колотушкина, д. 5"
	p, err := service.Register(client, address)
	if err != nil {
		fmt.Println("Ошибка регистрации:", err)
		return
	}

	newAddress := "Саратов, д. Верхние Зори, ул. Козлова, д. 25"
	err = service.ChangeAddress(p.Number, newAddress)
	if err != nil {
		fmt.Println("Ошибка изменения адреса:", err)
		return
	}

	err = service.NextStatus(p.Number)
	if err != nil {
		fmt.Println("Ошибка изменения статуса:", err)
		return
	}

	err = service.PrintClientParcels(client)
	if err != nil {
		fmt.Println("Ошибка получения посылок:", err)
		return
	}

	err = service.Delete(p.Number)
	if err != nil {
		fmt.Println("Ошибка удаления (ожидаемо):", err)
	}

	err = service.PrintClientParcels(client)
	if err != nil {
		fmt.Println("Ошибка получения посылок:", err)
		return
	}

	p, err = service.Register(client, address)
	if err != nil {
		fmt.Println("Ошибка регистрации:", err)
		return
	}

	err = service.Delete(p.Number)
	if err != nil {
		fmt.Println("Ошибка удаления:", err)
	}

	err = service.PrintClientParcels(client)
	if err != nil {
		fmt.Println("Ошибка получения посылок:", err)
		return
	}
}
