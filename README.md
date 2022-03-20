<p align="center">
  <img width="176.5" height="202" src="https://github.com/hidracloud/hidra/blob/main/docs/logo.svg?raw=true">
</p>

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

# hidra

Don't lose your mind monitoring your services. Hidra lends you its head.

# ICMP

If you want to use ICMP scenario, you should activate on your system:

  sudo sysctl -w net.ipv4.ping_group_range="0 2147483647"
