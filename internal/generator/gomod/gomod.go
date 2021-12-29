package gomod

import (
	_ "embed"
	"errors"
	"io/fs"

	"gitlab.com/mnm/bud/gen"
	"gitlab.com/mnm/bud/go/mod"
	"gitlab.com/mnm/bud/internal/gotemplate"
)

//go:embed gomod.gotext
var template string

var generator = gotemplate.MustParse("gomod", template)

type Go struct {
	Version string
}

type Require struct {
	Path     string
	Version  string
	Indirect bool
}

type Replace struct {
	Old string
	New string
}

type Generator struct {
	FS        fs.FS
	ModFinder *mod.Finder
	Go        *Go
	Requires  []*Require
	Replaces  []*Replace
}

func (g *Generator) GenerateFile(_ gen.F, file *gen.File) error {
	code, err := fs.ReadFile(g.FS, "go.mod")
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return err
		}
		return g.createFile(file)
	}
	return g.updateFile(file, code)
}

func (g *Generator) updateFile(file *gen.File, code []byte) error {
	module, err := mod.New().Parse("go.mod", code)
	if err != nil {
		return err
	}
	modFile := module.File()
	// Add any additional requires and replaces if they don't exist already
	for _, require := range g.Requires {
		if err := modFile.AddRequire(require.Path, require.Version); err != nil {
			return err
		}
	}
	for _, replace := range g.Replaces {
		if err := modFile.AddReplace(replace.Old, "", replace.New, ""); err != nil {
			return err
		}
	}
	file.Write(modFile.Format())
	return nil
}

type State struct {
	Module   *mod.Module
	Go       *Go
	Requires []*Require
	Replaces []*Replace
}

func (g *Generator) createFile(file *gen.File) error {
	module, err := g.ModFinder.Parse("go.mod", []byte("module app.com"))
	if err != nil {
		return err
	}
	code, err := generator.Generate(&State{
		Module:   module,
		Go:       g.Go,
		Requires: g.Requires,
		Replaces: g.Replaces,
	})
	if err != nil {
		return err
	}
	file.Write(code)
	return nil
}
