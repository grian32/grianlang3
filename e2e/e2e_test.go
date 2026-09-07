package e2e

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var compiler, repo string

func TestMain(m *testing.M) {
	// Keep unit-only runs cheap and avoid requiring clang in short mode.
	flag.Parse()
	if testing.Short() {
		os.Exit(m.Run())
	}
	code := func() int {
		var err error
		repo, err = filepath.Abs("..")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		dir, err := os.MkdirTemp("", "gl3-e2e-")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		defer os.RemoveAll(dir)
		compiler = filepath.Join(dir, "gl3")
		// Same build path as make: timestamp-based stdlib generation, then one
		// compiler build. Building in the checkout preserves Go's build cache.
		for _, command := range [][]string{{"make", "stdlib"}, {"go", "build", "-o", compiler, "."}} {
			result := run(repo, nil, "", 2*time.Minute, command[0], command[1:]...)
			if result.err != nil {
				fmt.Fprintf(os.Stderr, "e2e setup %v failed: %v\n%s%s", command, result.err, result.stdout, result.stderr)
				return 1
			}
		}
		return m.Run()
	}()
	os.Exit(code)
}

func TestPrograms(t *testing.T) {
	if testing.Short() {
		t.Skip("e2e tests require building the compiler and clang")
	}
	// Defaults: compile main.gl3 successfully, run with empty stdin, exit 0.
	// Expected stdout and stderr are fixture files; absent files mean empty.
	cases := []struct {
		name     string
		sources  []string
		args     []string
		stdin    string
		exit     int
		optimize bool
	}{
		{name: "unsigned_casts", optimize: true},
		{name: "unsigned_same_width", optimize: true},
		{name: "unsigned_float_cast", optimize: true},
		{name: "unsigned_arithmetic", optimize: true},
		{name: "array_resize_literal"},
		{name: "array_resize_push"},
		{name: "nested_loops", optimize: true},
		{name: "loop_return", optimize: true},
		{name: "branch_returns", optimize: true},
		{name: "struct_layout", optimize: true},
		{name: "address_of_field", optimize: true},
		{name: "address_of_index", optimize: true},
		{name: "address_of_dereference", optimize: true},
		{name: "import_types_constants", sources: []string{"main.gl3", "model.gl3"}},
		{name: "strings"},
		{name: "print_type_warning"},
		{name: "print_count_warning"},
		{name: "constant_assignment_warning"},
		{name: "missing_builtin_import"},
		{name: "hello"},
		{name: "arithmetic_casts"},
		{name: "control_flow"},
		{name: "pointers_structs"},
		{name: "arrays"},
		{name: "imports", sources: []string{"main.gl3", "math.gl3"}},
		{name: "stdin", stdin: "Z", exit: 7},
		{name: "parser_error"},
		{name: "unknown_function"},
		{name: "linker_error"},
		{name: "conflicting_flags", args: []string{"--O1", "--O2"}},
	}
	registered := make(map[string]bool)
	for _, test := range cases {
		registered[test.name] = true
	}
	entries, err := os.ReadDir("testdata")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() && !registered[entry.Name()] {
			t.Errorf("fixture %s is not registered", entry.Name())
		}
	}
	// Each optimized variant uses the same expected results as its default build.
	// These fixtures use no builtin imports: RunBuildCmd currently mutates the
	// optimization options while processing builtins, potentially forcing O3.
	for _, test := range cases {
		if test.optimize {
			optimized := test
			optimized.name += "/O3"
			optimized.args = append(append([]string(nil), test.args...), "--O3")
			cases = append(cases, optimized)
		}
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			fixture := filepath.Join("testdata", strings.Split(test.name, "/")[0])
			work := t.TempDir()
			if err := os.CopyFS(work, os.DirFS(fixture)); err != nil {
				t.Fatal(err)
			}
			// The CLI currently links C-backed modules by relative builtins/*.ll
			// paths. Supply that runtime layout without depending on the checkout cwd.
			if err := os.CopyFS(filepath.Join(work, "builtins"), os.DirFS(filepath.Join(repo, "builtins"))); err != nil {
				t.Fatal(err)
			}
			temp := filepath.Join(work, "tmp")
			if err := os.Mkdir(temp, 0700); err != nil {
				t.Fatal(err)
			}
			env := append(os.Environ(), "TMPDIR="+temp, "NO_COLOR=1", "TERM=dumb")
			executable := filepath.Join(work, "program")
			args := []string{"build"}
			if len(test.sources) == 0 {
				args = append(args, "main.gl3")
			} else {
				args = append(args, test.sources...)
			}
			args = append(args, "-o", executable)
			args = append(args, test.args...)
			compiled := run(work, env, "", 30*time.Second, compiler, args...)
			diagnostic := expected(t, fixture, "compile_error.txt")
			checkerDiagnostic := expected(t, fixture, "checker_error.txt")
			if checkerDiagnostic != "" {
				if diagnostic != "" {
					t.Fatal("fixture must not specify both compile_error.txt and checker_error.txt")
				}
				diagnostic = checkerDiagnostic
				found := false
				for _, line := range strings.Split(compiled.stdout+compiled.stderr, "\n") {
					if _, message, ok := strings.Cut(line, "checker error:"); ok && strings.Contains(message, strings.TrimSpace(diagnostic)) {
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("missing checker error %q\nstdout: %s\nstderr: %s", diagnostic, compiled.stdout, compiled.stderr)
				}
				if strings.Contains(compiled.stdout+compiled.stderr, "compiler error:") || strings.Contains(compiled.stdout+compiled.stderr, "compiler panic") {
					t.Fatalf("checker failure reached the emitter\n%s%s", compiled.stdout, compiled.stderr)
				}
			}
			if diagnostic != "" {
				if compiled.timeout {
					t.Fatalf("compiler timed out\n%s%s", compiled.stdout, compiled.stderr)
				}
				exit, ok := compiled.err.(*exec.ExitError)
				if !ok || exit.ExitCode() <= 0 {
					t.Fatalf("expected compiler failure, got %v\n%s%s", compiled.err, compiled.stdout, compiled.stderr)
				}
				if !strings.Contains(compiled.stdout+compiled.stderr, strings.TrimSpace(diagnostic)) {
					t.Fatalf("missing compiler diagnostic %q\nstdout: %s\nstderr: %s", diagnostic, compiled.stdout, compiled.stderr)
				}
				if _, err := os.Stat(executable); !os.IsNotExist(err) {
					t.Fatalf("failed compilation left an executable (stat: %v)", err)
				}
				return
			}
			if compiled.err != nil {
				t.Fatalf("compilation failed: %v\nstdout: %s\nstderr: %s", compiled.err, compiled.stdout, compiled.stderr)
			}
			result := run(work, env, test.stdin, 5*time.Second, executable)
			if result.timeout {
				t.Fatalf("program timed out\n%s%s", result.stdout, result.stderr)
			}
			code := 0
			if result.err != nil {
				exit, ok := result.err.(*exec.ExitError)
				if !ok {
					t.Fatalf("could not run executable: %v", result.err)
				}
				code = exit.ExitCode()
			}
			if code != test.exit {
				t.Errorf("program exit: want %d, got %d\nstdout: %s\nstderr: %s", test.exit, code, result.stdout, result.stderr)
			}
			if want := expected(t, fixture, "stdout.txt"); result.stdout != want {
				t.Errorf("stdout: want %q, got %q", want, result.stdout)
			}
			if want := expected(t, fixture, "stderr.txt"); result.stderr != want {
				t.Errorf("stderr: want %q, got %q", want, result.stderr)
			}
		})
	}
}

type commandResult struct {
	stdout, stderr string
	err            error
	timeout        bool
}

func run(dir string, env []string, stdin string, timeout time.Duration, executable string, args ...string) commandResult {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Dir, cmd.Env = dir, env
	cmd.Stdin = strings.NewReader(stdin)
	// Bound waiting on inherited pipes if a subprocess outlives its parent.
	cmd.WaitDelay = time.Second
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	return commandResult{stdout.String(), stderr.String(), err, ctx.Err() != nil}
}

func expected(t *testing.T, fixture, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(fixture, name))
	if os.IsNotExist(err) {
		return ""
	}
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
