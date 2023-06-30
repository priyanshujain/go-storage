package encoding

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

type NestedStruct struct {
	Field1 string
	Field2 int64
}

type MyStruct struct {
	StringField string
	IntField    int
	FloatField  float64
	BoolField   bool
	ArrayField  [3]int
	SliceField  []string
	MapField    map[string]int
	StructField NestedStruct
}

// Test to check type change in schema
func TestDecode(t *testing.T) {
	data := MyStruct{
		StringField: "Hello",
		FloatField:  3.14,
		BoolField:   false,
		ArrayField:  [3]int{1, 2, 3},
		SliceField:  []string{"a", "b", "c"},
		MapField:    map[string]int{"a": 1, "b": 2, "c": 3},
		StructField: NestedStruct{Field1: "test", Field2: 2},
	}

	t.Run("decode fails", func(t *testing.T) {
		t.Run("when base64 encoded string breaks in db", func(t *testing.T) {
			var decodedData MyStruct
			err := Decode("invalid", &decodedData)
			if !errors.Is(err, ErrBase64Decoding) {
				t.Errorf("Expected error %v but got %v", ErrBase64Decoding, err)
			}
		})

		t.Run("when schema changes", func(t *testing.T) {
			t.Run("when field type changes", func(t *testing.T) {
				t.Run("unsupported types", func(t *testing.T) {
					type NewStruct struct {
						StringField string
						IntField    int
						FloatField  float64
						BoolField   interface{}
						ArrayField  [3]int
						SliceField  []string
						MapField    map[string]int
						StructField NestedStruct
					}
					encodedData, _ := Encode(data)
					var decodedData NewStruct
					err := Decode(encodedData, &decodedData)
					if !errors.Is(err, ErrUnsupportedType) {
						t.Errorf("Expected error %v but got %v", ErrUnsupportedType, err)
					}
				})
				t.Run("string changes to int", func(t *testing.T) {
					type NewStruct struct {
						StringField int
						IntField    int
						FloatField  float64
						BoolField   bool
						ArrayField  [3]int
						SliceField  []string
						MapField    map[string]int
						StructField NestedStruct
					}
					encodedData, _ := Encode(data)
					var decodedData NewStruct
					err := Decode(encodedData, &decodedData)
					if !errors.Is(err, ErrParseInt) {
						t.Errorf("Expected error %v but got %v", ErrParseInt, err)
					}
				})
				t.Run("string changes to float", func(t *testing.T) {
					type NewStruct struct {
						StringField float64
						IntField    int
						FloatField  float64
						BoolField   bool
						ArrayField  [3]int
						SliceField  []string
						MapField    map[string]int
						StructField NestedStruct
					}
					encodedData, _ := Encode(data)
					var decodedData NewStruct
					err := Decode(encodedData, &decodedData)
					if !errors.Is(err, ErrParseFloat) {
						t.Errorf("Expected error %v but got %v", ErrParseFloat, err)
					}
				})
				t.Run("string changes to bool", func(t *testing.T) {
					type NewStruct struct {
						StringField bool
						IntField    int
						FloatField  float64
						BoolField   bool
						ArrayField  [3]int
						SliceField  []string
						MapField    map[string]int
						StructField NestedStruct
					}
					encodedData, _ := Encode(data)
					var decodedData NewStruct
					err := Decode(encodedData, &decodedData)
					if !errors.Is(err, ErrParseBool) {
						t.Errorf("Expected error %v but got %v", ErrParseBool, err)
					}
				})
				t.Run("string changes to map", func(t *testing.T) {
					type NewStruct struct {
						StringField map[string]int
						IntField    int
						FloatField  float64
						BoolField   bool
						ArrayField  [3]int
						SliceField  []string
						MapField    map[string]int
						StructField NestedStruct
					}
					encodedData, _ := Encode(data)
					var decodedData NewStruct
					err := Decode(encodedData, &decodedData)
					if !errors.Is(err, ErrParseMap) {
						t.Errorf("Expected error %v but got %v", ErrParseMap, err)
					}
				})
				t.Run("map changes to struct", func(t *testing.T) {
					type NewStruct struct {
						StringField string
						IntField    int
						FloatField  float64
						BoolField   bool
						ArrayField  [3]int
						SliceField  []string
						MapField    struct {
							A int
							B float64
						}
						StructField NestedStruct
					}
					encodedData, _ := Encode(data)
					var decodedData NewStruct
					err := Decode(encodedData, &decodedData)
					if !errors.Is(err, ErrParseStruct) {
						t.Errorf("Expected error %v but got %v", ErrParseStruct, err)
					}
				})
				t.Run("string changes to array", func(t *testing.T) {
					type NewStruct struct {
						StringField [3]int
						IntField    int
						FloatField  float64
						BoolField   bool
						ArrayField  [3]int
						SliceField  []string
						MapField    map[string]int
						StructField NestedStruct
					}
					encodedData, _ := Encode(data)
					var decodedData NewStruct
					err := Decode(encodedData, &decodedData)
					if !errors.Is(err, ErrParseArray) {
						t.Errorf("Expected error %v but got %v", ErrParseArray, err)
					}
				})
				t.Run("string changes to slice", func(t *testing.T) {
					type NewStruct struct {
						StringField []string
						IntField    int
						FloatField  float64
						BoolField   bool
						ArrayField  [3]int
						SliceField  []string
						MapField    map[string]int
						StructField NestedStruct
					}
					encodedData, _ := Encode(data)
					var decodedData NewStruct
					err := Decode(encodedData, &decodedData)
					if !errors.Is(err, ErrParseSlice) {
						t.Errorf("Expected error %v but got %v", ErrParseSlice, err)
					}
				})
				t.Run("nested types", func(t *testing.T) {
					type NestedChan struct {
						ChanField chan struct {
							CharField string
						}
					}
					type NestChar struct {
						CharField string
					}
					type NewNestedStruct struct {
						StringField []string
						IntField    int
						FloatField  float64
						BoolField   bool
						ArrayField  [3]int
						SliceField  []NestChar
					}
					testData := NewNestedStruct{
						StringField: []string{"a", "b", "c"},
						IntField:    1,
						FloatField:  1.1,
						BoolField:   true,
						ArrayField:  [3]int{1, 2, 3},
						SliceField: []NestChar{
							{
								CharField: "a",
							},
						},
					}

					type DecodeStruct struct {
						StringField []string
						IntField    int
						FloatField  float64
						BoolField   bool
						ArrayField  [3]int
						SliceField  []NestedChan
					}

					encodedData, err := Encode(testData)
					if err != nil {
						t.Fatalf("Failed to encode data: %v", err)
					}
					var decodedData DecodeStruct
					err = Decode(encodedData, &decodedData)
					if !errors.Is(err, ErrParseSlice) {
						t.Errorf("Expected error %v but got %v", ErrParseSlice, err)
					}
				})
			})
		})
	})
	t.Run("decode success", func(t *testing.T) {
		encodedData, err := Encode(data)
		if err != nil {
			t.Fatalf("Failed to encode data: %v", err)
		}

		var decodedData MyStruct
		err = Decode(encodedData, &decodedData)
		if err != nil {
			t.Fatalf("Failed to decode data: %v", err)
		}

		if !reflect.DeepEqual(data, decodedData) {
			t.Errorf("Decoded data does not match original data.\nExpected: %+v\nGot: %+v", data, decodedData)
		}
	})
}

func TestEncode(t *testing.T) {
	t.Run("encode fails", func(t *testing.T) {
		t.Run("unsupported type", func(t *testing.T) {
			type UnsupportedType chan int
			data := UnsupportedType(make(chan int))
			_, err := Encode(data)
			if !errors.Is(err, ErrUnsupportedType) {
				t.Errorf("Expected error %v but got %v", ErrUnsupportedType, err)
			}
		})
		t.Run("unsupported type parameter", func(t *testing.T) {
			type NewStruct struct {
				StringField []string
				IntField    int
				FloatField  func() float64
				BoolField   chan bool
				ArrayField  [3]int
				SliceField  []string
				MapField    map[string]int
				StructField NestedStruct
			}
			data := NewStruct{
				StringField: []string{"a", "b", "c"},
				IntField:    1,
			}
			_, err := Encode(data)
			if !errors.Is(err, ErrUnsupportedType) {
				t.Errorf("Expected error %v but got %v", ErrUnsupportedType, err)
			}
		})
		t.Run("unsupported type parameter nested", func(t *testing.T) {
			type NewStruct struct {
				StringField string
				IntField    int
				FloatField  float64
				BoolField   bool
				ArrayField  [3]int
				SliceField  []string
				MapField    map[string]int
				StructField struct {
					A int
					B chan float64
				}
			}
			data := NewStruct{
				StringField: "Hello",
				IntField:    1,
				StructField: struct {
					A int
					B chan float64
				}{
					A: 1,
					B: make(chan float64),
				},
			}
			_, err := Encode(data)
			if !errors.Is(err, ErrUnsupportedType) {
				t.Errorf("Expected error %v but got %v", ErrUnsupportedType, err)
			}
		})
		t.Run("unsupported type parameter inside array", func(t *testing.T) {
			type NewStruct struct {
				StringField string
				IntField    int
				FloatField  float64
				BoolField   bool
				ArrayField  [1]chan int
				SliceField  []string
				MapField    map[string]int
				StructField NestedStruct
			}
			data := NewStruct{
				StringField: "Hello",
				IntField:    1,
				ArrayField: [1]chan int{
					make(chan int),
				},
			}
			_, err := Encode(data)
			if !errors.Is(err, ErrUnsupportedType) {
				t.Errorf("Expected error %v but got %v", ErrUnsupportedType, err)
			}
		})
		t.Run("unsupported type parameter inside slice", func(t *testing.T) {
			type NewStruct struct {
				StringField string
				IntField    int
				FloatField  float64
				BoolField   bool
				ArrayField  [3]int
				SliceField  []chan string
				MapField    map[string]int
				StructField NestedStruct
			}
			data := NewStruct{
				StringField: "Hello",
				IntField:    1,
				SliceField: []chan string{
					make(chan string),
				},
			}
			_, err := Encode(data)
			if !errors.Is(err, ErrUnsupportedType) {
				t.Errorf("Expected error %v but got %v", ErrUnsupportedType, err)
			}
		})
		t.Run("unsupported type parameter inside map", func(t *testing.T) {
			type NewStruct struct {
				StringField string
				IntField    int
				FloatField  float64
				BoolField   bool
				ArrayField  [3]int
				SliceField  []string
				MapField    map[string]chan int
				StructField NestedStruct
			}
			data := NewStruct{
				StringField: "Hello",
				IntField:    1,
				MapField: map[string]chan int{
					"a": make(chan int),
				},
			}
			_, err := Encode(data)
			if !errors.Is(err, ErrUnsupportedType) {
				t.Errorf("Expected error %v but got %v", ErrUnsupportedType, err)
			}
		})
		t.Run("unsupported type parameter inside struct slice", func(t *testing.T) {
			type NewStruct struct {
				StringField string
				IntField    int
				FloatField  float64
				BoolField   bool
				ArrayField  [3]int
				SliceField  []struct {
					A int
					B chan float64
				}
				MapField    map[string]chan int
				StructField NestedStruct
			}
			data := NewStruct{
				StringField: "Hello",
				IntField:    1,
				SliceField: []struct {
					A int
					B chan float64
				}{
					{
						A: 1,
						B: make(chan float64),
					},
				},
			}
			_, err := Encode(data)
			if !errors.Is(err, ErrUnsupportedType) {
				t.Errorf("Expected error %v but got %v", ErrUnsupportedType, err)
			}
		})
	})
}

func TestEncodingDecoding(t *testing.T) {
	t.Run("successful encode and decode", func(t *testing.T) {
		t.Run("simple case", func(t *testing.T) {
			data := MyStruct{
				StringField: "Hello",
				FloatField:  3.14,
				BoolField:   false,
				ArrayField:  [3]int{1, 2, 3},
				SliceField:  []string{"a", "b", "c"},
				MapField:    map[string]int{"a": 1, "b": 2, "c": 3},
				StructField: NestedStruct{Field1: "test", Field2: 2},
			}
			encodedData, err := Encode(data)
			if err != nil {
				t.Fatalf("Failed to encode data: %v", err)
			}

			var decodedData MyStruct
			err = Decode(encodedData, &decodedData)
			if err != nil {
				t.Fatalf("Failed to decode data: %v", err)
			}

			if !reflect.DeepEqual(data, decodedData) {
				t.Errorf("Decoded data does not match original data.\nExpected: %+v\nGot: %+v", data, decodedData)
			}
		})
		t.Run("simple case with null values", func(t *testing.T) {
			data := MyStruct{
				StringField: "Hello",
				FloatField:  3.14,
				BoolField:   false,
				ArrayField:  [3]int{1, 2, 3},
				SliceField:  nil,
				MapField:    nil,
				StructField: NestedStruct{Field1: "test", Field2: 2},
			}
			encodedData, err := Encode(data)
			if err != nil {
				t.Fatalf("Failed to encode data: %v", err)
			}

			var decodedData MyStruct
			err = Decode(encodedData, &decodedData)
			if err != nil {
				t.Fatalf("Failed to decode data: %v", err)
			}

			if !reflect.DeepEqual(data, decodedData) {
				a, _ := json.Marshal(data)
				b, _ := json.Marshal(decodedData)
				t.Errorf("Decoded data does not match original data.\nExpected: %s\nGot: %s", a, b)
				t.Errorf("Decoded data does not match original data.\nExpected: %+v\nGot: %+v", data, decodedData)
			}
		})
		t.Run("nested case", func(t *testing.T) {
			type Address struct {
				Street     string
				City       string
				State      string
				PostalCode string
			}

			type Contact struct {
				Phone   string
				Email   string
				Address Address
			}

			type Person struct {
				Name    string
				Age     int
				Contact Contact
			}

			type Company struct {
				Name            string
				Location        Address
				CEO             Person
				Staff           [][2]Person
				Departments     []string
				SalaryMap       map[string]float64
				EmployeeHistory map[string]map[string]int
				Availability    []bool
				Projects        [][]string
			}

			address := Address{
				Street:     "123 Main Street",
				City:       "New York",
				State:      "NY",
				PostalCode: "10001",
			}

			ceoContact := Contact{
				Phone:   "555-1234",
				Email:   "ceo@example.com",
				Address: address,
			}

			ceo := Person{
				Name:    "John Doe",
				Age:     40,
				Contact: ceoContact,
			}

			employee1Contact := Contact{
				Phone:   "555-5678",
				Email:   "employee1@example.com",
				Address: address,
			}

			employee1 := Person{
				Name:    "Jane Smith",
				Age:     30,
				Contact: employee1Contact,
			}

			employee2Contact := Contact{
				Phone:   "555-9090",
				Email:   "employee2@example.com",
				Address: address,
			}

			employee2 := Person{
				Name:    "David Johnson",
				Age:     35,
				Contact: employee2Contact,
			}

			companyAddress := Address{
				Street:     "456 Market Street",
				City:       "San Francisco",
				State:      "CA",
				PostalCode: "94101",
			}

			staff := [][2]Person{
				{employee2, employee1},
			}

			availability := []bool{true, false, true}

			projects := [][]string{
				{"Project 1", "Project 2"},
				{"Project 3", "Project 4", "Project 5"},
			}

			company := Company{
				Name:        "Acme Inc.",
				Location:    companyAddress,
				CEO:         ceo,
				Staff:       staff,
				Departments: []string{"Sales", "Marketing", "Engineering"},
				SalaryMap:   map[string]float64{"John Doe": 100000.0, "Jane Smith": 80000.0, "David Johnson": 90000.0},
				EmployeeHistory: map[string]map[string]int{"John Doe": {"Sales": 2, "Marketing": 4, "Engineering": 3},
					"Jane Smith":    {"Sales": 1, "Marketing": 3, "Engineering": 2},
					"David Johnson": {"Sales": 3, "Marketing": 5, "Engineering": 4}},
				Availability: availability,
				Projects:     projects,
			}
			encodedData, err := Encode(company)
			if err != nil {
				t.Fatalf("Failed to encode data: %v", err)
			}
			var decodedData Company
			err = Decode(encodedData, &decodedData)
			if err != nil {
				t.Fatalf("Failed to decode data: %v", err)
			}
		})
	})
}
