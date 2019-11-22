package pbe

import (
	"fmt"
	"os"
	"sync"

	"github.com/howeyc/gopass"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// PbePwd defines the keyword for client flag.
const PbePwd = "pbepwd"

// DeclarePflags declares the pbe required pflags.
func DeclarePflags() {
	pflag.StringP(PbePwd, "", "", "pbe password")
	pflag.StringSliceP("pbe", "", nil, "PrintEncrypt by pbe")
	pflag.StringSliceP("ebp", "", nil, "PrintDecrypt by pbe")
}

// DealPflag deals the request by the pflags.
func DealPflag() bool {
	pbes := viper.GetStringSlice("pbe")
	ebps := viper.GetStringSlice("ebp")

	if len(pbes) == 0 && len(ebps) == 0 {
		return false
	}

	alreadyHasOutput := false
	passStr := GetPbePwd()

	if len(pbes) > 0 {
		PrintEncrypt(passStr, pbes...)

		alreadyHasOutput = true
	}

	if len(ebps) > 0 {
		if alreadyHasOutput {
			fmt.Println()
		}

		PrintDecrypt(passStr, ebps...)
	}

	return true
}

var pbePwdOnce sync.Once // nolint
var pbePwd string        // nolint

// GetPbePwd read pbe password from viper, or from stdin.
func GetPbePwd() string {
	pbePwdOnce.Do(readInternal)

	return pbePwd
}

func readInternal() {
	pbePwd = viper.GetString(PbePwd)
	if pbePwd != "" {
		return
	}

	fmt.Printf("PBE Password: ")

	pass, err := gopass.GetPasswdMasked()
	if err != nil {
		fmt.Fprintf(os.Stderr, "GetPasswd error %v", err)
		os.Exit(1)
	}

	pbePwd = string(pass)
}
