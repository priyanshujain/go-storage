package storage

import (
	"github.com/priyanshujain/go-storage/drivers/inmemory"
)

type Storage interface {
	Init()
	CreateTable(tableType interface{}, pk string) error
	Insert(record interface{}) error
	Get(tableType interface{}, pk string) (interface{}, error)
}

type EngineType string

var StorageEngine map[EngineType]Storage = map[EngineType]Storage{
	"inmemory": &inmemory.Database{},
}
