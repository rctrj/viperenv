package viperenv

import "os"

type Env string

const (
	Dev     Env = "dev"
	Testing Env = "test"
	Staging Env = "staging"
	Prod    Env = "prod"
)

func (e Env) IsDev() bool     { return e == Dev }
func (e Env) IsTesting() bool { return e == Testing }
func (e Env) IsStaging() bool { return e == Staging }
func (e Env) IsProd() bool    { return e == Prod }

func (e Env) string() string {
	return string(e)
}

func ExtractEnvVariableFromOs(key, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		val = defaultVal
	}
	return val
}
