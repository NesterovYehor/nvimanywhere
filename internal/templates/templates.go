package templates

import (
	"fmt"
	"html/template"
	"io/fs"
	webfs "nvimanywhere"
	"path/filepath"
	"strings"
)

type TemplateCache map[string]*template.Template

func NewTemplateCache() (TemplateCache, error) {
	tc := make(TemplateCache)
	sub, err := fs.Sub(webfs.TmplFS, "web/templates")
	if err != nil {
		return nil, fmt.Errorf("subfs: %w", err)
	}
	err = fs.WalkDir(sub, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".html") {
			return nil
		}
		k := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))

		t, err := template.ParseFS(sub, path)
		if err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
		tc[k] = t
		return nil
	})
	if err != nil {
		return nil, err
	}
	return tc, nil
}
