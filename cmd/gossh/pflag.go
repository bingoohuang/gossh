package main

import (
	"fmt"
	"io/ioutil"

	"github.com/bingoohuang/gou/cnf"
	"github.com/bingoohuang/toml"
	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// LoadByPflag load values to cfgValue from pflag cnf specified path.
func LoadByPflag(cmdPrefixTag string, cfgValues ...interface{}) {
	f, _ := homedir.Expand(viper.GetString("cnf"))
	Load(cmdPrefixTag, f, cfgValues...)
}

// Load loads the cnfFile content and viper bindings to value.
func Load(cmdPrefixTag, cnfFile string, values ...interface{}) {
	if err := LoadE(cmdPrefixTag, cnfFile, values...); err != nil {
		logrus.Panicf("Load Cnf %s error %v", cnfFile, err)
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

	bs, err := ioutil.ReadFile(f)
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
