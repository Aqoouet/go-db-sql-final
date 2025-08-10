package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/assert"

)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
    require.NoErrorf(t, err, "Failed to initialize connection with database")
    tx,err := db.Begin()
    defer tx.Rollback()
    require.NoErrorf(t, err, "Failed to begin transaction")
    store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора

    number, err := store.Add(parcel)
    require.NoErrorf(t, err, "Failed to add parcel")
    require.NotEmpty(t, number)

	// get
	// получите только что добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel

    returnedParcel, err := store.Get(number)
    require.NoErrorf(t, err, "Failed to get parcel with number = %d", number)
    assert.Equalf(t, parcel.Client, returnedParcel.Client, "Row number %d: expected client = %d not equal to returned client = %d",  number, parcel.Client,  returnedParcel.Client)
    assert.Equalf(t, parcel.Status, returnedParcel.Status, "Row number %d: expected status = %s not equal to returned status = %s",  number, parcel.Status,  returnedParcel.Status)
    assert.Equalf(t, parcel.Address, returnedParcel.Address, "Row number %d: expected address = %s not equal to returned address = %s",  number, parcel.Address,  returnedParcel.Address)
    assert.Equalf(t, parcel.CreatedAt, returnedParcel.CreatedAt, "Row number %d: expected created_at time  = %v not equal to returned created_at time = %v",  number, parcel.CreatedAt,  returnedParcel.CreatedAt)


	// delete
	// удалите добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что посылку больше нельзя получить из БД

    err=store.Delete(number)
    require.NoErrorf(t, err, "Failed to delete parcel with number = %d", number)

    _, err = store.Get(number)
    require.Errorf(t, err, "Get request returned no error, while there must be no rows satisfing the request")
    require.ErrorIsf(t, sql.ErrNoRows, err , "Returned unexpected error %w", err)


}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare

	db, err := sql.Open("sqlite", "tracker.db")
    require.NoErrorf(t, err, "Failed to initialize connection with database")
    tx,err := db.Begin()
    defer tx.Rollback()
    require.NoErrorf(t, err, "Failed to begin transaction")
    store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора

    number, err := store.Add(parcel)
    require.NoErrorf(t, err, "Failed to add parcel")
    require.NotEmpty(t, number)

	// set address
	// обновите адрес, убедитесь в отсутствии ошибки
	newAddress := "new test address"

    err = store.SetAddress(number, newAddress)
    require.NoErrorf(t, err, "Failed to set address")

	// check
	// получите добавленную посылку и убедитесь, что адрес обновился

    returnedParcel, err = store.Get(number)
    require.NoErrorf(t, err, "Failed to get parcel with number = %d", number)
    assert.Equalf(t, newAddress, returnedParcel.Address, "Row number %d: expected address = %s not equal to returned address = %s",  number, newAddress,  returnedParcel.Address)


}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare

	db, err := sql.Open("sqlite", "tracker.db")
    require.NoErrorf(t, err, "Failed to initialize connection with database")
    tx,err := db.Begin()
    defer tx.Rollback()
    require.NoErrorf(t, err, "Failed to begin transaction")
    store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора

    number, err := store.Add(parcel)
    require.NoErrorf(t, err, "Failed to add parcel")
    require.NotEmpty(t, number)


	// set status
	// обновите статус, убедитесь в отсутствии ошибки
    err =  store.SetStatus(number, ParcelStatusDelivered)


	// check
	// получите добавленную посылку и убедитесь, что статус обновился
    returnedParcel, err = store.Get(number)
    require.NoErrorf(t, err, "Failed to get parcel with number = %d", number)
    assert.Equalf(t,  ParcelStatusDelivered, returnedParcel.Status, "Row number %d: expected status = %s not equal to returned status = %s",  number, ParcelStatusDelivered   ,  returnedParcel.Status)


}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare

	db, err := sql.Open("sqlite", "tracker.db")
    require.NoErrorf(t, err, "Failed to initialize connection with database")
    tx,err := db.Begin()
    defer tx.Rollback()
    require.NoErrorf(t, err, "Failed to begin transaction")
    store := NewParcelStore(db)
	parcel := getTestParcel()
    
	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}

	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := // добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора

		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := // получите список посылок по идентификатору клиента, сохранённого в переменной client
	// убедитесь в отсутствии ошибки
	// убедитесь, что количество полученных посылок совпадает с количеством добавленных

	// check
	for _, parcel := range storedParcels {
		// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
		// убедитесь, что все посылки из storedParcels есть в parcelMap
		// убедитесь, что значения полей полученных посылок заполнены верно
	}
}
