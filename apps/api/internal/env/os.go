package env

import "os"

func envGet(key string) string {
	return os.Getenv(key)
}
