package pbe

import (
	"fmt"
	"os"
	"strings"

	"github.com/howeyc/gopass"
	"github.com/jedib0t/go-pretty/table"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func Pflags() {
	pflag.StringP("password", "", "", "pbe password")
	pflag.StringSliceP("pbe", "", nil, "encrypt by pbe")
	pflag.StringSliceP("ebp", "", nil, "decrypt by pbe")
}

func DealPflag() {
	pbes := viper.GetStringSlice("pbe")
	ebps := viper.GetStringSlice("ebp")

	if len(pbes) == 0 && len(ebps) == 0 {
		return
	}

	alreadyHasOutput := false
	passStr := getPassword()

	if len(pbes) > 0 {
		pbeEncrypt(pbes, passStr)

		alreadyHasOutput = true
	}

	if len(ebps) > 0 {
		if alreadyHasOutput {
			fmt.Println()
		}

		pbeDecrypt(ebps, passStr)
	}

	os.Exit(0)
}

func getPassword() string {
	passStr := viper.GetString("password")
	if passStr != "" {
		return passStr
	}

	fmt.Printf("Password: ")

	pass, err := gopass.GetPasswdMasked()
	if err != nil {
		fmt.Fprintf(os.Stderr, "GetPasswd error %v", err)
		os.Exit(1)
	}

	return string(pass)
}

const iterations = 19
const pbePrefix = `{PBE}`

func pbeEncrypt(pbes []string, passStr string) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"#", "Plain", "Encrypted"})

	for i, p := range pbes {
		pbed, err := Encrypt(p, passStr, iterations)
		if err != nil {
			fmt.Fprintf(os.Stderr, "pbe.Encrypt error %v", err)
			os.Exit(1)
		}

		t.AppendRow(table.Row{i + 1, p, pbePrefix + pbed})
	}

	t.Render()
}

func pbeDecrypt(ebps []string, passStr string) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"#", "Encrypted", "Plain"})

	for i, ebp := range ebps {
		ebpx := strings.TrimPrefix(ebp, pbePrefix)

		p, err := Decrypt(ebpx, passStr, iterations)
		if err != nil {
			fmt.Fprintf(os.Stderr, "pbe.Decrypt error %v", err)
			os.Exit(1)
		}

		t.AppendRow(table.Row{i + 1, ebp, p})
	}

	t.Render()
}
