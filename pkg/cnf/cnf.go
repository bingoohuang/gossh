package cnf

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/bingoohuang/gg/pkg/filex"

	"github.com/tkrajina/go-reflector/reflector"

	"github.com/bingoohuang/gg/pkg/strcase"
	"github.com/bingoohuang/gou/str"

	"github.com/bingoohuang/toml"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// CheckUnknownPFlags checks the pflag and exiting.
func CheckUnknownPFlags() {
	if args := pflag.Args(); len(args) > 0 {
		fmt.Printf("Unknown args %s\n", strings.Join(args, " "))
		pflag.PrintDefaults()
		os.Exit(1) // nolint gomnd
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
	if filex.Exists(cnfFile) {
		return cnfFile, nil
	}

	if wd, _ := os.Getwd(); wd != "" {
		if cnf := filepath.Join(wd, "cnf.toml"); filex.Exists(cnf) {
			return cnf, nil
		}
	}

	if ex, err := os.Executable(); err == nil {
		if cnf := filepath.Join(filepath.Dir(ex), "cnf.toml"); filex.Exists(cnf) {
			return cnf, nil
		}
	}

	return "", fmt.Errorf("unable to find cnf file %s, error %w", cnfFile, os.ErrNotExist)
}

// LoadE similar to Load.
func LoadE(cnfFile string, values ...interface{}) error {
	f, err := FindFile(cnfFile)
	if err != nil {
		return fmt.Errorf("FindFile error %w", err)
	}

	for _, value := range values {
		if _, err = toml.DecodeFile(f, value); err != nil {
			return fmt.Errorf("DecodeFile error %w", err)
		}
	}

	return nil
}

// Load loads the cnfFile content and viper bindings to value.
func Load(cnfFile string, values ...interface{}) {
	if err := LoadE(cnfFile, values...); err != nil {
		logrus.Panicf("Load Cnf %s error %v", cnfFile, err)
	}

	ViperToStruct(values...)
}

// ViperToStruct read viper value to struct
func ViperToStruct(structVars ...interface{}) {
	separator := ","
	for _, structVar := range structVars {
		separator = GetSeparator(structVar, separator)

		viperObjectFields(structVar, separator)
	}
}

func viperObjectFields(pObj interface{}, separator string) {
	pValue := reflect.ValueOf(pObj)

	viperObjectFieldsValue(pValue, separator)
}

func viperObjectFieldsValue(pValue reflect.Value, separator string) {
	if pValue.Kind() != reflect.Ptr {
		return
	}

	objVV := pValue.Elem()
	if objVV.Kind() != reflect.Struct {
		return
	}

	objVT := objVV.Type()

	for i := 0; i < objVV.NumField(); i++ {
		ft := objVT.Field(i)
		fv := objVV.Field(i)

		if ft.PkgPath != "" { // not exportable
			continue
		}

		if ft.Anonymous || ft.Type.Kind() == reflect.Struct {
			viperObjectFieldsValue(fv.Addr(), separator)

			continue
		}

		setField(strcase.ToCamelLower(ft.Name), fv, separator)
	}
}

func setField(name string, fv reflect.Value, separator string) {
	switch ft := fv.Type(); ft.Kind() {
	case reflect.Slice:
		values := viper.GetStringSlice(name)
		if len(values) == 1 {
			if v := strings.TrimSpace(viper.GetString(name)); v != "" {
				fv.Set(reflect.ValueOf(str.SplitX(v, separator)))
			}
		} else if len(values) > 1 {
			fv.Set(reflect.ValueOf(values))
		}
	case reflect.String:
		if v := strings.TrimSpace(viper.GetString(name)); v != "" {
			fv.SetString(v)
		}
	case reflect.Int:
		if v := viper.GetInt(name); v != 0 {
			fv.SetInt(int64(v))
		}
	case reflect.Bool:
		if v := viper.GetBool(name); v {
			fv.SetBool(v)
		}
	default:
		if v := viper.GetString(name); v != "" {
			if ft == reflect.TypeOf(time.Second) {
				fv.Set(reflect.ValueOf(str.ParseDuration(v)))
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

// DeclarePflagsByStruct declares flags from struct fields' name and type
func DeclarePflagsByStruct(structVars ...interface{}) {
	for _, structVar := range structVars {
		for _, f := range reflector.New(structVar).Fields() {
			if !f.IsExported() {
				continue
			}

			if f.IsAnonymous() || f.Kind() == reflect.Struct {
				fv, _ := f.Get()
				DeclarePflagsByStruct(fv)

				continue
			}

			name := strcase.ToCamelLower(f.Name())
			tag := str.DecodeTag(str.PickFirst(f.Tag("pflag")))
			shorthand := tag.GetOpt("shorthand")
			defaultValue := tag.GetOpt("default")

			switch t, _ := f.Get(); t.(type) {
			case int:
				pflag.IntP(name, shorthand, str.ParseInt(defaultValue), tag.Main)
			case bool:
				pflag.BoolP(name, shorthand, str.ParseBool(defaultValue), tag.Main)
			case []string:
				var values []string
				if defaultValue != "" {
					values = append(values, defaultValue)
				}
				pflag.StringArrayP(name, shorthand, values, tag.Main)
			default: // string, []string, time.Interval:
				pflag.StringP(name, shorthand, defaultValue, tag.Main)
			}
		}
	}
}
