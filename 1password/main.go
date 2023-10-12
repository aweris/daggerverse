package main

import (
	"context"
	"fmt"
	"runtime"
	"strings"
)

type Onepassword struct {
	Token *Secret
}

func (m *Onepassword) Auth(_ context.Context, token string) *Onepassword {
	return &Onepassword{Token: dag.SetSecret("OP_SERVICE_ACCOUNT_TOKEN", token)}
}

func (m *Onepassword) Read(ref string) (string, error) {
	out, err := m.container().WithExec([]string{"op", "read", ref}).Stdout(context.Background())
	if err != nil {
		return "", err
	}

	out = strings.TrimSpace(out)

	return out, nil
}

func (c *Container) With1PasswordSecret(token, ref, name string) (*Container, error) {
	op := &Onepassword{Token: dag.SetSecret("OP_SERVICE_ACCOUNT_TOKEN", token)}

	val, err := op.Read(ref)
	if err != nil {
		return c, err
	}

	return c.WithSecretVariable(name, dag.SetSecret(name, val)), nil
}

func (m *Onepassword) container() *Container {
	downloadURL := fmt.Sprintf("https://cache.agilebits.com/dist/1P/op2/pkg/v2.20.0/op_linux_%s_v2.20.0.zip", runtime.GOARCH)

	return dag.Container().
		From("alpine:latest").
		WithExec([]string{"apk", "add", "zip"}).
		WithExec([]string{"wget", downloadURL, "-O", "op.zip"}).
		WithExec([]string{"unzip", "op.zip", "-d", "/usr/local/bin"}).
		WithExec([]string{"chmod", "+x", "/usr/local/bin/op"}).
		WithSecretVariable("OP_SERVICE_ACCOUNT_TOKEN", m.Token)
}
