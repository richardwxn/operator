package helm

import (
	"fmt"
	"github.com/jasonlvhit/gocron"
	"io/ioutil"
	"istio.io/pkg/log"
	"os"
	"strings"
)

type poller struct {
	// polling target
	url string
	// polling interval in minutes
	minInterval uint64
	// existing sha value
	existingHash string
}

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
	newHash := strings.Fields(string(hashAll))[0]
	// compare with existing hash
	// if sha file updated then fetch latest charts
	if !strings.EqualFold(newHash, p.existingHash) {
		p.existingHash = newHash
		err := uf.fetchChartAndVerify(shaF)
		log.Errorf("Error fetching charts and verify: %v", err)
		return err
	}
	return nil
}

// Run the polling job with target directory at specific interval
func Run(udir string, interval uint64) {
	po := &poller{
		url:         udir,
		minInterval: interval,
	}
	destDir, err := ioutil.TempDir("", "charts-")
	if err != nil {
		log.Fatal("failed to create temp directory for charts")
		return
	}
	uf := &URLFetcher{
		url:        po.url + "/istio-installer.tar.gz",
		verifyUrl:  po.url + "/istio-installer.tar.gz.sha256",
		untar:      true,
		untarDir:   "untar",
		verify:     true,
		destDir:    destDir,
		downloader: newFileDownloader(),
	}
	gocron.Every(interval).Minutes().Do(po.checkUpdate, uf)
	<-gocron.Start()

	defer os.RemoveAll(destDir)
}
