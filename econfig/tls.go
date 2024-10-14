package main

import (
	"crypto"
	"crypto/ecdsa"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"github.com/QQGoblin/go-sdk/pkg/pkiutil"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/keyutil"
	"math"
	"math/big"
	"net"
	"time"
)

var (
	caCertFile string
	caKeyFile  string
	certFile   string
	keyFile    string
	dnsList    []string
	ipList     []string
)

func LoadCACertificateAndKey(caPath string, keyPath string) (*x509.Certificate, crypto.Signer, error) {

	certs, err := certutil.CertsFromFile(caPath)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "couldn't load the certificate file %s", caPath)
	}

	// We are only putting one certificate in the certificate pem file, so it's safe to just pick the first one
	// TODO: Support multiple certs here in order to be able to rotate certs
	cert := certs[0]

	// Parse the private key from a file
	privKey, err := keyutil.PrivateKeyFromFile(keyPath)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "couldn't load the private key file %s", keyPath)
	}

	// Allow RSA and ECDSA formats only
	var key crypto.Signer
	switch k := privKey.(type) {
	case *rsa.PrivateKey:
		key = k
	case *ecdsa.PrivateKey:
		key = k
	default:
		return nil, nil, errors.Errorf("the private key file %s is neither in RSA nor ECDSA format", keyPath)
	}

	return cert, key, nil
}

func CreateServerCertAndKey(dnsList []string, addressList []net.IP, caCert *x509.Certificate, caKey crypto.Signer, certfile, keyfile string) error {

	// 获取节点 VIP 信息

	// 创建证书模板
	notBefore, _ := time.Parse("2006-01-02 15:04:05", "1970-01-01 00:00:00")
	notAfter, _ := time.Parse("2006-01-02 15:04:05", "2170-01-01 00:00:00")
	serial, _ := cryptorand.Int(cryptorand.Reader, new(big.Int).SetInt64(math.MaxInt64))

	serverCertTempl := &x509.Certificate{
		Subject: pkix.Name{
			CommonName: DefaultServerCertCommonName,
		},
		IPAddresses:           addressList,
		DNSNames:              dnsList,
		SerialNumber:          serial,
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
	}

	// 通过根证书生成 ServerCert 和 ServerKey
	cert, key, err := pkiutil.CreateSignedCertAndKey(serverCertTempl, caCert, caKey)
	if err != nil {
		return err
	}

	// 写入证书文件
	if writeErr := certutil.WriteCert(certfile, pkiutil.EncodeCertPEM(cert)); writeErr != nil {
		return writeErr
	}

	// 写入Key文件
	encoded, err := keyutil.MarshalPrivateKeyToPEM(key)
	if err != nil {
		return err
	}
	if err := keyutil.WriteKey(keyfile, encoded); err != nil {
		return err
	}

	return nil
}

func generateTLS(dnslist, iplist []string, caCertfile, caKeyfile, certfile, keyfile string) error {

	addresslist := []net.IP{
		net.ParseIP("127.0.0.1"),
	}

	dnsSet := sets.NewString(dnslist...)
	dnsSet.Insert("localhost")
	dnsSet.Insert("etcd")
	dnsSet.Insert("etcd.default")
	dnsSet.Insert("etcd.default.svc")
	dnsSet.Insert("etcd.default.svc.cluster")
	dnsSet.Insert("etcd.default.svc.cluster.local")

	for _, ip := range iplist {
		addresslist = append(addresslist, net.ParseIP(ip))
	}

	var (
		err    error
		caCert *x509.Certificate
		caKey  crypto.Signer
	)
	if caKeyfile == "" || caCertfile == "" {
		caCert, caKey, err = pkiutil.LoadCertificateAndKeyFromString(pkiutil.DefaultCACert, pkiutil.DefaultCAKey)
		if err != nil {
			return err
		}
	} else {
		caCert, caKey, err = LoadCACertificateAndKey(caCertfile, caKeyfile)
		if err != nil {
			return err
		}
	}

	return CreateServerCertAndKey(dnsSet.List(), addresslist, caCert, caKey, certfile, keyfile)
}
