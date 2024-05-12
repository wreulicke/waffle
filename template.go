package waffle

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-billy/v5/util"
	"github.com/influxdata/go-prompt"
)

type env map[string]any

type Template struct {
	template billy.Filesystem
}

func OpenTemplate(path string) *Template {
	fs := osfs.New(path)
	return newTemplate(fs)
}

func newTemplate(template billy.Filesystem) *Template {
	return &Template{template: template}
}

func (t *Template) Generate(out billy.Filesystem) error {
	f, err := t.template.Open("template.json")
	if err != nil {
		return fmt.Errorf("cannot read template.json: %w", err)
	}
	defer f.Close()

	s, err := newSettings(f)
	if err != nil {
		return fmt.Errorf("cannot create settings: %w", err)
	}

	env, err := s.Generate()
	if err != nil {
		return fmt.Errorf("cannot generate settings: %w", err)
	}

	return util.Walk(t.template, ".", func(path string, d fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || path == "template.json" {
			return nil
		}

		templateFile, err := t.template.Open(path)
		if err != nil {
			return fmt.Errorf("cannot open template from path:%s: %w", path, err)
		}
		bs, err := io.ReadAll(templateFile)
		if err != nil {
			return fmt.Errorf("cannot read file. path: %s: %w", path, err)
		}

		var b bytes.Buffer
		err = generate(path, env, &b)
		if err != nil {
			return fmt.Errorf("cannot expand path %s: %w", path, err)
		}

		outPath := b.String()

		f, err := out.OpenFile(outPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			return fmt.Errorf("cannot open file path: %s: %w", outPath, err)
		}
		defer f.Close()

		err = generate(string(bs), env, f)
		if err != nil {
			return fmt.Errorf("cannot generate content from template. template:%s, output: %s, : %w", path, outPath, err)
		}
		return nil
	})
}

func generate(tpl string, e env, w io.Writer) error {
	// TODO fix funcs
	t, err := template.New("template").Funcs(funcs()).Parse(tpl)
	if err != nil {
		return err
	}
	return t.Execute(w, e)
}

type settings struct {
	template *template.Template
}

func input(name string) string {
	return prompt.Input(name+": ", func(d prompt.Document) []prompt.Suggest {
		return []prompt.Suggest{}
	}, prompt.OptionPrefixTextColor(prompt.Green))
}

func choose(name string, opts ...string) string {
	return prompt.Choose(name+": ", opts, prompt.OptionPrefixTextColor(prompt.Green))
}

func funcs() template.FuncMap {
	m := sprig.TxtFuncMap()
	m["i"] = input
	m["input"] = input
	m["choose"] = choose
	m["select"] = choose
	return m
}

func newSettings(r io.Reader) (*settings, error) {
	bs, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("cannot read template.json: %w", err)
	}
	t, err := template.New("template").Funcs(funcs()).Parse(string(bs))
	if err != nil {
		return nil, fmt.Errorf("cannot parse template.json: %w", err)
	}
	return &settings{template: t}, nil
}

func (s *settings) Generate() (env, error) {
	var b bytes.Buffer
	err := s.template.Execute(&b, nil)
	if err != nil {
		return nil, err
	}
	var m env
	err = json.Unmarshal(b.Bytes(), &m)
	return m, err
}
