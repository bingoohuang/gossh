package main

import (
	"fmt"
	"os"

	"github.com/bingoohuang/gossh/pkg/cnf"
	"github.com/bingoohuang/ngg/ss"
	"github.com/bingoohuang/toml"
	"github.com/spf13/viper"
)

// LoadByPflag load values to cfgValue from pflag cnf specified path.
func LoadByPflag(cmdPrefixTag string, cfgValues ...interface{}) {
	f := ss.ExpandHome(viper.GetString("cnf"))
	Load(cmdPrefixTag, f, cfgValues...)
}

// Load loads the cnfFile content and viper bindings to value.
func Load(cmdPrefixTag, cnfFile string, values ...interface{}) {
	if cnfFile != "" {
		if err := LoadE(cmdPrefixTag, cnfFile, values...); err != nil {
			log.Printf("P! Load Cnf %s error %v", cnfFile, err)
		}
	}
	cnf.ViperToStruct(values...)
}

// LoadE similar to Load.
func LoadE(cmdPrefixTag, cnfFile string, values ...interface{}) error {
	f, err := cnf.FindFile(cnfFile)
	if err != nil {
		return fmt.Errorf("FindFile error %w", err)
	}

	prefixMap := toml.WithPrefixMap(map[string]string{"cmds": cmdPrefixTag})

	bs, err := os.ReadFile(f)
	if err != nil {
		return err
	}

	for _, value := range values {
		if _, err = toml.Decode(string(bs), value, prefixMap); err != nil {
			return fmt.Errorf("DecodeFile error %w", err)
		}
	}

	return nil
}
