
package main

import (
    "database/sql"
    "errors"
)

type ParcelStore struct {
    db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
    return ParcelStore{db: db}
}

func (s ParcelStore) Add(p Parcel) (int, error) {
    query := `INSERT INTO parcel (client, status, address, created_at) VALUES (?, ?, ?, ?)`
    res, err := s.db.Exec(query, p.Client, p.Status, p.Address, p.CreatedAt)
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
    query := `SELECT number, client, status, address, created_at FROM parcel WHERE number = ?`
    row := s.db.QueryRow(query, number)
    p := Parcel{}
    err := row.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
    if err != nil {
        return Parcel{}, err
    }
    return p, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
    query := `SELECT number, client, status, address, created_at FROM parcel WHERE client = ?`
    rows, err := s.db.Query(query, client)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var parcels []Parcel
    for rows.Next() {
        p := Parcel{}
        err := rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
        if err != nil {
            return nil, err
        }
        parcels = append(parcels, p)
    }
    return parcels, nil
}

func (s ParcelStore) SetStatus(number int, status string) error {
    query := `UPDATE parcel SET status = ? WHERE number = ?`
    _, err := s.db.Exec(query, status, number)
    return err
}

func (s ParcelStore) SetAddress(number int, address string) error {
    p, err := s.Get(number)
    if err != nil {
        return err
    }
    if p.Status != ParcelStatusRegistered {
        return errors.New("cannot change address unless status is 'registered'")
    }
    query := `UPDATE parcel SET address = ? WHERE number = ?`
    _, err = s.db.Exec(query, address, number)
    return err
}

func (s ParcelStore) Delete(number int) error {
    p, err := s.Get(number)
    if err != nil {
        return err
    }
    if p.Status != ParcelStatusRegistered {
        return errors.New("cannot delete parcel unless status is 'registered'")
    }
    query := `DELETE FROM parcel WHERE number = ?`
    _, err = s.db.Exec(query, number)
    return err
}
