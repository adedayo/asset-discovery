package discover

import (
	"crypto/x509"
	"net"
)

//Config asset discovery config
type Config struct {
}

//Asset is a digtal asset
type Asset struct {
	DomainName   string
	Certificates []*x509.Certificate
	MXRecords    []*net.MX //MX records, if any
}
