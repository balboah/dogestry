// +build integration

package cli

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/dogestry/dogestry/config"
	"github.com/dogestry/dogestry/remote"
	docker "github.com/fsouza/go-dockerclient"
)

const fixture = "busybox:latest"

func init() {
	cmd := exec.Command("docker", "pull", fixture)
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

func TestExportImageToFiles(t *testing.T) {
	cfg, err := config.NewConfig(false, 22375, false, false, false)
	if err != nil {
		t.Fatal(err)
	}
	cli, err := NewDogestryCli(cfg, hosts, testTmpDirRoot)
	if err != nil {
		t.Fatal(err)
	}
	tmpDir, err := cli.CreateAndReturnTempDir()
	if err != nil {
		t.Fatal(err)
	}

	if err := cli.exportToFiles(fixture, &remoteMock{}, tmpDir); err != nil {
		t.Error(err)
	}
	filepath.Walk(filepath.Join(tmpDir, "images"), func(path string, _ os.FileInfo, err error) error {
		if err != nil {
			t.Fatal(err)
		}
		t.Log(path)
		return nil
	})
}

type remoteMock struct{}

func (m *remoteMock) Push(image, imageRoot string) error {
	return nil
}
func (m *remoteMock) PullImageId(id remote.ID, imageRoot string) error {
	return nil
}
func (m *remoteMock) ParseTag(repo, tag string) (remote.ID, error) {
	return remote.ID(""), nil
}
func (m *remoteMock) ResolveImageNameToId(image string) (remote.ID, error) {
	return remote.ID(""), nil
}
func (m *remoteMock) ImageFullId(id remote.ID) (remote.ID, error) {
	return remote.ID(""), nil
}
func (m *remoteMock) ImageMetadata(id remote.ID) (docker.Image, error) {
	return docker.Image{}, errors.New("mocked remote has no image metadata")
}
func (m *remoteMock) ParseImagePath(path string, prefix string) (repo, tag string) {
	return "", ""
}
func (m *remoteMock) WalkImages(id remote.ID, walker remote.ImageWalkFn) error {
	return nil
}
func (m *remoteMock) Validate() error {
	return nil
}
func (m *remoteMock) Desc() string {
	return ""
}
func (m *remoteMock) List() ([]remote.Image, error) {
	return []remote.Image{}, nil
}
