package inmemory

import (
	"errors"
	"fmt"
	"sync"
	"testing"
)

func TestInMemoryStorage(t *testing.T) {
	storage := &InMemoryStorage{data: make(map[string]string)}

	// Insert
	err := storage.Insert("key1", "value1")
	if err != nil {
		t.Errorf("Failed to insert record: %s", err)
	}

	// Get
	value, err := storage.Get("key1")
	if err != nil {
		t.Errorf("Failed to get record: %s", err)
	}
	expectedValue := "value1"
	if value != expectedValue {
		t.Errorf("Expected value %q, got %q", expectedValue, value)
	}

	// Update
	err = storage.Update("key1", "new value")
	if err != nil {
		t.Errorf("Failed to update record: %s", err)
	}

	// Get after update
	value, err = storage.Get("key1")
	if err != nil {
		t.Errorf("Failed to get record after update: %s", err)
	}
	expectedValue = "new value"
	if value != expectedValue {
		t.Errorf("Expected value %q, got %q", expectedValue, value)
	}

	// Delete
	err = storage.Delete("key1")
	if err != nil {
		t.Errorf("Failed to delete record: %s", err)
	}

	// Get after delete
	_, err = storage.Get("key1")
	if err == nil {
		t.Errorf("Expected error for non-existent key, got nil")
	}

	// delete non-existent key
	err = storage.Delete("key1")
	if err != errKeyNotFound {
		t.Errorf("Expected error %q, got %q", errKeyNotFound, err)
	}

	// update non-existent key
	err = storage.Update("key1", "value1")
	if err != errKeyNotFound {
		t.Errorf("Expected error %q, got %q", errKeyNotFound, err)
	}

	// insert existing key
	_ = storage.Insert("key1", "value1")
	err = storage.Insert("key1", "value1")
	if err != errKeyAlreadyExists {
		t.Errorf("Expected error %q, got %q", errKeyAlreadyExists, err)
	}
}

func TestInMemoryStorage_ConcurrentAccess(t *testing.T) {
	storage := &InMemoryStorage{data: make(map[string]string)}

	// Concurrent inserts
	concurrentInserts := 100
	done := make(chan struct{})
	for i := 0; i < concurrentInserts; i++ {
		go func(key string, value string) {
			err := storage.Insert(key, value)
			if err != nil {
				t.Errorf("Failed to insert record: %s", err)
			}
			done <- struct{}{}
		}(fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i))
	}

	// Wait for all inserts to complete
	for i := 0; i < concurrentInserts; i++ {
		<-done
	}

	// Verify all records are present
	for i := 0; i < concurrentInserts; i++ {
		key := fmt.Sprintf("key%d", i)
		value := fmt.Sprintf("value%d", i)
		retrievedValue, err := storage.Get(key)
		if err != nil {
			t.Errorf("Failed to get record for key %q: %s", key, err)
		}
		if retrievedValue != value {
			t.Errorf("Mismatched value for key %q: expected %q, got %q", key, value, retrievedValue)
		}
	}
}

func TestInMemoryStorage_ConcurrentUpdate(t *testing.T) {
	storage := &InMemoryStorage{data: make(map[string]string)}

	// Insert initial record
	err := storage.Insert("key1", "value1")
	if err != nil {
		t.Errorf("Failed to insert initial record: %s", err)
	}

	// Concurrent updates
	concurrentUpdates := 10
	numIterations := 100

	// Wait for all goroutines to finish
	var wg sync.WaitGroup
	wg.Add(concurrentUpdates)

	// Concurrent updates
	for i := 0; i < concurrentUpdates; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				err := storage.Update("key1", "new value")
				if err != nil {
					t.Errorf("Failed to update record: %s", err)
				}
			}
		}()
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Verify updated value
	value, err := storage.Get("key1")
	if err != nil {
		t.Errorf("Failed to get record: %s", err)
	}
	expectedValue := "new value"
	if value != expectedValue {
		t.Errorf("Mismatched value: expected %q, got %q", expectedValue, value)
	}
}

func TestInMemoryStorage_ConcurrentUpdateDelete(t *testing.T) {
	storage := &InMemoryStorage{data: make(map[string]string)}

	// Concurrent delete and insert operations
	concurrentOps := 10

	// initial insert
	for i := 0; i < concurrentOps; i++ {
		key := fmt.Sprintf("key%d", i)
		err := storage.Insert(key, "value1")
		if err != nil {
			t.Errorf("Failed to insert record: %s", err)
		}
	}

	// Concurrent updates and deletes
	done := make(chan struct{})
	for i := 0; i < concurrentOps; i++ {
		go func(key string, value string) {
			err := storage.Update(key, value)
			if err != nil {
				t.Errorf("Failed to update record: %s", err)
			}
			done <- struct{}{}
			err = storage.Delete(key)
			if err != nil {
				t.Errorf("Failed to delete record: %s", err)
			}
			done <- struct{}{}
		}(fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i))
	}

	// Wait for all inserts and updates to complete
	for i := 0; i < concurrentOps*2; i++ {
		<-done
	}

	// Verify record states
	_, err := storage.Get("key1")
	if err == nil {
		t.Errorf("Expected error for non-existent key, got nil")
	}
}

func TestInMemoryStorage_ConcurrentInsertUpdate(t *testing.T) {
	storage := &InMemoryStorage{data: make(map[string]string)}

	// Concurrent inserts and updates
	concurrentInserts := 50
	done := make(chan struct{})

	for i := 0; i < concurrentInserts; i++ {
		go func(key string, value string) {
			err := storage.Insert(key, value)
			if err != nil {
				t.Errorf("Failed to insert record: %s", err)
			}
			done <- struct{}{}
			err = storage.Update(key, "new value")
			if err != nil {
				t.Errorf("Failed to update record: %s", err)
			}
			done <- struct{}{}
		}(fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i))
	}

	// Wait for all inserts and updates to complete
	for i := 0; i < concurrentInserts*2; i++ {
		<-done
	}

	// Verify all records are updated
	for i := 0; i < concurrentInserts; i++ {
		key := fmt.Sprintf("key%d", i)
		retrievedValue, err := storage.Get(key)
		if err != nil {
			t.Errorf("Failed to get record for key %q: %s", key, err)
		}
		expectedValue := "new value"
		if retrievedValue != expectedValue {
			t.Errorf("Mismatched value for key %q: expected %q, got %q", key, expectedValue, retrievedValue)
		}
	}
}

type ExampleStruct struct {
	ID   string
	Name string
}

func TestDatabase_CreateTable(t *testing.T) {
	db := New()

	// Test creating a new table
	err := db.CreateTable(ExampleStruct{}, "ID")
	if err != nil {
		t.Errorf("Failed to create table: %v", err)
	}

	// Test creating duplicate table
	err = db.CreateTable(ExampleStruct{}, "ID")
	if err != ErrTableExists {
		t.Errorf("Expected ErrTableExists, but got: %v", err)
	}

	db = New()

	// Test creating a table with an invalid primary key
	err = db.CreateTable(ExampleStruct{}, "InvalidKey")
	if err != ErrInvalidPk {
		t.Errorf("Expected ErrInvalidPk, but got: %v", err)
	}
}

func TestDatabase_Insert(t *testing.T) {
	db := New()

	// Create a table and insert a record
	err := db.CreateTable(ExampleStruct{}, "ID")
	if err != nil {
		t.Errorf("Failed to create table: %v", err)
	}

	record := ExampleStruct{ID: "1", Name: "John Doe"}
	err = db.Insert(record)
	if err != nil {
		t.Errorf("Failed to insert record: %v", err)
	}

	// Try inserting a record into a non-existing table
	err = db.Insert("invalid record")
	if err != ErrInvalidTableName {
		t.Errorf("Expected ErrInvalidTableName, but got: %v", err)
	}

	// Try inserting a duplicate record
	err = db.Insert(record)
	if err == nil {
		t.Error("Expected an error when inserting duplicate record, but got nil")
	}

	t.Run("encoding failure", func(t *testing.T) {
		db = &Database{}
		db.Init()

		type ExampleStruct struct {
			ID   string
			Name chan int
		}
		// Create a table and insert a record
		_ = db.CreateTable(ExampleStruct{}, "ID")

		record := ExampleStruct{ID: "1", Name: make(chan int)}
		err = db.Insert(record)
		if !errors.Is(err, ErrInvalidEncoding) {
			t.Errorf("Expected ErrInvalidEncoding, but got: %v", err)
		}

	})
	t.Run("pointer record", func(t *testing.T) {
		db = &Database{}
		db.Init()

		type ExampleStruct struct {
			ID   string
			Name string
		}
		// Create a table and insert a record
		_ = db.CreateTable(ExampleStruct{}, "ID")

		record := &ExampleStruct{ID: "1", Name: "John Doe"}
		err = db.Insert(record)
		if err != nil {
			t.Errorf("Failed to insert record: %v", err)
		}
	})
}

func TestDatabase_Get(t *testing.T) {
	db := New()

	// Create a table and insert a record
	err := db.CreateTable(ExampleStruct{}, "ID")
	if err != nil {
		t.Errorf("Failed to create table: %v", err)
	}

	record := ExampleStruct{ID: "1", Name: "John Doe"}
	err = db.Insert(record)
	if err != nil {
		t.Errorf("Failed to insert record: %v", err)
	}

	// Get an existing record
	result, err := db.Get(ExampleStruct{}, "1")
	if err != nil {
		t.Errorf("Failed to get record: %v", err)
	}

	// Check the retrieved record
	retrievedRecord, ok := result.(*ExampleStruct)
	if !ok {
		t.Errorf("Expected result to be of type *ExampleStruct, but got: %T", result)
	}
	if retrievedRecord.ID != "1" || retrievedRecord.Name != "John Doe" {
		t.Errorf("Unexpected record values. Expected: %+v, Got: %+v", record, retrievedRecord)
	}

	// Try getting a record from a non-existing table
	_, err = db.Get("InvalidTable", "1")
	if err != ErrInvalidTableName {
		t.Errorf("Expected ErrInvalidTableName, but got: %v", err)
	}

	// Try getting a non-existing record
	_, err = db.Get(ExampleStruct{}, "2")
	if err != ErrRecordNotFound {
		t.Errorf("Expected ErrRecordNotFound, but got: %v", err)
	}
}
