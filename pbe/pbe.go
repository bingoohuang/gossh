package pbe

import (
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/table"
)

const iterations = 19
const pbePrefix = `{PBE}`

func PbeEncrypt(p string) (string, error) {
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

func PbeDecrypt(p string) (string, error) {
	pwd := GetPbePwd()
	if pwd == "" {
		return "", fmt.Errorf("pbepwd is requird")
	}

	if strings.HasPrefix(p, pbePrefix) {
		return Decrypt(p[len(pbePrefix):], pwd, iterations)
	}

	return p, nil
}

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
