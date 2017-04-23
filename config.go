package yanktv

import (
	"github.com/BurntSushi/toml"
)

type DatabaseConf struct {
	Path string
}

type Conf struct {
	Verbose         bool
	DownloadTimeout int // in seconds
	Shows           []string
	OutputFile      string

	Database DatabaseConf
}

func LoadConf(path string) (Conf, error) {
	var c Conf
	_, err := toml.DecodeFile(path, &c)
	if err != nil {
		return Conf{}, err
	}
	return c, nil
}
