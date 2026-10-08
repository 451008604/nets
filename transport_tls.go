package nets

import "crypto/tls"

// Load once before listening. Each admitted connection performs its own bounded
// handshake in Open, so a stalled client cannot hold up the accept loop.
func loadStreamTLSConfig(conf ServerConf) (*tls.Config, error) {
	if conf.TLSCertPath == "" && conf.TLSKeyPath == "" {
		return nil, nil
	}
	certificate, err := tls.LoadX509KeyPair(conf.TLSCertPath, conf.TLSKeyPath)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		Certificates: []tls.Certificate{certificate},
		MinVersion:   tls.VersionTLS12,
	}, nil
}
