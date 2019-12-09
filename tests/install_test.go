package tests

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/release"
	"io"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"log"
	"os"
	"testing"
)

// 验证helm install foo ./foo-0.1.0.tgz命令
func TestInstall(t *testing.T) {
	settings.Debug = true

	namespace := "comon"
	apiServer := "https://192.168.84.111:52222"
	token := "10ced45d9ef9246deb9cf4ebac7bb3d5"
	releaseName := "foo"
	chartPath := "./foo-0.1.0.tgz"

	// or settings.KubeConfig = "./config"

	config := &genericclioptions.ConfigFlags{
		APIServer:   stringptr(apiServer),
		BearerToken: stringptr(token),
		Insecure:    boolptr(true),
		Namespace:   stringptr(namespace),
	}

	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(
		config,
		namespace,
		"memory",
		debug); err != nil {
		debug("%+v", err)
		os.Exit(1)
	}

	args := make([]string, 0)
	args = append(args, releaseName)
	args = append(args, chartPath)

	client := action.NewInstall(actionConfig)
	client.Namespace = namespace

	valueOpts := &values.Options{}

	rel, err := runInstall(args, client, valueOpts, os.Stdout)
	if err != nil {
		fmt.Println(err)
		return
	}

	b, _ := json.Marshal(rel)
	fmt.Println(string(b))
}

func stringptr(val string) *string {
	return &val
}

func boolptr(val bool) *bool {
	return &val
}

type statusPrinter struct {
	release *release.Release
	debug   bool
}

func runInstall(args []string, client *action.Install, valueOpts *values.Options, out io.Writer) (*release.Release, error) {
	debug("Original chart version: %q", client.Version)
	if client.Version == "" && client.Devel {
		debug("setting version to >0.0.0-0")
		client.Version = ">0.0.0-0"
	}

	name, chart, err := client.NameAndChart(args)
	if err != nil {
		return nil, err
	}
	client.ReleaseName = name

	cp, err := client.ChartPathOptions.LocateChart(chart, settings)
	if err != nil {
		return nil, err
	}

	debug("CHART PATH: %s\n", cp)

	p := getter.All(settings)
	vals, err := valueOpts.MergeValues(p)
	if err != nil {
		return nil, err
	}

	// Check chart dependencies to make sure all are present in /charts
	chartRequested, err := loader.Load(cp)
	if err != nil {
		return nil, err
	}

	validInstallableChart, err := isChartInstallable(chartRequested)
	if !validInstallableChart {
		return nil, err
	}

	if req := chartRequested.Metadata.Dependencies; req != nil {
		// If CheckDependencies returns an error, we have unfulfilled dependencies.
		// As of Helm 2.4.0, this is treated as a stopping condition:
		// https://github.com/helm/helm/issues/2209
		if err := action.CheckDependencies(chartRequested, req); err != nil {
			if client.DependencyUpdate {
				man := &downloader.Manager{
					Out:              out,
					ChartPath:        cp,
					Keyring:          client.ChartPathOptions.Keyring,
					SkipUpdate:       false,
					Getters:          p,
					RepositoryConfig: settings.RepositoryConfig,
					RepositoryCache:  settings.RepositoryCache,
				}
				if err := man.Update(); err != nil {
					return nil, err
				}
			} else {
				return nil, err
			}
		}
	}

	client.Namespace = settings.Namespace()
	return client.Run(chartRequested, vals)
}

// isChartInstallable validates if a chart can be installed
//
// Application chart type is only installable
func isChartInstallable(ch *chart.Chart) (bool, error) {
	switch ch.Metadata.Type {
	case "", "application":
		return true, nil
	}
	return false, errors.Errorf("%s charts are not installable", ch.Metadata.Type)
}

func debug(format string, v ...interface{}) {
	if settings.Debug {
		format = fmt.Sprintf("[debug] %s\n", format)
		log.Output(2, fmt.Sprintf(format, v...))
	}
}
