package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/atotto/clipboard"
	"github.com/bingoohuang/gossh"
	"github.com/bingoohuang/gou/pbe"
	"github.com/howeyc/gopass"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// PbePwd defines the keyword for client flag.
const PbePwd = "pbepwd"

// DeclarePbePflags declares the pbe required pflags.
func DeclarePbePflags() {
	pflag.StringP(PbePwd, "", "", "pbe password")
	pflag.StringP("pbe", "", "", "PrintEncrypt by pbe, string or @file")
	pflag.StringP("ebp", "", "", "PrintDecrypt by pbe, string or @file")
	pflag.StringP("pbechg", "", "", "file to be change with another pbes")
	pflag.StringP("pbepwdnew", "", "", "new pbe pwd")
}

// DealPbePflag deals the request by the pflags.
func DealPbePflag() bool {
	pbes := viper.GetString("pbe")
	ebps := viper.GetString("ebp")
	pbechg := viper.GetString("pbechg")

	if len(pbes) == 0 && len(ebps) == 0 && pbechg == "" {
		return false
	}

	alreadyHasOutput := false

	gossh.DecryptPassphrase("")
	passStr := pbe.GetPbePwd()

	if len(pbes) > 0 {
		pbe.PrintEncrypt(passStr, pbes)
		if val, err := pbe.Pbe(pbes); err == nil {
			if err := clipboard.WriteAll(val); err == nil {
				fmt.Printf("Copied to clipboard\n")
			}
		}
		alreadyHasOutput = true
	}

	if len(ebps) > 0 {
		if alreadyHasOutput {
			fmt.Println()
		}

		pbe.PrintDecrypt(passStr, ebps)

		if val, err := pbe.Ebp(ebps); err == nil {
			if err := clipboard.WriteAll(val); err == nil {
				fmt.Printf("Copied to clipboard\n")
			}
		}

		alreadyHasOutput = true
	}

	if pbechg != "" {
		if alreadyHasOutput {
			fmt.Println()
		}

		processPbeChgFile(pbechg, passStr, viper.GetString("pbepwdnew"))
	}

	return true
}

var (
	pbePwdOnce sync.Once // nolint
	pbePwd     string    // nolint
)

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
		os.Exit(1) // nolint gomnd
	}

	pbePwd = string(pass)
}

func processPbeChgFile(filename, passStr, pbenew string) {
	if f, err := homedir.Expand(filename); err == nil {
		filename = f
	}

	file, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	text, err := pbe.Config{Passphrase: passStr}.ChangePbe(string(file), pbenew)
	if err != nil {
		panic(err)
	}

	ft, _ := os.Stat(filename)

	if err := os.WriteFile(filename, []byte(text), ft.Mode()); err != nil {
		panic(err)
	}
}
