# viperenv
A viper implementation for reading config variables

Sample Usage:
All config files are in yaml and present in `conf` folder \
All secrets are stored either in env variables or `secrets` folder

```
import (
	"embed"
	"github.com/rctrj/viperenv"
	"io/fs"
)

//go:embed conf
var conf embed.FS

type Config struct {
	Env viperenv.Env
}

func newConfig() *Config {
	viperConf := viperenv.ViperConfig{
		EnvKey:        "env",
		DefaultEnv:    "dev",
		EnvPrefix:     "sample",
		ConfigType:    "yaml",
		KeySecretsDir: "secrets",
	}

	c := &Config{}
	files, _ := fs.Sub(conf, "conf")
	viperenv.NewFromFS(viperConf, files, c, "world")
	return c
}
```