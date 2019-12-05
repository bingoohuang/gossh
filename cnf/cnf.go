package cnf

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bingoohuang/gossh/elf"
	"github.com/bingoohuang/strcase"

	"github.com/BurntSushi/toml"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/tkrajina/go-reflector/reflector"
)

// CheckUnknownPFlags checks the pflag and exiting.
func CheckUnknownPFlags() {
	if args := pflag.Args(); len(args) > 0 {
		fmt.Printf("Unknown args %s\n", strings.Join(args, " "))
		pflag.PrintDefaults()
		os.Exit(1)
	}
}

// DeclarePflags declares cnf pflags.
func DeclarePflags() {
	pflag.StringP("cnf", "c", "", "cnf file path")
}

// LoadByPflag load values to cfgValue from pflag cnf specified path.
func LoadByPflag(cfgValues ...interface{}) {
	f, _ := homedir.Expand(viper.GetString("cnf"))
	Load(f, cfgValues...)
}

// ParsePflags parse pflags and bind to viper
func ParsePflags(envPrefix string) error {
	pflag.Parse()

	CheckUnknownPFlags()

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
func LoadE(cnfFile string, values ...interface{}) error {
	file, err := FindFile(cnfFile)
	if err != nil {
		return fmt.Errorf("FindFile error %w", err)
	}

	for _, value := range values {
		if _, err = toml.DecodeFile(file, value); err != nil {
			return fmt.Errorf("DecodeFile error %w", err)
		}
	}

	return nil
}

// Load loads the cnfFile content and viper bindings to value.
func Load(cnfFile string, values ...interface{}) {
	if err := LoadE(cnfFile, values...); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			logrus.Warnf("Load Cnf %s error %v", cnfFile, err)
		}
	}

	ViperToStruct(values...)
}

// ViperToStruct read viper value to struct
func ViperToStruct(structVars ...interface{}) {
	separator := ","
	for _, structVar := range structVars {
		separator = GetSeparator(structVar, separator)

		for _, f := range reflector.New(structVar).Fields() {
			if !f.IsExported() {
				continue
			}

			name := strcase.ToCamelLower(f.Name())

			switch t, _ := f.Get(); t.(type) {
			case []string:
				if v := strings.TrimSpace(viper.GetString(name)); v != "" {
					setField(f, elf.SplitX(v, separator))
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
func DeclarePflagsByStruct(structVars ...interface{}) {
	for _, structVar := range structVars {
		for _, f := range reflector.New(structVar).Fields() {
			if !f.IsExported() {
				continue
			}

			name := strcase.ToCamelLower(f.Name())
			tag := elf.DecodeTag(elf.PickFirst(f.Tag("pflag")))
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
}
