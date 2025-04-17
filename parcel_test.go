
package main

import (
    "database/sql"
    "math/rand"
    "testing"
    "time"

    "github.com/stretchr/testify/require"
    _ "modernc.org/sqlite"
)

var (
    randSource = rand.NewSource(time.Now().UnixNano())
    randRange  = rand.New(randSource)
)

func getTestParcel() Parcel {
    return Parcel{
        Client:    1000,
        Status:    ParcelStatusRegistered,
        Address:   "test",
        CreatedAt: time.Now().UTC().Format(time.RFC3339),
    }
}

func openTestDB(t *testing.T) *sql.DB {
    db, err := sql.Open("sqlite", "tracker.db")
    require.NoError(t, err)
    return db
}

func TestAddGetDelete(t *testing.T) {
    db := openTestDB(t)
    store := NewParcelStore(db)
    parcel := getTestParcel()

    id, err := store.Add(parcel)
    require.NoError(t, err)
    require.NotZero(t, id)

    fetched, err := store.Get(id)
    require.NoError(t, err)
    require.Equal(t, parcel.Client, fetched.Client)
    require.Equal(t, parcel.Status, fetched.Status)
    require.Equal(t, parcel.Address, fetched.Address)

    err = store.Delete(id)
    require.NoError(t, err)

    _, err = store.Get(id)
    require.Error(t, err)
}

func TestSetAddress(t *testing.T) {
    db := openTestDB(t)
    store := NewParcelStore(db)
    parcel := getTestParcel()

    id, err := store.Add(parcel)
    require.NoError(t, err)

    newAddress := "new test address"
    err = store.SetAddress(id, newAddress)
    require.NoError(t, err)

    updated, err := store.Get(id)
    require.NoError(t, err)
    require.Equal(t, newAddress, updated.Address)
}

func TestSetStatus(t *testing.T) {
    db := openTestDB(t)
    store := NewParcelStore(db)
    parcel := getTestParcel()

    id, err := store.Add(parcel)
    require.NoError(t, err)

    err = store.SetStatus(id, ParcelStatusSent)
    require.NoError(t, err)

    updated, err := store.Get(id)
    require.NoError(t, err)
    require.Equal(t, ParcelStatusSent, updated.Status)
}

func TestGetByClient(t *testing.T) {
    db := openTestDB(t)
    store := NewParcelStore(db)

    client := randRange.Intn(10_000_000)
    parcels := []Parcel{
        getTestParcel(),
        getTestParcel(),
        getTestParcel(),
    }

    for i := range parcels {
        parcels[i].Client = client
        id, err := store.Add(parcels[i])
        require.NoError(t, err)
        parcels[i].Number = id
    }

    stored, err := store.GetByClient(client)
    require.NoError(t, err)
    require.Len(t, stored, len(parcels))

    m := map[int]Parcel{}
    for _, p := range parcels {
        m[p.Number] = p
    }

    for _, p := range stored {
        original, ok := m[p.Number]
        require.True(t, ok)
        require.Equal(t, original.Client, p.Client)
        require.Equal(t, original.Status, p.Status)
        require.Equal(t, original.Address, p.Address)
    }
}
