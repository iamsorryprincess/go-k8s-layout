package env

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func Parse[TConfig any]() (TConfig, error) {
	var config TConfig

	v := reflect.ValueOf(&config).Elem()

	if v.Kind() != reflect.Struct {
		return config, fmt.Errorf("Config must be a struct")
	}

	if err := parseStruct("", v); err != nil {
		return config, err
	}

	return config, nil
}

func parseStruct(prefix string, v reflect.Value) error {
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		fieldType := t.Field(i)
		fieldValue := v.Field(i)

		tagValue := fieldType.Tag.Get("env")
		if tagValue == "" {
			continue
		}

		tagValues := strings.Split(tagValue, ",")
		if len(tagValues) == 0 {
			return fmt.Errorf("invalid env tag values for %s", tagValue)
		}

		tagValue = tagValues[0]
		envPrefix := prefix + tagValue

		if fieldValue.Kind() == reflect.Struct {
			if err := parseStruct(envPrefix+"_", fieldValue); err != nil {
				return fmt.Errorf(
					"parsing struct %s with env prefix %s: %w",
					fieldType.Name,
					envPrefix,
					err,
				)
			}
			continue
		}

		envValue := os.Getenv(envPrefix)
		if envValue == "" {
			if len(tagValues) < 1 {
				continue
			}
			envValue = tagValues[1]
		}

		switch fieldValue.Kind() {
		case reflect.String:
			fieldValue.SetString(envValue)

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if fieldType.Type.Name() == "Duration" {
				duration, err := time.ParseDuration(envValue)
				if err != nil {
					return fmt.Errorf(
						"parsing %s env time.Duration value: %w",
						envPrefix,
						err,
					)
				}

				fieldValue.SetInt(int64(duration))
				continue
			}

			intValue, err := parseInt(envValue, envPrefix, fieldValue.Kind())
			if err != nil {
				return err
			}

			fieldValue.SetInt(intValue)

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			uintValue, err := parseUint(envValue, envPrefix, fieldValue.Kind())
			if err != nil {
				return err
			}

			fieldValue.SetUint(uintValue)

		case reflect.Float32, reflect.Float64:
			floatValue, err := parseFloat(envValue, envPrefix, fieldValue.Kind())
			if err != nil {
				return err
			}

			fieldValue.SetFloat(floatValue)

		case reflect.Bool:
			fieldValue.SetBool(envValue == "true")

		default:
			return fmt.Errorf(
				"unsupported field type %s for %s",
				fieldValue.Kind(),
				envPrefix,
			)
		}
	}

	return nil
}

func parseInt(value string, envPrefix string, reflectKind reflect.Kind) (int64, error) {
	var bitSize int
	var typeName string
	switch reflectKind {
	case reflect.Int:
		bitSize = 0
		typeName = "int"
	case reflect.Int8:
		bitSize = 8
		typeName = "int8"
	case reflect.Int16:
		bitSize = 16
		typeName = "int16"
	case reflect.Int32:
		bitSize = 32
		typeName = "int32"
	case reflect.Int64:
		bitSize = 64
		typeName = "int64"
	default:
		return 0, fmt.Errorf("parsing %s env value: invalid kind %s for int value", envPrefix, reflectKind)
	}

	result, err := strconv.ParseInt(value, 10, bitSize)
	if err != nil {
		return 0, fmt.Errorf(
			"parsing %s env %s value: %w",
			envPrefix,
			typeName,
			err,
		)
	}

	return result, nil
}

func parseUint(value string, envPrefix string, reflectKind reflect.Kind) (uint64, error) {
	var bitSize int
	var typeName string
	switch reflectKind {
	case reflect.Uint:
		bitSize = 0
		typeName = "uint"
	case reflect.Uint8:
		bitSize = 8
		typeName = "uint8"
	case reflect.Uint16:
		bitSize = 16
		typeName = "uint16"
	case reflect.Uint32:
		bitSize = 32
		typeName = "uint32"
	case reflect.Uint64:
		bitSize = 64
		typeName = "uint64"
	default:
		return 0, fmt.Errorf("parsing %s env value: invalid kind %s for uint value", envPrefix, reflectKind)
	}

	result, err := strconv.ParseUint(value, 10, bitSize)
	if err != nil {
		return 0, fmt.Errorf(
			"parsing %s env %s value: %w",
			envPrefix,
			typeName,
			err,
		)
	}

	return result, nil
}

func parseFloat(value string, envPrefix string, reflectKind reflect.Kind) (float64, error) {
	var bitSize int
	var typeName string
	switch reflectKind {
	case reflect.Float32:
		bitSize = 32
		typeName = "float32"
	case reflect.Float64:
		bitSize = 64
		typeName = "float64"
	default:
		return 0, fmt.Errorf("parsing %s env value: invalid kind %s for float value", envPrefix, reflectKind)
	}

	result, err := strconv.ParseFloat(value, bitSize)
	if err != nil {
		return 0, fmt.Errorf(
			"parsing %s env %s value: %w",
			envPrefix,
			typeName,
			err,
		)
	}

	return result, nil
}
