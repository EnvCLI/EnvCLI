package config

import (
	"github.com/rs/zerolog/log"
	"os"
	"reflect"
)

/**
 * Get the working directory
 */
func GetWorkingDirectory() string {
	workingDir, err := os.Getwd()
	if err != nil {
		log.Fatal().Err(err).Msg("Couldn't detect working directory!")
	}

	return workingDir
}

/**
 * Checks if a object is part of a array
 */
func InArray(val interface{}, array interface{}) (exists bool, index int) {
	exists = false
	index = -1

	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(array)

		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(val, s.Index(i).Interface()) == true {
				index = i
				exists = true
				return
			}
		}
	}

	return
}

/**
 * GetOrDefault gets the value or the specified default value if not found or empty
 */
func GetOrDefault(entity map[string]string, key string, defaultValue string) (val string) {
	value, found := entity[key]

	if found {
		return value
	}

	return defaultValue
}
