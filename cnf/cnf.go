package cnf

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bingoohuang/gossh/elf"

	"github.com/BurntSushi/toml"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/tkrajina/go-reflector/reflector"
)

// DeclarePflags declares cnf pflags.
func DeclarePflags() {
	pflag.StringP("cnf", "", "", "cnf file path")
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

	return "", fmt.Errorf("unable to find cnf file %s", cnfFile)
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
		logrus.Warnf("Load Cnf %s error %v", cnfFile, err)
	}

	ViperToStruct(value)
}

// ViperToStruct read viper value to struct
func ViperToStruct(structVar interface{}) {
	for _, f := range reflector.New(structVar).Fields() {
		if !f.IsExported() {
			continue
		}

		switch t, _ := f.Get(); t.(type) {
		case []string:
			if v := viper.GetStringSlice(f.Name()); len(v) > 0 {
				setField(f, v)
			}
		case string:
			if v := strings.TrimSpace(viper.GetString(f.Name())); v != "" {
				setField(f, v)
			}
		case int:
			if v := viper.GetInt(f.Name()); v != 0 {
				setField(f, v)
			}
		case bool:
			if v := viper.GetBool(f.Name()); v {
				setField(f, v)
			}
		}
	}
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

		name := f.Name()

		switch t, _ := f.Get(); t.(type) {
		case []string:
			pflag.StringSliceP(name, "", nil, name)
		case string:
			pflag.StringP(name, "", "", name)
		case int:
			pflag.IntP(name, "", 0, name)
		case bool:
			pflag.BoolP(name, "", false, name)
		}
	}
}
