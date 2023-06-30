package inmemory

import (
	"errors"
	"fmt"
	"github.com/priyanshujain/go-storage/encoding"
	"reflect"
	"sync"
)

type Record struct {
	Key   string
	Value string
}

type InMemoryStorage struct {
	data  map[string]string
	mutex sync.RWMutex
}

// local errors
var errKeyAlreadyExists = errors.New("key already exists")
var errKeyNotFound = errors.New("key not found")

func (s *InMemoryStorage) Insert(key, value string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, ok := s.data[key]; ok {
		return errKeyAlreadyExists
	}

	s.data[key] = value
	return nil
}

func (s *InMemoryStorage) Get(key string) (string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	value, ok := s.data[key]
	if !ok {
		return "", errKeyNotFound
	}

	return value, nil
}

func (s *InMemoryStorage) Update(key, value string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, ok := s.data[key]; !ok {
		return errKeyNotFound
	}

	s.data[key] = value
	return nil
}

func (s *InMemoryStorage) Delete(key string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, ok := s.data[key]; !ok {
		return errKeyNotFound
	}

	delete(s.data, key)
	return nil
}

type Table struct {
	Name    string
	Pk      string
	Fields  reflect.Type
	Records []*Record
}

type Database struct {
	Tables  map[string]*Table
	Storage *InMemoryStorage
}

func (db *Database) Init() {
	db.Tables = make(map[string]*Table)
	db.Storage = &InMemoryStorage{data: make(map[string]string)}
}

func New() *Database {
	return &Database{
		Tables:  make(map[string]*Table),
		Storage: &InMemoryStorage{data: make(map[string]string)},
	}
}

var ErrInvalidPk = errors.New("invalid primary key")
var ErrInvalidTableName = errors.New("invalid table name")
var ErrInvalidEncoding = errors.New("invalid encoding")
var ErrRecordNotFound = errors.New("record not found")
var ErrTableExists = errors.New("table already exists")
var ErrDuplicateRecord = errors.New("duplicate record")

// create a new table in the database
func (db *Database) CreateTable(tType interface{}, pk string) error {
	// get the name of the struct using reflection
	tableType := reflect.TypeOf(tType)
	name := reflect.TypeOf(tType).Name()

	if _, ok := db.Tables[name]; ok {
		return ErrTableExists
	}

	_, found := tableType.FieldByName(pk)

	if !found {
		return ErrInvalidPk
	}
	db.Tables[name] = &Table{Name: name, Fields: tableType, Pk: pk}
	return nil
}

// insert a record into the table
func (db *Database) Insert(record interface{}) error {
	if reflect.TypeOf(record).Kind() == reflect.Ptr {
		value := reflect.ValueOf(record).Elem()
		newValue := reflect.New(value.Type()).Elem()
		newValue.Set(value)
		record = newValue.Interface()
	}

	tableName := reflect.TypeOf(record).Name()
	table, ok := db.Tables[tableName]

	if !ok {
		return ErrInvalidTableName
	}

	// get the value of the primary key
	pk := reflect.ValueOf(record).FieldByName(table.Pk).String()

	// check if the record already exists
	for _, r := range table.Records {
		if r.Key == pk {
			return ErrDuplicateRecord
		}
	}

	value, err := encoding.Encode(record)
	if err != nil {
		return fmt.Errorf("error encoding record: %v %w", err, ErrInvalidEncoding)
	}

	// insert the record
	table.Records = append(table.Records, &Record{Key: pk, Value: value})
	return nil
}

// get a record from the table
func (db *Database) Get(tableType interface{}, pk string) (interface{}, error) {
	tableName := reflect.TypeOf(tableType).Name()
	table, ok := db.Tables[tableName]

	if !ok {
		return nil, ErrInvalidTableName
	}

	// get the record
	for _, r := range table.Records {
		if r.Key == pk {
			record := reflect.New(table.Fields).Interface()
			// decoding failure can not happen until we change the table fields and we are not doing it as of now
			_ = encoding.Decode(r.Value, record)
			return record, nil
		}
	}

	return nil, ErrRecordNotFound
}
