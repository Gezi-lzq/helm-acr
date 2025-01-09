package helm

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
)

var testTarballPath = "../../testdata/charts/helm2/mychart/charts/mariadb-5.11.3.tgz"

func TestSetVersion(t *testing.T) {
	c, err := GetChartByName(testTarballPath)
	if err != nil {
		t.Error("unexpected error getting test tarball chart", err)
	}
	c.SetVersion("latest")
	if c.V2.Metadata.Version != "latest" {
		t.Errorf("expected chart version to be latest, instead got %s", c.V2.Metadata.Version)
	}
}

func TestGetChartByName(t *testing.T) {
	// Bad name
	_, err := GetChartByName("/non/existent/path/mariadb-5.11.3.tgz")
	if err == nil {
		t.Error("expected error getting chart with bad name, instead got nil")
	}

	// Valid name
	c, err := GetChartByName(testTarballPath)
	if err != nil {
		t.Error("unexpected error getting test tarball chart", err)
	}
	if c.V2.Metadata.Name != "mariadb" {
		t.Errorf("expexted chart name to be mariadb, instead got %s", c.V2.Metadata.Name)
	}
	if c.V2.Metadata.Version != "5.11.3" {
		t.Errorf("expexted chart version to be 5.11.3, instead got %s", c.V2.Metadata.Version)
	}
}

func TestCreateChartPackage(t *testing.T) {
	c, err := GetChartByName(testTarballPath)
	if err != nil {
		t.Error("unexpected error getting test tarball chart", err)
	}

	tmp, err := ioutil.TempDir("", "helm-push-test")
	if err != nil {
		t.Error("unexpected error creating temp test dir", err)
	}
	defer os.RemoveAll(tmp)

	chartPackagePath, err := CreateChartPackage(c, tmp)
	if err != nil {
		t.Error("unexpected error creating chart package", err)
	}

	expectedPath := path.Join(tmp, "mariadb-5.11.3.tgz")
	if chartPackagePath != expectedPath {
		t.Errorf("expected chart path to be %s, but was %s", expectedPath, chartPackagePath)
	}
}
