package dom

import (
	"strings"
)

func GetFirstNonTLD(domain string) (out string) {

	domains := strings.Split(domain, ".")

	currentTLD := *tlds
	for i := len(domains) - 1; i >= 0; i-- {
		d := domains[i]
		if t, exists := currentTLD[d]; exists {
			if t != nil {
				currentTLD = *t
			}
		} else if tt, present := currentTLD["*"]; present {
			if tt != nil {
				currentTLD = *tt
			}
		} else {
			return d
		}
	}
	return
}
