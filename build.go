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
	DB     string // database provider: "" (none) or "supabase"
}

// dbProviders are the values accepted by --db; dbDeps are the extra modules each needs.
var dbProviders = []string{"supabase"}

var dbDeps = map[string][]string{
	"supabase": {"github.com/supabase-community/supabase-go@latest"},
}

func contains(l []string, s string) bool {
	for _, v := range l {
		if v == s {
			return true
		}
	}
	return false
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
	db := fset.String("db", "", "database provider: supabase (default: none)")

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
	if *db != "" && !contains(dbProviders, *db) {
		return fmt.Errorf("unknown --db %q (supported: %s)", *db, strings.Join(dbProviders, ", "))
	}
	data := Data{Module: *module, Prefix: p, DB: *db}

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
		if d.Name() == ".DS_Store" {
			return nil
		}
		rel := strings.TrimPrefix(p, "templates/")
		// database templates are opt-in: database/*.tmpl and database/<provider>/ for the chosen one
		if strings.HasPrefix(rel, "database/") {
			parts := strings.Split(rel, "/")
			if data.DB == "" || (len(parts) > 2 && parts[1] != data.DB) {
				return nil
			}
		}
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

	if err := goMod(dir, data, *skip); err != nil {
		return err
	}

	fmt.Printf("\nDone. Next:\n  cd %s\n", dir)
	if data.DB != "" {
		fmt.Println("  cp .env.example .env   # then fill in the database credentials")
	}
	fmt.Printf("  go run .\n  curl localhost:8080%s/test/ping\n", data.Prefix)
	if data.DB != "" {
		fmt.Printf("  curl localhost:8080%s/test/db\n", data.Prefix)
	}
	return nil
}

func goMod(dir string, data Data, skip bool) error {
	module := data.Module
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
	deps := append([]string{}, goDeps...)
	deps = append(deps, dbDeps[data.DB]...)
	fmt.Println("installing dependencies ...")
	if err := runIn(dir, "go", append([]string{"get"}, deps...)...); err != nil {
		fmt.Printf("\nwarning: could not install dependencies (%v).\nThe project files were created; run this inside %s when you are online:\n  go get %s && go mod tidy\n", err, dir, strings.Join(deps, " "))
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
