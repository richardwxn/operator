// Copyright 2019 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package helm

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/mholt/archiver"

	"istio.io/istio/pkg/log"
)

// FileDownloader is wrapper of HTTP client to download files
type FileDownloader struct {
	// client is a HTTP/HTTPS client.
	client *http.Client
}

// URLFetcher is used to fetch and manipulate charts from remote url
type URLFetcher struct {
	// url is url to download the charts
	url string
	// verifyUrl is url to download the verification file
	verifyUrl string
	// verify indicates whether the downloaded tar should be verified
	verify bool
	// destDir is path of charts downloaded to, empty as default to temp dir
	destDir string
	// untarDir is destination of untar
	untarDir string
	// downloader
	downloader *FileDownloader
}

// fetchChart fetches the charts and verifies charts against shaF if required
func (f *URLFetcher) fetchChart(shaF string) error {
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
	return archiver.Unarchive(saved, f.untarDir)
}

// fetch sha file
func (f *URLFetcher) fetchSha() (string, error) {
	shaF, err := f.downloader.DownloadTo(f.verifyUrl, f.destDir)
	if err != nil {
		return "", err
	}
	return shaF, nil
}

func newFileDownloader() *FileDownloader {
	return &FileDownloader{
		client: &http.Client{
			Transport: &http.Transport{
				DisableCompression: true,
				Proxy:              http.ProxyFromEnvironment,
			}},
	}
}

// ExpandFile expands the src file into the dest directory.
func ExtractFile(dest, src string) error {
	h, err := os.Open(src)
	if err != nil {
		return err
	}
	defer h.Close()
	return
}

// Send GET Request
func (c *FileDownloader) Get(href string) (*bytes.Buffer, error) {
	buf := bytes.NewBuffer(nil)

	req, err := http.NewRequest("GET", href, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to fetch URL %s : %s", href, resp.Status)
	}

	_, err = io.Copy(buf, resp.Body)
	return buf, err
}

// Ref is remote url to download from, dest is local file path to download to
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
