package encoding

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"errors"
)

var (
	// type parsing errors
	ErrParseBool   = errors.New("cannot decode bool type")
	ErrParseInt    = errors.New("cannot decode int type")
	ErrParseFloat  = errors.New("cannot decode float type")
	ErrParseSlice  = errors.New("cannot decode slice type")
	ErrParseMap    = errors.New("cannot decode map type")
	ErrParseStruct = errors.New("cannot decode struct type")
	ErrParseArray  = errors.New("cannot decode array type")
	ErrParsePtr    = errors.New("cannot decode pointer type")

	// encoding errors
	ErrBase64Decoding = errors.New("cannot base64 decode")

	// type support errors
	ErrUnsupportedType = errors.New("unsupported type")

	// field validation errors
	ErrInvalidFieldValues = errors.New("invalid field values")
)

func Decode(record string, data interface{}) error {
	// base64 decode string
	decodedRecord, err := base64.StdEncoding.DecodeString(record)
	if err != nil {
		return fmt.Errorf("%v: %w", err, ErrBase64Decoding)
	}
	record = string(decodedRecord)
	fieldValues := strings.Split(record, ",")

	v := reflect.ValueOf(data).Elem()
	for i := 0; i < v.NumField(); i++ {
		fieldValue := fieldValues[i]
		fieldType := v.Field(i).Type()

		switch fieldType.Kind() {
		case reflect.String:
			v.Field(i).SetString(fieldValue)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			intValue, err := strconv.ParseInt(fieldValue, 10, 64)
			if err != nil {
				return fmt.Errorf("%v: %w", err, ErrParseInt)
			}
			v.Field(i).SetInt(intValue)
		case reflect.Float32, reflect.Float64:
			floatValue, err := strconv.ParseFloat(fieldValue, 64)
			if err != nil {
				return fmt.Errorf("%v: %w", err, ErrParseFloat)
			}
			v.Field(i).SetFloat(floatValue)
		case reflect.Bool:
			boolValue, err := strconv.ParseBool(fieldValue)
			if err != nil {
				return fmt.Errorf("%v: %w", err, ErrParseBool)
			}
			v.Field(i).SetBool(boolValue)
		case reflect.Struct:
			st, err := decodeStruct(fieldType, fieldValue)
			if err != nil {
				return fmt.Errorf("%v: %w", err, ErrParseStruct)
			}
			v.Field(i).Set(st)
		case reflect.Array:
			array, err := decodeArray(fieldType, fieldValue)
			if err != nil {
				return fmt.Errorf("%v: %w", err, ErrParseArray)
			}
			v.Field(i).Set(array)
		case reflect.Slice:
			slice, err := decodeSlice(fieldType, fieldValue)
			if err != nil {
				return fmt.Errorf("%v: %w", err, ErrParseSlice)
			}
			v.Field(i).Set(slice)
		case reflect.Map:
			mapValue, err := decodeMap(fieldType, fieldValue)
			if err != nil {
				return fmt.Errorf("%v: %w", err, ErrParseMap)
			}
			v.Field(i).Set(mapValue)
		case reflect.Ptr:
			ptrValue, err := decodePtr(fieldType, fieldValue)
			if err != nil {
				return fmt.Errorf("%v: %w", err, ErrParsePtr)
			}
			v.Field(i).Set(ptrValue)
		default:
			return ErrUnsupportedType
		}
	}
	return nil
}

// decodePtr decodes a pointer type
func decodePtr(ptrType reflect.Type, value string) (reflect.Value, error) {
	if value == "" {
		return reflect.Zero(ptrType), nil
	}
	ptrValue := reflect.New(ptrType.Elem())
	valueType := ptrType.Elem()
	decodedValue, err := decodeValue(valueType, value)
	if err != nil {
		return reflect.Zero(ptrType), err
	}
	ptrValue.Elem().Set(decodedValue)
	return ptrValue, nil
}

func decodeArray(arrayType reflect.Type, value string) (reflect.Value, error) {
	decodedValue, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return reflect.Zero(arrayType), fmt.Errorf("%v: %w", err, ErrBase64Decoding)
	}
	fieldValues := strings.Split(string(decodedValue), ",")
	arrayLen := arrayType.Len()
	if len(fieldValues) < arrayLen {
		return reflect.Zero(arrayType), ErrInvalidFieldValues
	}
	array := reflect.New(arrayType).Elem()
	for i := 0; i < arrayLen; i++ {
		valueType := arrayType.Elem()
		value, err := decodeValue(valueType, fieldValues[i])
		if err != nil {
			return reflect.Zero(arrayType), err
		}
		array.Index(i).Set(value)
	}
	return array, nil
}

func decodeSlice(sliceType reflect.Type, value string) (reflect.Value, error) {
	decodedValue, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return reflect.Zero(sliceType), fmt.Errorf("%v: %w", err, ErrBase64Decoding)
	}
	fieldValues := strings.Split(string(decodedValue), ",")
	sliceElemType := sliceType.Elem()
	slice := reflect.MakeSlice(sliceType, 0, len(fieldValues))

	for i, fieldValue := range fieldValues {
		value, err := decodeValue(sliceElemType, fieldValue)
		if err != nil {
			return reflect.Zero(sliceType), err
		}
		if i < slice.Len() {
			slice.Index(i).Set(value)
		} else {
			slice = reflect.Append(slice, value)
		}
	}
	// check if slice is empty
	if slice.Len() == 0 || slice.Index(0).IsZero() {
		return reflect.Zero(sliceType), nil
	}
	return slice, nil
}

// decodeMap decodes a map from a base64 encoded string
func decodeMap(mapType reflect.Type, value string) (reflect.Value, error) {
	decodedValue, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return reflect.Zero(mapType), fmt.Errorf("%v: %w", err, ErrBase64Decoding)
	}
	fieldValues := strings.Split(string(decodedValue), ",")
	m := reflect.MakeMap(mapType)
	// get key and value types of the map
	keyType := mapType.Key()
	valueType := mapType.Elem()

	for i := 0; i < len(fieldValues); i += 1 {
		if fieldValues[i] == "" {
			continue
		}
		// split the key and value
		mapVal := strings.Split(fieldValues[i], ":")
		key, value := mapVal[0], mapVal[1]
		keyValue, err := decodeValue(keyType, key)
		if err != nil {
			return reflect.Zero(mapType), err
		}
		valueValue, err := decodeValue(valueType, value)
		if err != nil {
			return reflect.Zero(mapType), err
		}
		m.SetMapIndex(keyValue, valueValue)
	}
	// if the map is empty, return a zero value
	if m.Len() == 0 {
		m = reflect.Zero(mapType)
	}
	return m, nil
}

// decodeStruct decodes a struct from a base64 encoded string
func decodeStruct(structType reflect.Type, value string) (reflect.Value, error) {
	decodedValue, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return reflect.Zero(structType), fmt.Errorf("%v: %w", err, ErrBase64Decoding)
	}
	fieldValues := strings.Split(string(decodedValue), ",")
	structValue := reflect.New(structType).Elem()
	for i := 0; i < structType.NumField(); i++ {
		fieldType := structType.Field(i).Type
		fieldValue := fieldValues[i]
		value, err := decodeValue(fieldType, fieldValue)
		if err != nil {
			return reflect.Zero(structType), err
		}
		structValue.Field(i).Set(value)
	}
	return structValue, nil
}

func decodeValue(valueType reflect.Type, fieldValue string) (reflect.Value, error) {
	value := reflect.New(valueType).Elem()

	switch valueType.Kind() {
	case reflect.String:
		value.SetString(fieldValue)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intValue, err := strconv.ParseInt(fieldValue, 10, 64)
		if err != nil {
			return reflect.Zero(valueType), fmt.Errorf("%v: %w", err, ErrParseInt)
		}
		value.SetInt(intValue)
	case reflect.Float32, reflect.Float64:
		floatValue, err := strconv.ParseFloat(fieldValue, 64)
		if err != nil {
			return reflect.Zero(valueType), fmt.Errorf("%v: %w", err, ErrParseFloat)
		}
		value.SetFloat(floatValue)
	case reflect.Bool:
		boolValue, err := strconv.ParseBool(fieldValue)
		if err != nil {
			return reflect.Zero(valueType), fmt.Errorf("%v: %w", err, ErrParseBool)
		}
		value.SetBool(boolValue)
	case reflect.Struct:
		value, err := decodeStruct(valueType, fieldValue)
		if err != nil {
			return reflect.Zero(valueType), err
		}
		return value, nil
	case reflect.Array:
		array, err := decodeArray(valueType, fieldValue)
		if err != nil {
			return reflect.Zero(valueType), err
		}
		return array, nil

	case reflect.Slice:
		slice, err := decodeSlice(valueType, fieldValue)
		if err != nil {
			return reflect.Zero(valueType), fmt.Errorf("%v: %w", err, ErrParseSlice)
		}
		return slice, nil
	case reflect.Map:
		m, err := decodeMap(valueType, fieldValue)
		if err != nil {
			return reflect.Zero(valueType), fmt.Errorf("%v: %w", err, ErrParseMap)
		}
		return m, nil
	case reflect.Ptr:
		value, err := decodeValue(valueType.Elem(), fieldValue)
		if err != nil {
			return reflect.Zero(valueType), err
		}
		return value.Addr(), nil
	default:
		return reflect.Zero(valueType), ErrUnsupportedType
	}
	return value, nil
}

func Encode(data interface{}) (string, error) {
	// check for unsupported types
	switch reflect.TypeOf(data).Kind() {
	case reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64, reflect.Float32, reflect.Float64, reflect.Bool, reflect.Struct,
		reflect.Array, reflect.Slice, reflect.Map:
		// do nothing
	default:
		return "", ErrUnsupportedType
	}

	var fieldValues []string
	v := reflect.ValueOf(data)
	for i := 0; i < v.NumField(); i++ {
		fieldType := v.Field(i).Type()
		fieldValue := fmt.Sprintf("%v", v.Field(i).Interface())
		var err error
		switch fieldType.Kind() {
		case reflect.Struct:
			fieldValue, err = Encode(v.Field(i).Interface())
			if err != nil {
				return "", err
			}
		case reflect.Array:
			fieldValue, err = encodeArray(v.Field(i))
			if err != nil {
				return "", err
			}
		case reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Float32, reflect.Float64, reflect.Bool, reflect.Uint, reflect.Uint8, reflect.Uint16,
			reflect.Uint32, reflect.Uint64:
		// do nothing
		case reflect.Slice:
			fieldValue, err = encodeSlice(v.Field(i))
			if err != nil {
				return "", err
			}
		case reflect.Map:
			fieldValue, err = encodeMap(v.Field(i))
			if err != nil {
				return "", err
			}
		case reflect.Ptr:
			fieldValue, err = encodePtr(v.Field(i))
			if err != nil {
				return "", err
			}
		default:
			return "", ErrUnsupportedType
		}
		fieldValues = append(fieldValues, fieldValue)
	}
	return base64.StdEncoding.EncodeToString([]byte(strings.Join(fieldValues, ","))), nil
}

func encodeValue(element reflect.Value) (string, error) {
	switch element.Kind() {
	case reflect.Struct:
		fieldValue, err := Encode(element.Interface())
		if err != nil {
			return "", err
		}
		return fieldValue, nil
	case reflect.Array:
		fieldValue, err := encodeArray(element)
		if err != nil {
			return "", err
		}
		return fieldValue, nil
	case reflect.Slice:
		fieldValue, err := encodeSlice(element)
		if err != nil {
			return "", err
		}
		return fieldValue, nil
	case reflect.Map:
		fieldValue, err := encodeMap(element)
		if err != nil {
			return "", err
		}
		return fieldValue, nil
	case reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Float32, reflect.Float64, reflect.Bool, reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64:
		// do nothing
		return fmt.Sprintf("%v", element.Interface()), nil
	case reflect.Ptr:
		fieldValue, err := encodePtr(element)
		if err != nil {
			return "", err
		}
		return fieldValue, nil
	default:
		return "", ErrUnsupportedType
	}
}

// encodePtr encodes pointer to string
func encodePtr(ptr reflect.Value) (string, error) {
	if ptr.IsNil() {
		return "", nil
	}
	return encodeValue(ptr.Elem())
}

// encodeArray encodes array to string
func encodeArray(array reflect.Value) (string, error) {
	var fieldValues []string
	for i := 0; i < array.Len(); i++ {
		fieldValue, err := encodeValue(array.Index(i))
		if err != nil {
			return "", err
		}
		fieldValues = append(fieldValues, fieldValue)
	}
	return base64.StdEncoding.EncodeToString([]byte(strings.Join(fieldValues, ","))), nil
}

// encodeSlice encodes slice to string
func encodeSlice(slice reflect.Value) (string, error) {
	var fieldValues []string
	for i := 0; i < slice.Len(); i++ {
		fieldValue, err := encodeValue(slice.Index(i))
		if err != nil {
			return "", err
		}
		fieldValues = append(fieldValues, fieldValue)
	}
	return base64.StdEncoding.EncodeToString([]byte(strings.Join(fieldValues, ","))), nil
}

// encodeMap encodes map to string
func encodeMap(m reflect.Value) (string, error) {
	var fieldValues []string
	for _, key := range m.MapKeys() {
		fieldValue, err := encodeValue(m.MapIndex(key))
		if err != nil {
			return "", err
		}
		keyValue, err := encodeValue(key)
		if err != nil {
			return "", err
		}
		fieldValues = append(fieldValues, fmt.Sprintf("%v:%v", keyValue, fieldValue))
	}
	return base64.StdEncoding.EncodeToString([]byte(strings.Join(fieldValues, ","))), nil
}
