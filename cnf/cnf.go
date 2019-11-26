package cnf

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bingoohuang/gou/str"

	"github.com/bingoohuang/gossh/elf"
	"github.com/bingoohuang/strcase"

	"github.com/BurntSushi/toml"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/tkrajina/go-reflector/reflector"
)

// DeclarePflags declares cnf pflags.
func DeclarePflags() {
	pflag.StringP("cnf", "c", "", "cnf file path")
}

// LoadByPflag load values to cfgValue from pflag cnf specified path.
func LoadByPflag(cfgValue interface{}) {
	f, _ := homedir.Expand(viper.GetString("cnf"))
	Load(f, cfgValue)
}

// ParsePflags parse pflags and bind to viper
func ParsePflags(envPrefix string) error {
	pflag.Parse()

	if args := pflag.Args(); len(args) > 0 {
		fmt.Printf("Unknown args %s\n", strings.Join(args, " "))
		pflag.PrintDefaults()
		os.Exit(1)
	}

	if envPrefix != "" {
		viper.SetEnvPrefix(envPrefix)
		viper.AutomaticEnv()
	}

	return viper.BindPFlags(pflag.CommandLine)
}

// FindFile tries to find cnfFile from specified path, or current path cnf.toml, executable path cnf.toml.
func FindFile(cnfFile string) (string, error) {
	if elf.SingleFileExists(cnfFile) == nil {
		return cnfFile, nil
	}

	if wd, _ := os.Getwd(); wd != "" {
		if cnfFile := filepath.Join(wd, "cnf.toml"); elf.SingleFileExists(cnfFile) == nil {
			return cnfFile, nil
		}
	}

	if ex, err := os.Executable(); err == nil {
		if cnfFile := filepath.Join(filepath.Dir(ex), "cnf.toml"); elf.SingleFileExists(cnfFile) == nil {
			return cnfFile, nil
		}
	}

	return "", fmt.Errorf("unable to find cnf file %s, error %w", cnfFile, os.ErrNotExist)
}

// LoadE similar to Load.
func LoadE(cnfFile string, value interface{}) error {
	if file, err := FindFile(cnfFile); err != nil {
		return fmt.Errorf("FindFile error %w", err)
	} else if _, err = toml.DecodeFile(file, value); err != nil {
		return fmt.Errorf("DecodeFile error %w", err)
	}

	return nil
}

// Load loads the cnfFile content and viper bindings to value.
func Load(cnfFile string, value interface{}) {
	if err := LoadE(cnfFile, value); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			logrus.Warnf("Load Cnf %s error %v", cnfFile, err)
		}
	}

	ViperToStruct(value)
}

// ViperToStruct read viper value to struct
func ViperToStruct(structVar interface{}) {
	separator := GetSeparator(structVar, ",")

	for _, f := range reflector.New(structVar).Fields() {
		if !f.IsExported() {
			continue
		}

		name := strcase.ToCamelLower(f.Name())

		switch t, _ := f.Get(); t.(type) {
		case []string:
			if v := strings.TrimSpace(viper.GetString(name)); v != "" {
				setField(f, str.SplitN(v, separator, true, true))
			}
		case string:
			if v := strings.TrimSpace(viper.GetString(name)); v != "" {
				setField(f, v)
			}
		case int:
			if v := viper.GetInt(name); v != 0 {
				setField(f, v)
			}
		case bool:
			if v := viper.GetBool(name); v {
				setField(f, v)
			}
		}
	}
}

// Separator ...
type Separator interface {
	// GetSeparator get the separator
	GetSeparator() string
}

// GetSeparator get separator from
// 1. viper's separator
// 2. v which implements Separator interface
// 3. or default value
func GetSeparator(v interface{}, defaultSeparator string) string {
	if sep := viper.GetString("separator"); sep != "" {
		return sep
	}

	if sep, ok := v.(Separator); ok {
		if s := sep.GetSeparator(); s != "" {
			return s
		}
	}

	return defaultSeparator
}

func setField(f reflector.ObjField, value interface{}) {
	if err := f.Set(value); err != nil {
		logrus.Warnf("Fail to set %s to %v, error %v", f.Name(), value, err)
	}
}

// DeclarePflagsByStruct declares flags from struct fields'name and type
func DeclarePflagsByStruct(structVar interface{}) {
	for _, f := range reflector.New(structVar).Fields() {
		if !f.IsExported() {
			continue
		}

		name := strcase.ToCamelLower(f.Name())
		tag := elf.DecodeTag(elf.IgnoreError(f.Tag("pflag")))
		usage := tag.Main
		shorthand := tag.GetOpt("shorthand")

		switch t, _ := f.Get(); t.(type) {
		case []string:
			pflag.StringP(name, shorthand, "", usage)
		case string:
			pflag.StringP(name, shorthand, "", usage)
		case int:
			pflag.IntP(name, shorthand, 0, usage)
		case bool:
			pflag.BoolP(name, shorthand, false, usage)
		}
	}
}
