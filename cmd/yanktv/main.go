package main

import (
	"html/template"
	"os"
	"path/filepath"
	"sort"

	"github.com/lmas/yanktv"
	_ "github.com/lmas/yanktv/sites"
)

type TorrentGroup struct {
	Date     string
	Torrents []yanktv.Torrent
}

func groupTorrents(torrents []yanktv.Torrent) []TorrentGroup {
	// Since you can't sort maps...
	tmp := make(map[string][]yanktv.Torrent)
	for _, t := range torrents {
		d := t.Timestamp.Format("2006-01-02")
		tmp[d] = append(tmp[d], t)
	}

	// ...I got to convert to a custom struct instead...
	var groups []TorrentGroup
	for k, v := range tmp {
		d := TorrentGroup{k, v}
		groups = append(groups, d)
	}

	// ...which I then can do a reverse sort on
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Date > groups[j].Date
	})
	return groups
}

func handleErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	funcMap := template.FuncMap{
		"safeUrl": func(s string) template.URL {
			return template.URL(s)
		},
	}
	tmpl := template.Must(template.New("base").Funcs(funcMap).Parse(tmplBase))

	c, err := yanktv.LoadConf(".yanktv.conf") // TODO: add flag for custom path
	handleErr(err)

	err = os.MkdirAll(filepath.Dir(c.OutputFile), 0777)
	handleErr(err)
	f, err := os.Create(c.OutputFile)
	handleErr(err)
	defer f.Close()

	app, err := yanktv.New(c)
	handleErr(err)

	//err = app.UpdateShows()
	//handleErr(err)

	torrents, err := app.GetTorrentsFromLastMonth()
	handleErr(err)

	err = tmpl.Execute(f, map[string]interface{}{
		"Groups": groupTorrents(torrents),
	})
	handleErr(err)
}

const tmplBase string = `<!DOCTYPE html>
<html>
	<head>
		<title>Latest episodes</title>
		<style type="text/css">
			body {
				margin: 20px 40px;
				line-height: 1.5;
				font-size: 16px;
				font-family: monospace;
				background-color: #FFF;
				color: #444;
			}
			a, a:hover, a:visited {
				text-decoration:none;
				color: #444;
			}
			a:hover {
				color: #000;
			}
			article {
				border: 1px solid #DDD;
				background-color: #FFF;
				border-radius: 3px;
				margin-bottom: 20px;
			}
			article > h2 {
				background-color: #EEE;
				border-bottom: 1px solid #DDD;
				padding: 10px 15px;
				margin: 0;
				font-size: 16px;
				font-weight: bold;
				text-align: center;
			}
			article > ol, p {
				margin: 15px;
			}
			article > ol > li > a:visited {
				color: #AAA;
			}
			footer {
				text-align: center;
				font-size: 12px;
			}
			.center {
				text-align: center;
			}
		</style>
	</head>
	<body>
		<section id="content">
		{{range .Groups}}
			<article>
				<h2>{{.Date}}</h2>
				<ol>{{range .Torrents}}
					<li><a href="{{.MagnetUrl | safeUrl}}">{{.Title}}</a></li>
				{{- end}}
				</ol>
			</article>
		{{else}}
			<p class="center">Sorry, no episodes in the database!</p>
		{{- end}}
		</section>
		<footer>
			Generated with <a href="https://github.com/lmas/yanktv">Yanktv</a>
		</footer>
	</body>
</html>
`
