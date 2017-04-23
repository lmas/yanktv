package yanktv

import (
	"net/url"
	"strings"

	"github.com/BurntSushi/toml"
)

type DatabaseConf struct {
	Path string
}

type Conf struct {
	Verbose         bool
	DownloadTimeout int // in seconds
	TorrentURL      string
	Shows           []string

	Database DatabaseConf
}

// TODO: remove the pointer usage?
func LoadConf(path string) (*Conf, error) {
	var c *Conf
	_, err := toml.DecodeFile(path, &c)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Conf) searchUrl(search string) string {
	s := url.QueryEscape(strings.TrimSpace(search))
	//if *fDebug {
	//return s
	//}
	return strings.Replace(c.TorrentURL, "%SEARCH", s, -1)
}
