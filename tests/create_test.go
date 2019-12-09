package tests

import (
	"fmt"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// 验证helm create foo命令
func TestCreate(t *testing.T) {
	o := &createOptions{
		name: "foo",
	}
	fmt.Println(o.create(os.Stdout))
}

type createOptions struct {
	starter    string // --starter
	name       string
	starterDir string
}

func (o *createOptions) create(out io.Writer) error {
	fmt.Fprintf(out, "Creating %s\n", o.name)

	chartname := filepath.Base(o.name)
	cfile := &chart.Metadata{
		Name:        chartname,
		Description: "A Helm chart for Kubernetes",
		Type:        "application",
		Version:     "0.1.0",
		AppVersion:  "0.1.0",
		APIVersion:  chart.APIVersionV2,
	}

	if o.starter != "" {
		// Create from the starter
		lstarter := filepath.Join(o.starterDir, o.starter)
		// If path is absolute, we don't want to prefix it with helm starters folder
		if filepath.IsAbs(o.starter) {
			lstarter = o.starter
		}
		return chartutil.CreateFrom(cfile, filepath.Dir(o.name), lstarter)
	}

	_, err := chartutil.Create(chartname, filepath.Dir(o.name))
	return err
}
