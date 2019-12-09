package tests

import (
	"errors"
	"fmt"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"os"
	"testing"
)

var (
	settings = cli.New()
)

// 验证helm package foo -d charts命令
func TestPack(t *testing.T) {
	fmt.Println(pack("foo", "charts/"))
}

func pack(path, destination string) error {
	client := action.NewPackage()
	client.Destination = destination
	if client.Sign {
		if client.Key == "" {
			return errors.New("--key is required for signing a package")
		}
		if client.Keyring == "" {
			return errors.New("--keyring is required for signing a package")
		}
	}
	client.RepositoryConfig = settings.RepositoryConfig
	client.RepositoryCache = settings.RepositoryCache

	p, err := client.Run(path, make(map[string]interface{}))
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "Successfully packaged chart and saved it to: %s\n", p)
	return nil
}
