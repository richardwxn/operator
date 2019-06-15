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
	minInterval uint64
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

// Check whether sha file being updated or not
// Fetch latest charts if yes
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
	for {
		select {
		case <-p.ticker.C:
			// When the ticker fires
			err := p.checkUpdate(uf)
			log.Errorf("Error polling charts: %v", err)
		}
	}
}

// Run the polling job with target directory at specific interval
func Run(dirURL string, interval uint64) error {
	po := &poller{
		url:         dirURL,
		minInterval: interval,
	}
	destDir, err := ioutil.TempDir("", ChartsTempFilePrefix)
	if err != nil {
		log.Fatal("failed to create temp directory for charts")
		return err
	}
	uf := &URLFetcher{
		url:        po.url + "/" + InstallationChartsFileName,
		verifyUrl:  po.url + "/" + InstallationShaFileName,
		untarDir:   "untar",
		verify:     true,
		destDir:    destDir,
		downloader: newFileDownloader(),
	}

	go po.poll(uf)

	os.RemoveAll(destDir)
	return nil
}
