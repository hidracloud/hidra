# tls
TLS plugin is used to check TLS certificates
## Available actions
### connectTo
Connects to a TLS server
#### Parameters
- to: Host to connect to
### dnsShouldBePresent
Checks if a DNS record should be present
#### Parameters
- dns: DNS name to check
### shouldBeValidFor
Checks if a certificate is valid for a given host
#### Parameters
- for: Duration for which the certificate should be valid
### onClose
Close the connection
#### Parameters
