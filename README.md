[![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/hidra)](https://artifacthub.io/packages/search?repo=hidra)
[![Go Report Card](https://goreportcard.com/badge/github.com/hidracloud/hidra)](https://goreportcard.com/report/github.com/hidracloud/hidra) [![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/5722/badge)](https://bestpractices.coreinfrastructure.org/projects/5722)
[![Bugs](https://sonarcloud.io/api/project_badges/measure?project=hidracloud_hidra&metric=bugs)](https://sonarcloud.io/summary/new_code?id=hidracloud_hidra)
[![Code Smells](https://sonarcloud.io/api/project_badges/measure?project=hidracloud_hidra&metric=code_smells)](https://sonarcloud.io/summary/new_code?id=hidracloud_hidra)
[![Lines of Code](https://sonarcloud.io/api/project_badges/measure?project=hidracloud_hidra&metric=ncloc)](https://sonarcloud.io/summary/new_code?id=hidracloud_hidra)
[![Maintainability Rating](https://sonarcloud.io/api/project_badges/measure?project=hidracloud_hidra&metric=sqale_rating)](https://sonarcloud.io/summary/new_code?id=hidracloud_hidra)
[![Reliability Rating](https://sonarcloud.io/api/project_badges/measure?project=hidracloud_hidra&metric=reliability_rating)](https://sonarcloud.io/summary/new_code?id=hidracloud_hidra)
[![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=hidracloud_hidra&metric=security_rating)](https://sonarcloud.io/summary/new_code?id=hidracloud_hidra)
[![Technical Debt](https://sonarcloud.io/api/project_badges/measure?project=hidracloud_hidra&metric=sqale_index)](https://sonarcloud.io/summary/new_code?id=hidracloud_hidra)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=hidracloud_hidra&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=hidracloud_hidra)

# Hidra

Hidra allows you to monitor the status of your services without headaches.

## Installation

### Precompiles binaries

Precompiled binaries for released versions are available in the release section on Github. Using the latest production release binary is the recommended way of installing Hidra. You can find latest release [here](https://github.com/hidracloud/hidra/releases/latest)

### Docker images

Docker images are available on [Github Container Registry](https://github.com/hidracloud/hidra/pkgs/container/hidra).

### Package repositories
If you want to install Hidra easily, please use the package repositories. 

```bash
# Debian/Ubuntu
curl https://repo.hidra.cloud/apt/gpg.key | sudo apt-key add -
echo "deb [trusted=yes] https://repo.hidra.cloud/apt /" | sudo tee /etc/apt/sources.list.d/hidra.list

# RedHat/CentOS
curl https://repo.hidra.cloud/rpm/gpg.key | sudo rpm --import -
echo "[hidra]" | sudo tee /etc/yum.repos.d/hidra.repo
echo "name=Hidra" | sudo tee -a /etc/yum.repos.d/hidra.repo
echo "baseurl=https://repo.hidra.cloud/rpm/" | sudo tee -a /etc/yum.repos.d/hidra.repo
echo "enabled=1" | sudo tee -a /etc/yum.repos.d/hidra.repo
echo "gpgcheck=1" | sudo tee -a /etc/yum.repos.d/hidra.repo
```



### Using install script

You can use the install script to install Hidra on your system. The script will download the latest release binary and install it in your system. You can find the script [here](https://raw.githubusercontent.com/hidracloud/hidra/main/install.sh).

```bash
sudo bash -c "$(curl -fsSL https://raw.githubusercontent.com/hidracloud/hidra/main/install.sh)"
```

### Build from source

To build Hidra from source code, you need:

- Go version 1.19 or greater.
- [Goreleaser](https://goreleaser.com)

To build Hidra, run the following command:

```bash
goreleaser release --snapshot --rm-dist
```

You can find the binaries in the `dist` folder.

## Usage

### Exporter

Hidra has support for exposing metrics to Prometheus. If you want to use Hidra in exporter mode, run:

```bash
hidra exporter /etc/hidra/exporter.yml
```

You can find an example of the configuration file [here](https://github.com/hidracloud/hidra/blob/main/configs/hidra/exporter.yml)

#### Grafana

You can find a Grafana dashboard [here](https://github.com/hidracloud/hidra/blob/main/configs/grafana)

### Test mode

Hidra has support for running in test mode. Test mode will allow you to run one time a set of samples, and check if the results are as expected. If you want to use Hidra in test mode, run:

```bash
hidra test sample1.yml sample2.yml ... samplen.yml
```

If you want to exit on error, just add the flag `--exit-on-error`.

### Sample examples

You can find some sample examples [here](https://github.com/hidracloud/hidra/blob/main/configs/hidra/samples/)

### Docker compose

You can find an example of a docker compose file [here](https://github.com/hidracloud/hidra/blob/main/docker-compose.yml)

## Samples

Samples are the way Hidra knows what to do. A sample is a YAML file that contains the information needed to run a test. You can find some sample examples [here](https://github.com/hidracloud/hidra/blob/main/configs/hidra/samples). You can also find a sample example below:

```yaml
# Description of the sample
description: "This is a sample to test the HTTP plugin"
# Tags is a key-value list of tags that will be added to the sample. You can add here whatever you want.
tags:
  tenant: "hidra"
# Interval is the time between each execution of the sample.
interval: "1m"
# Timeout is the time that Hidra will wait for the sample to finish.
timeout: "10s"
# Steps is a list of steps that will be executed in order.
steps:
    # Plugin is the name of the plugin that will be used to execute the step.
  - plugin: http
    # Action is the action that will be executed by the plugin.
    action: request
    # Parameters is a key-value list of parameters that will be passed to the plugin.
    parameters:
      url: https://google.com/
  - plugin: http
    action: statusCodeShouldBe
    parameters:
      statusCode: 301
```

You can find more information about plugins in next section.

## Plugins

- [browser](https://github.com/hidracloud/hidra/blob/main/docs/plugins/browser/README.md)
- [dns](https://github.com/hidracloud/hidra/blob/main/docs/plugins/dns/README.md)
- [ftp](https://github.com/hidracloud/hidra/blob/main/docs/plugins/ftp/README.md)
- [http](https://github.com/hidracloud/hidra/blob/main/docs/plugins/http/README.md)
- [icmp](https://github.com/hidracloud/hidra/blob/main/docs/plugins/icmp/README.md)
- [tcp](https://github.com/hidracloud/hidra/blob/main/docs/plugins/tcp/README.md)
- [tls](https://github.com/hidracloud/hidra/blob/main/docs/plugins/tls/README.md)
- [udp](https://github.com/hidracloud/hidra/blob/main/docs/plugins/udp/README.md)
- [string](https://github.com/hidracloud/hidra/blob/main/docs/plugins/string/README.md)
