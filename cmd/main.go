package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/BurntSushi/toml"
	"github.com/mkock/vivino_quickconf/qconf"
)

var (
	confName, project string
	push, pull        string
	edit              bool
)

func init() {
	flag.StringVar(&confName, "conf", "qconf.toml", "name of TOML config file to load, relative to binary's directory")
	flag.StringVar(&project, "project", "", "name of project to edit (use names from config file)")
	flag.StringVar(&push, "push", "", "name of file to push config file to")
	flag.StringVar(&pull, "pull", "", "name of file to pull config file into")
	flag.BoolVar(&edit, "edit", false, "edit file in your $EDITOR and push it once done editing; if this flag is not set, the configuration file will be piped to stdout")
}

func main() {
	flag.Parse()

	if project == "" {
		printAndExit(errors.New(`argument "project" is required`))
	}

	rawConf, err := readFile(confName)
	if err != nil {
		printAndExit(fmt.Errorf("config: %w", err))
	}

	ed := os.Getenv("EDITOR")

	p := qconf.Project{
		SelectedConfig: project,
	}

	_, err = toml.Decode(rawConf, &p.Configs)
	if err != nil {
		printAndExit(fmt.Errorf("decode config: %w", err))
	}
	if err := p.Init(); err != nil {
		printAndExit(fmt.Errorf("init: %w", err))
	}
	switch {
	// Pull, edit in favorite editor and then push.
	case edit:
		origContents, err := p.Get()
		if err != nil {
			printAndExit(err)
		}
		if edit {
			tmp, err := os.CreateTemp("", project+"_")
			if err != nil {
				printAndExit(fmt.Errorf("edit: %w", err))
			}
			n, err := tmp.Write([]byte(origContents))
			if err != nil {
				printAndExit(fmt.Errorf("edit: %w", err))
			}
			fmt.Printf("Wrote %d bytes to %s\n", n, tmp.Name())
			tmp.Close()
			// Make a copy before editing.
			err = os.WriteFile(tmp.Name()+".backup", []byte(origContents), 0666)
			if err != nil {
				printAndExit(fmt.Errorf("edit: %w", err))
			}
			fmt.Println("Backup created at", tmp.Name()+".backup")
			cmd := exec.Command(ed, tmp.Name())
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			fmt.Println("Waiting for editor to exit...")
			err = cmd.Run()
			if err != nil {
				_ = os.Remove(tmp.Name())
				printAndExit(fmt.Errorf("edit: %w", err))
			}
			newContents, err := readFile(tmp.Name())
			if err != nil {
				printAndExit(fmt.Errorf("edit: %w", err))
			}
			// Check if contents has changed.
			if newContents == origContents {
				fmt.Println("File contents unchanged, aborting!")
				os.Exit(0)
			}
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Upload [Y/n]? ")
			ans, err := reader.ReadString('\n')
			if err != nil {
				printAndExit(fmt.Errorf("edit: %w", err))
			}
			if ans != "Y\n" && ans != "y\n" {
				fmt.Println("Aborted!")
				os.Exit(0)
			}
			// Upload.
			err = p.Put(newContents)
			if err != nil {
				printAndExit(fmt.Errorf("edit: %w", err))
			}
			_ = os.Remove(tmp.Name())
			fmt.Println("Successfully uploaded", p.Filename())
		}

	// Push local file without deleting it.
	case push != "":
		contents, err := readFile(push)
		if err != nil {
			printAndExit(fmt.Errorf("push: %w", err))
		}
		if contents == "" {
			printAndExit(fmt.Errorf("push: %w", errors.New("file is empty")))
		}
		err = p.Put(contents)
		if err != nil {
			printAndExit(fmt.Errorf("put: %w", err))
		}
		fmt.Println("Successfully uploaded", p.Filename())

	// Pull into local file without editing or pushing it.
	case pull != "":
		_, err := os.Stat(pull)
		if !errors.Is(err, os.ErrNotExist) {
			printAndExit(fmt.Errorf("pull: %w", errors.New("file already exists")))
		}
		if contents, err := p.Get(); err != nil {
			printAndExit(fmt.Errorf("pull: %w", err))
		} else {
			err = os.WriteFile(pull, []byte(contents), 0666)
			if err != nil {
				printAndExit(fmt.Errorf("pull: %w", err))
			}
			fmt.Println("Wrote", pull)
		}

		// Pull contents and pipe to stdout.
	default:
		if contents, err := p.Get(); err != nil {
			printAndExit(fmt.Errorf("pipe: %w", err))
		} else {
			fmt.Fprint(os.Stdout, contents)
		}
	}
	os.Exit(0)
}

func readFile(fname string) (string, error) {
	f, err := os.Open(fname)
	if err != nil {
		return "", err
	}
	defer f.Close()
	bytes, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func printAndExit(err error) {
	fmt.Fprintln(os.Stderr, "Error:", err)
	os.Exit(1)
}
