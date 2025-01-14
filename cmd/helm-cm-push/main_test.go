// Modifications copyright (C) 2019 Alibaba Group Holding Limited / Yuning Xie (xyn1016@gmail.com)

package main

import (
	"crypto/rand"
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"k8s.io/helm/pkg/getter"
	"k8s.io/helm/pkg/helm/helmpath"
	"k8s.io/helm/pkg/repo"
	"k8s.io/helm/pkg/tlsutil"
)

var (
	testTarballPath    = "../../testdata/charts/helm2/mychart/charts/mariadb-5.11.3.tgz"
	testCertPath       = "../../testdata/tls/test_cert.crt"
	testKeyPath        = "../../testdata/tls/test_key.key"
	testCAPath         = "../../testdata/tls/ca.crt"
	testServerCAPath   = "../../testdata/tls/server_ca.crt"
	testServerCertPath = "../../testdata/tls/test_server.crt"
	testServerKeyPath  = "../../testdata/tls/test_server.key"
)

func TestPushCmd(t *testing.T) {
	statusCode := 201
	body := "{\"success\": true}"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		w.Write([]byte(body))
	}))
	defer ts.Close()

	// Create new Helm home w/ test repo
	tmp, err := ioutil.TempDir("", "helm-push-test")
	if err != nil {
		t.Error("unexpected error creating temp test dir", err)
	}
	defer os.RemoveAll(tmp)

	home := helmpath.Home(tmp)
	f := repo.NewRepoFile()

	entry := repo.Entry{}
	entry.Name = "helm-push-test"
	entry.URL = ts.URL

	_, err = repo.NewChartRepository(&entry, getter.All(v2settings))
	if err != nil {
		t.Error("unexpected error created test repository", err)
	}

	f.Update(&entry)
	os.MkdirAll(home.Repository(), 0777)
	f.WriteFile(home.RepositoryFile(), 0644)

	os.Setenv("HELM_HOME", home.String())
	os.Setenv("HELM_REPO_USERNAME", "myuser")
	os.Setenv("HELM_REPO_PASSWORD", "mypass")
	os.Setenv("HELM_REPO_CONTEXT_PATH", "/x/y/z")

	// Not enough args
	args := []string{}
	cmd := newPushCmd(args)
	err = cmd.RunE(cmd, args)
	if err == nil {
		t.Error("expecting error with missing args, instead got nil")
	}

	// Bad chart path
	args = []string{"/this/this/not/a/chart", "helm-push-test"}
	cmd = newPushCmd(args)
	err = cmd.RunE(cmd, args)
	if err == nil {
		t.Error("expecting error with bad chart path, instead got nil")
	}

	// Bad repo name
	args = []string{testTarballPath, "wkerjbnkwejrnkj"}
	cmd = newPushCmd(args)
	err = cmd.RunE(cmd, args)
	if err == nil {
		t.Error("expecting error with bad repo name, instead got nil")
	}

	// Happy path
	args = []string{testTarballPath, "helm-push-test"}
	cmd = newPushCmd(args)
	err = cmd.RunE(cmd, args)
	if err != nil {
		t.Error("unexpecting error uploading tarball", err)
	}

	// Happy path, by repo URL
	args = []string{testTarballPath, ts.URL}
	cmd = newPushCmd(args)
	err = cmd.RunE(cmd, args)
	if err != nil {
		t.Error("unexpecting error uploading tarball, using repo URL", err)
	}

	// Trigger 409
	statusCode = 409
	body = "{\"error\": \"package already exists\"}"
	args = []string{testTarballPath, "helm-push-test"}
	cmd = newPushCmd(args)
	err = cmd.RunE(cmd, args)
	if err == nil {
		t.Error("expecting error with 409, instead got nil")
	}

	// Unable to parse JSON response body
	statusCode = 500
	body = "qkewjrnvqejrnbvjern"
	args = []string{testTarballPath, "helm-push-test"}
	cmd = newPushCmd(args)
	err = cmd.RunE(cmd, args)
	if err == nil {
		t.Error("expecting error with bad response body, instead got nil")
	}

	// acr:// downloader
	statusCode = 200
	os.Setenv("HELM_REPO_USE_HTTP", "true")
	downloaderBaseURL := strings.Replace(ts.URL, "http://", Protocol, 1)

	// fails with no file path
	args = []string{"", "", "", downloaderBaseURL}
	cmd = newPushCmd(args)
	err = cmd.RunE(cmd, args)
	if err == nil {
		t.Error("expecting error with bad acr:// url, instead got nil")
	}

	// index.yaml
	args = []string{"", "", "", downloaderBaseURL + "/index.yaml"}
	cmd = newPushCmd(args)
	err = cmd.RunE(cmd, args)
	if err != nil {
		t.Error("unexpected error trying to download index.yaml", err)
	}

	// charts/mariadb-5.11.3.tgz
	args = []string{"", "", "", downloaderBaseURL + "/charts/mariadb-5.11.3.tgz"}
	cmd = newPushCmd(args)
	err = cmd.RunE(cmd, args)
	if err != nil {
		t.Error("unexpected error trying to download charts/mariadb-5.11.3.tgz", err)
	}
}

func TestPushCmdWithTlsEnabledServer(t *testing.T) {
	statusCode := 201
	body := "{\"success\": true}"
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		w.Write([]byte(body))
	}))
	cert, err := tls.LoadX509KeyPair(testCertPath, testKeyPath)
	if err != nil {
		t.Fatalf("failed to load certificate and key with error: %s", err.Error())
	}

	caCertPool, err := tlsutil.CertPoolFromFile(testServerCAPath)
	if err != nil {
		t.Fatalf("load server CA file failed with error: %s", err.Error())
	}

	ts.TLS = &tls.Config{
		ClientCAs:    caCertPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{cert},
		Rand:         rand.Reader,
	}
	ts.StartTLS()
	defer ts.Close()

	// Create new Helm home w/ test repo
	tmp, err := ioutil.TempDir("", "helm-push-test")
	if err != nil {
		t.Error("unexpected error creating temp test dir", err)
	}
	defer os.RemoveAll(tmp)

	home := helmpath.Home(tmp)
	f := repo.NewRepoFile()

	entry := repo.Entry{}
	entry.Name = "helm-push-test"
	entry.URL = ts.URL

	_, err = repo.NewChartRepository(&entry, getter.All(v2settings))
	if err != nil {
		t.Error("unexpected error created test repository", err)
	}

	f.Update(&entry)
	os.MkdirAll(home.Repository(), 0777)
	f.WriteFile(home.RepositoryFile(), 0644)

	os.Setenv("HELM_HOME", home.String())
	os.Setenv("HELM_REPO_USERNAME", "myuser")
	os.Setenv("HELM_REPO_PASSWORD", "mypass")
	os.Setenv("HELM_REPO_CONTEXT_PATH", "/x/y/z")

	//no certificate options
	args := []string{testTarballPath, "helm-push-test"}
	cmd := newPushCmd(args)
	err = cmd.RunE(cmd, args)
	if err == nil {
		t.Fatal("expected non nil error but got nil when run cmd without certificate option")
	}

	os.Setenv("HELM_REPO_CA_FILE", testCAPath)
	os.Setenv("HELM_REPO_CERT_FILE", testServerCertPath)
	os.Setenv("HELM_REPO_KEY_FILE", testServerKeyPath)
	os.Setenv("HELM_REPO_INSECURE", "true")

	err = cmd.RunE(cmd, args)
	if err != nil {
		t.Fatalf("unexpecting error uploading tarball: %s", err)
	}
}
