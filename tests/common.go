package tests

import (
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/onsi/ginkgo"
)

func GenerateGUID(prefix string) string {
	guid := uuid.NewString()

	return fmt.Sprintf("%s-%s", prefix, guid[:13])
}

func GetEnvOrSkip(env string) string {
	value := os.Getenv(env)
	if value == "" {
		ginkgo.Skip(fmt.Sprintf("%s not exported", env))
	}

	return value
}
