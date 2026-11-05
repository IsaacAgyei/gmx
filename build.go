package main

import (
	"embed"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed all:templates
var templatesFS embed.FS

// Data is what every template receives.
type Data struct {
	Module string // Go module path, e.g. "backend"
	Prefix string // parent route group, e.g. "/api/v1"
}

// goDeps are installed into the new project with `go get`.
var goDeps = []string{
	"github.com/gin-gonic/gin@latest",
	"github.com/gin-contrib/cors@latest",
	"github.com/joho/godotenv@latest",
}

func runBuild(args []string) error {
	fset := flag.NewFlagSet("build", flag.ContinueOnError)
	module := fset.String("module", "", "Go module path")
	prefix := fset.String("prefix", "/api/v1", "parent route group prefix")
	skip := fset.Bool("skip-install", false, "skip go get / go mod tidy")
	force := fset.Bool("force", false, "allow a non-empty target directory")

	// Allow the directory argument before or after the flags.
	var positional []string
	rest := args
	for len(rest) > 0 {
		if err := fset.Parse(rest); err != nil {
			return err
		}
		rest = fset.Args()
		if len(rest) > 0 {
			positional = append(positional, rest[0])
			rest = rest[1:]
		}
	}
	if len(positional) > 1 {
		return errors.New("build takes at most one directory argument")
	}

	dir := "backend"
	if len(positional) == 1 {
		dir = positional[0]
	}
	if *module == "" {
		*module = filepath.Base(filepath.Clean(dir))
	}
	p := "/" + strings.Trim(*prefix, "/")
	data := Data{Module: *module, Prefix: p}

	if entries, err := os.ReadDir(dir); err == nil && len(entries) > 0 && !*force {
		return fmt.Errorf("%s is not empty (use --force to write anyway)", dir)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	count := 0
	err := fs.WalkDir(templatesFS, "templates", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel := strings.TrimPrefix(p, "templates/")
		rel = strings.TrimSuffix(rel, ".tmpl")
		rel = path.Clean(rel)
		// dotfiles are stored without the dot so embed/git don't treat them specially
		if strings.HasPrefix(rel, "dot_") {
			rel = "." + strings.TrimPrefix(rel, "dot_")
		}
		raw, err := templatesFS.ReadFile(p)
		if err != nil {
			return err
		}
		tpl, err := template.New(rel).Parse(string(raw))
		if err != nil {
			return fmt.Errorf("template %s: %w", p, err)
		}
		out := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
			return err
		}
		f, err := os.Create(out)
		if err != nil {
			return err
		}
		defer f.Close()
		count++
		return tpl.Execute(f, data)
	})
	if err != nil {
		return err
	}
	fmt.Printf("created %d files in %s (module %q, prefix %q)\n", count, dir, data.Module, data.Prefix)

	if err := goMod(dir, data.Module, *skip); err != nil {
		return err
	}

	fmt.Printf("\nDone. Next:\n  cd %s\n  go run .\n  curl localhost:8080%s/test/ping\n", dir, data.Prefix)
	return nil
}

func goMod(dir, module string, skip bool) error {
	if _, err := exec.LookPath("go"); err != nil {
		fmt.Println("go not found on PATH; skipping module setup. Run `go mod init` and `go mod tidy` yourself.")
		return nil
	}
	if err := runIn(dir, "go", "mod", "init", module); err != nil {
		return err
	}
	if skip {
		fmt.Println("skipped dependency install (--skip-install). Run `go mod tidy` when online.")
		return nil
	}
	fmt.Println("installing gin, cors, godotenv ...")
	if err := runIn(dir, "go", append([]string{"get"}, goDeps...)...); err != nil {
		fmt.Printf("\nwarning: could not install dependencies (%v).\nThe project files were created; run this inside %s when you are online:\n  go get %s && go mod tidy\n", err, dir, strings.Join(goDeps, " "))
		return nil
	}
	if err := runIn(dir, "go", "mod", "tidy"); err != nil {
		fmt.Printf("\nwarning: go mod tidy failed (%v); run it yourself inside %s\n", err, dir)
		return nil
	}
	return runIn(dir, "gofmt", "-l", "-w", ".")
}

func runIn(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s %s: %w", name, strings.Join(args, " "), err)
	}
	return nil
}
