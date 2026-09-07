package env

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
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

		switch fieldValue.Kind() {
		case reflect.String:
			fieldValue.SetString(envValue)

		case reflect.Int:
			intValue, err := strconv.Atoi(envValue)
			if err != nil {
				return fmt.Errorf(
					"parsing %s env value: %w",
					envPrefix,
					err,
				)
			}

			fieldValue.SetInt(int64(intValue))

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
