package waffle

import (
	"testing"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-billy/v5/util"
)

func TestTemplate_MemFS(t *testing.T) {
	f := memfs.New()
	write(f, "template.json", `{
		"name": "Test" 
	}`)
	write(f, "test", `Hello {{ .name }}!`)
	write(f, "a/b/c/d", `content`)
	write(f, "templated_{{ .name }}", `From naming template`)

	template := newTemplate(f)

	out := memfs.New()
	err := template.Generate(out)
	if err != nil {
		t.Fatal("cannot generate contents", err)
	}

	expects := []struct {
		path    string
		content string
	}{
		{path: "test", content: "Hello Test!"},
		{path: "templated_Test", content: "From naming template"},
		{path: "a/b/c/d", content: "content"},
	}
	for _, e := range expects {
		_, err = out.Stat(e.path)
		if err != nil {
			t.Errorf("cannot find file. path:%s, %s", e.path, err)
		}
	}
}

func TestTemplate_OSFS(t *testing.T) {
	f := osfs.New("testdata")

	template := newTemplate(f)

	out := memfs.New()
	err := template.Generate(out)
	if err != nil {
		t.Fatal("cannot generate contents", err)
	}

	expects := []struct {
		path    string
		content string
	}{
		{path: "test", content: "Hello Test!"},
		{path: "templated_Test", content: "From naming template"},
		{path: "a/b/c/d", content: "content"},
	}
	for _, e := range expects {
		_, err = out.Stat(e.path)
		if err != nil {
			t.Errorf("cannot find file. path:%s, %s", e.path, err)
		}
	}
}

func write(fs billy.Filesystem, path string, content string) {
	util.WriteFile(fs, path, []byte(content), 0755)
}
