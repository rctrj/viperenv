package viperenv

import (
	"bytes"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type ConfigType string

var (
	ConfigTypeJson ConfigType = "json"
	ConfigTypeYaml ConfigType = "yaml"
	ConfigTypeYml  ConfigType = "yml"
	ConfigTypeToml ConfigType = "toml"
)

type ViperConfig struct {
	EnvKey     string
	DefaultEnv string

	EnvPrefix string //prefix for all env variables

	ConfigType    ConfigType
	KeySecretsDir string //directory in which secrets are stored
}

// NewFromFS reads in and unmarshalls configuration into the target struct.
// returns the retrieved environment and the config directory path
func NewFromFS(cfg ViperConfig, files fs.FS, target interface{}, baseName string) {
	//default env
	viper.SetDefault(cfg.EnvKey, cfg.DefaultEnv)

	//setup to read env variables
	viper.SetEnvPrefix(cfg.EnvPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	env := Env(viper.GetString(cfg.EnvKey))

	//reads secrets from env. This way, secrets are not stored on git
	secretsDir := os.Getenv(cfg.KeySecretsDir)
	secrets, err := getSecretsFromSecretStoreFiles(secretsDir, cfg.EnvPrefix)
	if err != nil {
		log.Fatal("error getting secrets from secret store files", err)
	}

	//replace values in config with secrets
	for secretName, secret := range secrets {
		viper.Set(secretName, secret)
	}

	//adds file extension
	withExt := func(file string) string {
		var b strings.Builder
		b.WriteString(file)
		b.Write([]byte("."))
		b.WriteString(cfg.ConfigType.String())

		return b.String()
	}

	readFile := func(path string) []byte {
		read, err := fs.ReadFile(files, withExt(path))
		if err != nil {
			log.Fatal("error reading file : "+path, err)
		}
		return read
	}

	baseConfig := readFile(baseName)
	envConfig := readFile(env.string())

	// because there is no file extension in a stream of bytes, supported extensions are "json", "toml", "yaml", "yml", "properties", "props", "prop"
	viper.SetConfigType(cfg.ConfigType.String())

	if err := viper.ReadConfig(bytes.NewBuffer(baseConfig)); err != nil {
		log.Fatal("Failed to read base config", err)
	}

	if err := viper.MergeConfig(bytes.NewBuffer(envConfig)); err != nil {
		log.Fatal("Failed to merge env config", err)
	}

	if err := viper.Unmarshal(target); err != nil {
		log.Fatal("failed to unmarshal", err)
	}
}

//reads secrets files
/*
	How to use:
		For each key you wish to replace, create a new secret file.
	Example:
		To replace a variable called testSecretValue in config
		Create a file called `x_testSecretValue and fill the file if your keySecretsDir is x. Example:
			*x_testSecretValue
				some_value_that_is_string
		Then set the env variable to point at this directory
		That's it
*/
func getSecretsFromSecretStoreFiles(secretsDir, envPrefix string) (map[string]string, error) {
	prefix := strings.ToUpper(envPrefix) + "_"
	secrets := make(map[string]string)

	if secretsDir == "" {
		return secrets, nil
	}

	//alternative to terminal command `cd`. Starts at `~`
	err := filepath.WalkDir(secretsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return errors.WithMessagef(err, "error reading file/dir. Path: %v", path)
		}

		//if directory, move ahead
		if d.IsDir() {
			return nil
		}

		fileName := d.Name()
		if !strings.HasPrefix(fileName, prefix) {
			// do not process files that don't start with prefix
			return nil
		}

		fileBytes, err := os.ReadFile(path)
		if err != nil {
			return errors.WithMessagef(err, "error reading secret file. File: %v", path)
		}

		secretName := strings.TrimPrefix(fileName, prefix)
		hierarchicalSecretName := strings.Replace(secretName, "_", ".", -1)
		secrets[hierarchicalSecretName] = string(fileBytes)

		return nil
	})

	if err != nil {
		return secrets, errors.WithStack(err)
	}

	return secrets, nil
}

func (c ConfigType) String() string {
	return string(c)
}
