package storage

import (
	"testing"
)

func TestStorageEngine_InMemory(t *testing.T) {
	// Create an instance of the storage engine
	engine := StorageEngine[EngineType("inmemory")]
	engine.Init()

	type Person struct {
		Name string
		Id   string
	}

	// Test CreateTable method
	err := engine.CreateTable(Person{}, "Id")
	if err != nil {
		t.Errorf("Failed to create table: %v", err)
	}

	// Test Insert method
	record := Person{
		Name: "John Doe",
		Id:   "123",
	}
	err = engine.Insert(record)
	if err != nil {
		t.Errorf("Failed to insert record: %v", err)
	}

	// Test Get method
	_, err = engine.Get(Person{}, "123")
	if err != nil {
		t.Errorf("Failed to get record: %v", err)
	}
}
