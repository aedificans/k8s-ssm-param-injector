/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

import (
	"log"
	"os"
	"strconv"
)

func GetEnvBool(key string, defaultValue bool) bool {
	envVarValue := os.Getenv(key)
	if envVarValue == "" {
		return defaultValue
	}

	value, err := strconv.ParseBool(envVarValue)
	if err != nil {
		log.Fatal(err)
	}

	return value
}

func GetEnvInt(key string, defaultValue int) int {
	envVarValue := os.Getenv(key)
	if envVarValue == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(envVarValue)
	if err != nil {
		log.Fatal(err)
	}

	return value
}

func GetEnvString(key string, defaultValue string) string {
	val := os.Getenv(key)
	if val == "" {
		val = defaultValue
	}
	return val
}
