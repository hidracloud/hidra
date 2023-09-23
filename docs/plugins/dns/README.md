# dns
DNS plugin is used to check DNS information
## Available actions
### dig
Ask NS about the domain
#### Parameters
- ns: The NS to ask
- type: The type of the query
- host: The host to ask
### whoisFrom
Gets the whois information from a domain
#### Parameters
- domain: The domain to get the whois information
-  (optional) dateFormat: Date format to parse, default is 2006-01-02T15:04:05.999Z
### shouldBeValidFor
Checks if the domain is valid for a given number of duration
#### Parameters
- for: The duration to check if the domain is valid
-  (optional) dateFormat: Date format to parse, default is 2006-01-02T15:04:05.999Z
### dnsSecShouldBeValid
Checks if the domain has DNSSEC enabled
#### Parameters
- domain: The domain to check if DNSSEC is enabled
