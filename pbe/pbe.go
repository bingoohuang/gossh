package pbe

import (
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/table"
)

const iterations = 19
const pbePrefix = `{PBE}`

// Pbe encrypts p by PBEWithMD5AndDES with 19 iterations.
// it will prompt password if viper get none.
func Pbe(p string) (string, error) {
	pwd := GetPbePwd()
	if pwd == "" {
		return "", fmt.Errorf("pbepwd is requird")
	}

	encrypt, err := Encrypt(p, pwd, iterations)
	if err != nil {
		return "", err
	}

	return pbePrefix + encrypt, nil
}

// Ebp decrypts p by PBEWithMD5AndDES with 19 iterations.
func Ebp(p string) (string, error) {
	if !strings.HasPrefix(p, pbePrefix) {
		return p, nil
	}

	pwd := GetPbePwd()
	if pwd == "" {
		return "", fmt.Errorf("pbepwd is requird")
	}

	return Decrypt(p[len(pbePrefix):], pwd, iterations)
}

// PrintEncrypt prints the PBE encryption.
func PrintEncrypt(passStr string, plains ...string) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"#", "Plain", "Encrypted"})

	for i, p := range plains {
		pbed, err := Encrypt(p, passStr, iterations)
		if err != nil {
			fmt.Fprintf(os.Stderr, "pbe.Encrypt error %v", err)
			os.Exit(1)
		}

		t.AppendRow(table.Row{i + 1, p, pbePrefix + pbed})
	}

	t.Render()
}

// PrintDecrypt prints the PBE decryption.
func PrintDecrypt(passStr string, cipherText ...string) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"#", "Encrypted", "Plain"})

	for i, ebp := range cipherText {
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
