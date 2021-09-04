package dom

import (
	"bufio"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type tld map[string]*tld

var (
	tlds = &tld{}
)

const (
	tldFile = "public_suffix_list.dat"
	tldURL  = "https://publicsuffix.org/list/" + tldFile
)

func init() {
	loadTLDs()
}

func loadTLDs() {
	//if TLD file doesn't exist, download
	if _, err := os.Stat(tldFile); os.IsNotExist(err) {
		err = downloadTLDS()
		if err != nil {
			log.Printf("%v", err)
			return
		}
	}

	readTLDFile()

}

func readTLDFile() {
	file, err := os.Open(tldFile)
	if err != nil {
		log.Printf("%v", err)
		return
	}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		if line := scanner.Text(); len(line) != 0 && !strings.HasPrefix(line, `//`) {
			// log.Printf("Line: %s", line)
			domains := strings.Split(line, ".")
			currentTLD := *tlds
			for i := len(domains) - 1; i >= 0; i-- {
				d := domains[i]
				if nextTLD, present := currentTLD[d]; !present {
					nextTLD = &tld{}
					currentTLD[d] = nextTLD
					currentTLD = *nextTLD
				} else {
					currentTLD = *nextTLD
				}
			}
		}
	}

}

//TODO: need a mechanism for getting updates after the first download
func downloadTLDS() error {
	resp, err := http.Get(tldURL)

	if err != nil {
		log.Printf("%v", err)
		return err
	}
	defer resp.Body.Close()

	file, err := os.Create(tldFile)

	if err != nil {
		log.Printf("%v", err)
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, resp.Body)

	if err != nil {
		log.Printf("%v", err)
		return err
	}
	return nil
}
