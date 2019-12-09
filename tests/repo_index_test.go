package tests

import (
	"fmt"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/repo"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// 验证helm repo index charts命令
func TestRepoIndex(t *testing.T) {
	o := &repoIndexOptions{
		dir: "charts",
	}
	fmt.Println(o.index(os.Stdout))
}

type repoIndexOptions struct {
	dir   string
	url   string
	merge string
}

func (i *repoIndexOptions) index(out io.Writer) error {
	path, err := filepath.Abs(i.dir)
	if err != nil {
		return err
	}

	return index(path, i.url, i.merge)
}

func index(dir, url, mergeTo string) error {
	out := filepath.Join(dir, "index.yaml")

	i, err := repo.IndexDirectory(dir, url)
	if err != nil {
		return err
	}
	if mergeTo != "" {
		// if index.yaml is missing then create an empty one to merge into
		var i2 *repo.IndexFile
		if _, err := os.Stat(mergeTo); os.IsNotExist(err) {
			i2 = repo.NewIndexFile()
			i2.WriteFile(mergeTo, 0644)
		} else {
			i2, err = repo.LoadIndexFile(mergeTo)
			if err != nil {
				return errors.Wrap(err, "merge failed")
			}
		}
		i.Merge(i2)
	}
	i.SortEntries()
	return i.WriteFile(out, 0644)
}
