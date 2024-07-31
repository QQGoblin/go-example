CAROOT=$(pwd) mkcert -cert-file=tls.crt  -key-file=tls.key echo-request 127.0.0.1
CAROOT=$(pwd) mkcert -client -cert-file=client.crt  -key-file=client.key client

kubectl create secret generic default-gateway-tls --dry-run=client -o yaml \
--from-file=tls.crt=client.crt \
--from-file=tls.key=client.key