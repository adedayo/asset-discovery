package discover

import (
	"context"
	"crypto/x509"
	"fmt"
	"net"
	"sort"
	"strings"
	"sync"
	"time"

	dom "github.com/adedayo/asset-discovery/pkg/domains"
	certharvest "github.com/adedayo/certharvest/pkg"
)

var (
	worklist     map[string]nothing
	worklistLock sync.RWMutex
	domains      map[string]nothing
	domainLock   sync.RWMutex
)

type nothing struct{}

//FindAssets finds digital assets starting from a root domain
func FindAssets(ctx context.Context, domain string, config Config) (assets []Asset, err error) {

	_, assetMap := discoverRelatedDomains(ctx, domain, config)

	for asset, certs := range assetMap {
		hostName := strings.TrimPrefix(asset, "https://")
		asset := Asset{
			DomainName:   hostName,
			Certificates: certs,
		}
		if mxs, err := net.LookupMX(hostName); err == nil {
			asset.MXRecords = mxs
		}
		assets = append(assets, asset)
	}

	// fmt.Printf("All Domains (%d) %v\n", len(assets), assets)
	return
}

func GetBrands(ctx context.Context, domain string, config Config) []string {
	doms, _ := discoverRelatedDomains(ctx, domain, config)
	out := []string{}
	mOut := make(map[string]nothing)
	for _, d := range doms {
		mOut[dom.GetFirstNonTLD(d)] = nothing{}
	}
	for v := range mOut {
		if v != "" {
			out = append(out, v)
		}
	}
	sort.Strings(out)
	return out
}

func discoverRelatedDomains(ctx context.Context, domain string, config Config) ([]string, map[string][]*x509.Certificate) {
	domain = clean(domain)
	initialise(domain)

	// client := &rdap.Client{}

	// if resp, err := client.QueryDomain(domain); err == nil {
	// 	fmt.Printf("Domain Query: %#v \n", resp.SecureDNS)
	// 	// return
	// }

	assetMap := make(map[string][]*x509.Certificate)

	for doms, more := nextDomains(); more; doms, more = nextDomains() {
		websites := []string{}
		for _, dom := range doms {
			// log.Printf("To process: %v\n", dom)

			websites = append(websites, fmt.Sprintf("https://%s", dom))
		}

		for cerr := range certharvest.GetServerCertificates(certharvest.Config{TimeOut: 10 * time.Second}, websites...) {
			certs := cerr.CertificateChain
			if cerr.Error == nil {
				assetMap[cerr.URL] = cerr.CertificateChain
				if len(certs) > 0 {
					cert := certs[0]
					// fmt.Printf("Cert: %#v, %#v\n", cert.DNSNames, cert.Subject.CommonName)
					dd := cleanAll(append(cert.DNSNames, cert.Subject.CommonName))
					addDomains(dd)
				}
			}
		}
	}

	return getDomains(), assetMap
}

func getDomains() (out []string) {

	domainLock.RLock()
	defer domainLock.RUnlock()
	for d := range domains {
		out = append(out, d)
	}

	return
}

func isProcessed(dom string) bool {
	domainLock.RLock()
	defer domainLock.RUnlock()
	_, present := domains[dom]
	return present
}

func addDomains(doms []string) {
	for _, d := range doms {
		if !isProcessed(d) {
			worklistLock.Lock()
			worklist[d] = nothing{}
			worklistLock.Unlock()
		}
	}
}

func initialise(initialDomain string) {
	worklistLock.Lock()
	worklist = map[string]nothing{initialDomain: {}}
	worklistLock.Unlock()

	domainLock.Lock()
	domains = make(map[string]nothing)
	domainLock.Unlock()
}

func clean(domain string) string {
	return strings.TrimPrefix(strings.TrimSpace(domain), "*.")
}

func cleanAll(domains []string) (out []string) {
	for _, d := range domains {
		out = append(out, clean(d))
	}
	return
}

func nextDomains() (nextDomains []string, more bool) {

	worklistLock.Lock()
	defer worklistLock.Unlock()

	for dom := range worklist {
		//remove doman as we are processing it now
		delete(worklist, dom)
		if isProcessed(dom) {
			//domain previously processed
			continue
		}

		nextDomains = append(nextDomains, dom)
	}

	// log.Printf("Next domains: %v\n", nextDomains)

	if len(nextDomains) > 0 {
		//we found an unprocessed domain,add to processed list and return
		more = true
		domainLock.Lock()
		for _, dom := range nextDomains {
			domains[dom] = nothing{}
		}
		domainLock.Unlock()
		return
	}
	//empty worklist
	return
}
