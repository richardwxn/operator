package helm

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"istio.io/istio/pkg/log"
	"k8s.io/helm/pkg/chartutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type FileDownloader struct {
	// HTTP/HTTPS Client
	client *http.Client
}

type URLFetcher struct {
	// url to download the charts
	url string
	// url to download the verification file
	verifyUrl string
	// whether to verify downloaded tar
	verify bool
	// whether need to untar
	untar bool
	// dir of charts downloaded to, empty as default to temp dir
	destDir string
	// untar destination
	untarDir string
	// downloader
	downloader *FileDownloader
}

// fetch charts and verify with sha file
func (f *URLFetcher) fetchChartAndVerify(shaF string) error {
	c := f.downloader

	// fetch to default temp dir
	saved, err := c.DownloadTo(f.url, f.destDir)
	if err != nil {
		return err
	}
	file, err := os.Open(saved)
	if err != nil {
		return err
	}
	defer file.Close()

	if f.verify {
		// verify with sha file
		hashAll, err := ioutil.ReadFile(shaF)
		if err != nil {
			return fmt.Errorf("failed to read sha file: %s", err)
		}
		hash := strings.Fields(string(hashAll))[0]
		// Verify downloaded tar
		h := sha256.New()
		if _, err := io.Copy(h, file); err != nil {
			log.Error(err.Error())
		}
		sum := h.Sum(nil)
		if !strings.EqualFold(hex.EncodeToString(sum), hash) {
			return errors.New("checksum does not match")
		}
	}

	// After verification, untar the chart into the requested directory.
	if f.untar {
		ud := f.untarDir
		if !filepath.IsAbs(ud) {
			ud = filepath.Join(f.destDir, ud)
		}
		if fi, err := os.Stat(ud); err != nil {
			if err := os.MkdirAll(ud, 0755); err != nil {
				return fmt.Errorf("failed to untar (mkdir): %s", err)
			}

		} else if !fi.IsDir() {
			return fmt.Errorf("failed to untar: %s is not a directory", ud)
		}

		return chartutil.ExpandFile(ud, saved)
	}
	return nil
}

// fetch sha file
func (f *URLFetcher) fetchSha() (string, error) {
	c := f.downloader
	// fetch to default temp dir
	shaF, err := c.DownloadTo(f.verifyUrl, f.destDir)
	if err != nil {
		return "", err
	}
	return shaF, nil
}

func newFileDownloader() *FileDownloader {
	tr := &http.Transport{
		DisableCompression: true,
		Proxy:              http.ProxyFromEnvironment,
	}
	return &FileDownloader{
		client: &http.Client{Transport: tr},
	}
}

// Send GET Request
func (c *FileDownloader) Get(href string) (*bytes.Buffer, error) {
	buf := bytes.NewBuffer(nil)

	req, err := http.NewRequest("GET", href, nil)
	if err != nil {
		return buf, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return buf, err
	}
	if resp.StatusCode != 200 {
		return buf, fmt.Errorf("failed to fetch URL %s : %s", href, resp.Status)
	}

	_, err = io.Copy(buf, resp.Body)
	resp.Body.Close()
	return buf, err
}

// Download from URL to dest
func (c *FileDownloader) DownloadTo(ref, dest string) (string, error) {
	u, err := url.Parse(ref)
	if err != nil {
		return "", fmt.Errorf("invalid chart URL: %s", ref)
	}
	data, err := c.Get(u.String())
	if err != nil {
		return "", err
	}

	name := filepath.Base(u.Path)
	destFile := filepath.Join(dest, name)
	if err := ioutil.WriteFile(destFile, data.Bytes(), 0644); err != nil {
		return destFile, err
	}

	return destFile, nil
}
