package novelrss

import (
	"github.com/gobuffalo/packr/v2"
	"html/template"
	"net/http"
	"strings"
)

type binaryFileSystem struct {
	fs http.FileSystem
}

func (b *binaryFileSystem) Open(name string) (http.File, error) {
	return b.fs.Open(name)
}

func (b *binaryFileSystem) Exists(prefix string, filepath string) bool {

	if p := strings.TrimPrefix(filepath, prefix); len(p) < len(filepath) {
		if _, err := b.fs.Open(p); err != nil {
			return false
		}
		return true
	}
	return false
}

func WebAssetsDir(path string) *binaryFileSystem {
	fs := packr.New("assets", "./web/statics")
	//fs := packr.New("assets", "|")
	//fs.ResolutionDir = path
	return &binaryFileSystem{
		fs,
	}
}

func LoadTemplate(path string) (*template.Template, error) {
	t := template.New("")
	box := packr.New("templates", "./web/templates")
	//box := packr.New("templates", "|")
	//box.ResolutionDir = path
	files := box.List()
	for _, name := range files {
		if box.HasDir(name) || !strings.HasSuffix(name, ".tmpl") {
			continue
		}
		tmpl, err := box.FindString(name)
		if err != nil {
			return nil, err
		}
		t, err = t.New(name).Parse(tmpl)
		if err != nil {
			return nil, err
		}
	}
	return t, nil
}
