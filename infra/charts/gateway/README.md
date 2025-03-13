# Gateway

This Helm chart configures the API Gateway for the ea platform.

It provides a self-signed certificate that can be used to access services over HTTPS during development
in the minikube cluster.

## Steps to re-generate self-signed cert

```bash
# Generate Private Key and CSR:
(files)$ openssl req -newkey rsa:2048 -nodes -keyout tls.key -out tls.csr
# Create self-signed certificate
(files)$ openssl x509 -req -sha256 -days 365 -in tls.csr -signkey tls.key -out tls.crt
>>>
...
Country Name (2 letter code) [AU]:FR
State or Province Name (full name) [Some-State]:
Locality Name (eg, city) []:
Organization Name (eg, company) [Internet Widgits Pty Ltd]:EruLabs
Organizational Unit Name (eg, section) []:
Common Name (e.g. server FQDN or YOUR name) []:erulabs.local
Email Address []:

Please enter the following 'extra' attributes
to be sent with your certificate request
A challenge password []:
An optional company name []:
# Store Certificate in Kubernetes Secret:
(files)$ kubectl create secret tls my-certificate --cert=tls.crt --key=tls.key
# Port secret to Helm chart
(files)$ kubectl get secret my-certificate -o yaml > ../templates/secret.yaml
# Some editing of the secret required
```