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
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"istio.io/pkg/log"
)

// poller is used to poll files from remote url at specific internal
type poller struct {
	// url is remote target url to poll from
	url string
	// minInterval is time intervals in minutes for polling
	minInterval time.Duration
	// existingHash records last sha value of polled files
	existingHash string
	// time ticker
	ticker *time.Ticker
}

const (
	// Charts filename to fetch
	InstallationChartsFileName = "istio-installer.tar.gz"
	// Sha filename to verify
	InstallationShaFileName = "istio-installer.tar.gz.sha256"
	// Temporary Files prefix
	ChartsTempFilePrefix = "charts-"
)

// checkUpdate checks whether sha file being updated or not
// fetch latest charts if yes
func (p *poller) checkUpdate(uf *URLFetcher) error {
	shaF, err := uf.fetchSha()
	if err != nil {
		return err
	}
	hashAll, err := ioutil.ReadFile(shaF)
	if err != nil {
		return fmt.Errorf("failed to read sha file: %s", err)
	}
	// Original sha file name is formatted with "HashValue filename"
	newHash := strings.Fields(string(hashAll))[0]
	// compare with existing hash
	// if sha file updated then fetch latest charts
	if !strings.EqualFold(newHash, p.existingHash) {
		p.existingHash = newHash
		err := uf.fetchChart(shaF)
		log.Errorf("Error fetching charts and verify: %v", err)
		return err
	}
	return nil
}

func (p *poller) poll(uf *URLFetcher) {
	for t := range p.ticker.C {
		// When the ticker fires
		fmt.Println("Tick at", t)
		err := p.checkUpdate(uf)
		log.Errorf("Error polling charts: %v", err)
	}
}

// Run the polling job with target directory at specific interval
func Run(dirURL string, interval time.Duration) error {
	po := &poller{
		url:         dirURL,
		minInterval: interval,
		ticker:      time.NewTicker(time.Minute * interval),
	}
	destDir, err := ioutil.TempDir("", ChartsTempFilePrefix)
	if err != nil {
		log.Fatal("failed to create temp directory for charts")
		return err
	}
	uf := &URLFetcher{
		url:        po.url + "/" + InstallationChartsFileName,
		verifyURL:  po.url + "/" + InstallationShaFileName,
		verify:     true,
		destDir:    destDir,
		downloader: NewFileDownloader(),
	}

	go po.poll(uf)

	os.RemoveAll(destDir)
	return nil
}
