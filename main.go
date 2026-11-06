// gmx scaffolds Go + Gin backend projects.
//
//	gmx build [dir] [--module name] [--prefix /api/v1] [--db supabase] [--payment stripe] [--skip-install] [--force]
package main

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"
)

// version can be overridden with: go build -ldflags "-X main.version=1.2.3".
// With `go install ...@vX.Y.Z` it falls back to the module version.
var version = "dev"

func init() {
	if version != "dev" {
		return
	}
	if bi, ok := debug.ReadBuildInfo(); ok && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
		version = strings.TrimPrefix(bi.Main.Version, "v")
	}
}

const usage = `gmx - Go + Gin project generator

Usage:
  gmx build [dir] [flags]   e.g. gmx build myapi --db supabase --payment stripe
                            create a new Gin project in dir (default: backend)
  gmx version               print the gmx version
  gmx help                  show this help

Flags for build:
  --module string    Go module path of the new project (default: the directory name)
  --prefix string    parent route group prefix (default: /api/v1)
  --db string        database provider: supabase (default: none)
  --payment string   payment processor: stripe (default: none)
  --skip-install     do not run 'go get' / 'go mod tidy' (no network needed)
  --force            write into a non-empty directory
`

func main() {
	if len(os.Args) < 2 {
		fmt.Print(usage)
		os.Exit(2)
	}
	switch os.Args[1] {
	case "build":
		if err := runBuild(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "gmx: error:", err)
			os.Exit(1)
		}
	case "version", "--version", "-v":
		fmt.Println("gmx", version)
	case "help", "--help", "-h":
		fmt.Print(usage)
	default:
		fmt.Fprintf(os.Stderr, "gmx: unknown command %q\n\n%s", os.Args[1], usage)
		os.Exit(2)
	}
}
