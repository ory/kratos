package daemon

import (

	// #nosec G505 - This is required for certificate chains alongside sha256
	"crypto/tls"
	"crypto/x509"
	"path"
	"runtime"
	"sync"
	"time"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"

	"github.com/fsnotify/fsnotify"
)

type certificatesProvider struct {
	certs              []tls.Certificate
	mu                 sync.Mutex
	certLocation       *config.CertLocation
	d                  driver.Registry
	watcher            *fsnotify.Watcher
	getTLSCertificates func() ([]tls.Certificate, *config.CertLocation, error)
}

func newCertificatesProvider(certs []tls.Certificate, certLocation *config.CertLocation, d driver.Registry, getTLSCertificates func() ([]tls.Certificate, *config.CertLocation, error)) *certificatesProvider {
	ret := &certificatesProvider{
		certLocation:       certLocation,
		d:                  d,
		getTLSCertificates: getTLSCertificates,
	}
	ret.load(certs)
	if certLocation != nil {
		ret.watchCertificatesChanges()
	}

	runtime.SetFinalizer(ret, func(ret *certificatesProvider) { ret.stop() })

	return ret
}

func (p *certificatesProvider) load(certs []tls.Certificate) {
	for i := range certs {
		tlsCert := &certs[i]
		if tlsCert.Leaf != nil {
			continue
		}
		for _, bCert := range tlsCert.Certificate {
			cert, _ := x509.ParseCertificate(bCert)
			if !cert.IsCA {
				tlsCert.Leaf = cert
			}
		}
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	p.certs = certs
}

func (p *certificatesProvider) watchCertificatesChanges() {
	var err error
	p.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		p.d.Logger().WithError(err).Fatalf("Could not activate certificate change watcher")
	}

	go func() {
		p.d.Logger().Infof("Starting tls certificate auto-refresh")
		for {
			select {
			case _, ok := <-p.watcher.Events:
				if !ok {
					return
				}

				p.waitForAllFilesChanges()

				p.d.Logger().Infof("TLS certificates changed, updating")
				certs, _, err := p.getTLSCertificates()
				if err != nil {
					p.d.Logger().WithError(err).Fatalf("Error in the new tls certificates")
					return
				}
				p.load(certs)
			case err, ok := <-p.watcher.Errors:
				if !ok {
					return
				}
				p.d.Logger().WithError(err).Fatalf("Error occured in the tls certificate change watcher")
			}
		}
	}()

	certPath := path.Dir(p.certLocation.CertPath)
	keyPath := path.Dir(p.certLocation.KeyPath)

	err = p.watcher.Add(certPath)
	if err != nil {
		p.d.Logger().WithError(err).Fatalf("Error watching the certFolder for tls certificate change")
	}

	if certPath != keyPath {
		err = p.watcher.Add(keyPath)
		if err != nil {
			p.d.Logger().WithError(err).Fatalf("Error watching the keyFolder for tls certificate change")
		}
	}
}

func (p *certificatesProvider) waitForAllFilesChanges() {
	flushUntil := time.After(2 * time.Second)
	p.d.Logger().Infof("TLS certificates files changed, waiting for changes to finish")
	stop := false
	for {
		select {
		case <-flushUntil:
			stop = true
		case <-p.watcher.Events:
			continue
		}

		if stop {
			break
		}
	}
}

func (p *certificatesProvider) getCertificate(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if hello != nil {
		for _, cert := range p.certs {
			if cert.Leaf != nil && cert.Leaf.VerifyHostname(hello.ServerName) == nil {
				return &cert, nil
			}
		}
	}
	return &p.certs[0], nil
}

func (p *certificatesProvider) stop() {
	if p.watcher != nil {
		p.watcher.Close()
		p.watcher = nil
	}
}
