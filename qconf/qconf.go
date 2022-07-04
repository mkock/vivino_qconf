package qconf

import (
	"errors"
	"fmt"

	"github.com/birkirb/unicreds"
)

// Project encapsulates the configuration of qconf.
type Project struct {
	SelectedConfig string
	activeConfig   Config
	Configs        map[string]Config
}

// Config represents a AWS configuration.
type Config struct {
	Region          string `toml:"region"`
	Role            string `toml:"role"`
	File            string `toml:"file"`
	Alias           string `toml:"alias"`
	TableName       string `toml:"table_name"`
	EncodingContext string `toml:"encoding_context"`
}

// validate validates the Project configuration and returns a non-nil error in case of problems.
func (p *Project) validate() error {
	if len(p.Configs) == 0 {
		return errors.New("no configurations loaded, did you prepare a TOML configuration file?")
	}
	for key, conf := range p.Configs {
		if conf.Region == "" {
			return fmt.Errorf("%s: missing aws region", key)
		}
		if conf.Role == "" {
			return fmt.Errorf("%s: missing aws role", key)
		}
		if conf.EncodingContext == "" {
			return fmt.Errorf("%s: missing aws encoding context", key)
		}
		if conf.TableName == "" {
			return fmt.Errorf("%s: missing aws table name", key)
		}
		if conf.File == "" {
			return fmt.Errorf("%s: missing aws file", key)
		}
		if conf.Alias == "" {
			return fmt.Errorf("%s: missing aws alias", key)
		}
	}
	if p.SelectedConfig == "" {
		return errors.New("no configuration selected")
	}
	return nil
}

// Init must be called before any other method on Project. Init returns an error if initialization failed.
func (p *Project) Init() error {
	if err := p.validate(); err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	conf, ok := p.Configs[p.SelectedConfig]
	if !ok {
		return errors.New("selected config not found; check names in configuration file")
	}
	if err := unicreds.SetAwsConfig(&conf.Region, nil, &conf.Role); err != nil {
		return err
	}
	p.activeConfig = conf
	return nil
}

// Get fetches the requested credentials. Get returns an error if the operation failed.
func (p *Project) Get() (string, error) {
	encContext := unicreds.NewEncryptionContextValue()
	encContext.Set(p.activeConfig.EncodingContext)
	cred, err := unicreds.GetHighestVersionSecret(&p.activeConfig.TableName, p.activeConfig.File, encContext)
	if err != nil {
		return "", fmt.Errorf("GetHighestVersionSecret: %w", err)
	}
	if cred == nil {
		return "", fmt.Errorf("GetHighestVersionSecret: %w", errors.New("empty credentials"))
	}
	return cred.Secret, nil
}

// Put uploads the given file contents to the current project file. Put returns an error if the operation failed.
func (p *Project) Put(contents string) error {
	encContext := unicreds.NewEncryptionContextValue()
	encContext.Set(p.activeConfig.EncodingContext)
	ver, err := unicreds.ResolveVersion(&p.activeConfig.TableName, p.activeConfig.File, 0)
	if err != nil {
		return fmt.Errorf("ResolveVersion: %w", err)
	}
	if ver == "" {
		return fmt.Errorf("ResolveVersion: %w", errors.New("unable to determine version"))
	}
	return unicreds.PutSecret(&p.activeConfig.TableName, p.activeConfig.Alias, p.activeConfig.File, contents, ver, encContext)
}

// Filename returns the name of the file that is being processed.
func (p *Project) Filename() string {
	return p.activeConfig.File
}

