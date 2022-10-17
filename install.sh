#!/bin/bash

# Check if root
if [ "$EUID" -ne 0 ]
  then echo "Please run as root"
  exit
fi

# Get latest release from github
RELEASE=$(curl -s https://api.github.com/repos/hidracloud/hidra/releases/latest | grep "tag_name" | cut -d '"' -f 4)

# Remove v from tag
VERSION=${RELEASE:1}

# Detect if darwin or linux
if [[ "$OSTYPE" == "darwin"* ]]; then
    OS="darwin"
else
    OS="linux"
fi

# Detect architecture
ARCH=$(uname -m)

# If x86_64, use amd64
if [[ "$ARCH" == "x86_64" ]]; then
    ARCH="amd64"
fi

echo "Installing hidra version $VERSION for $OS-$ARCH"

# Download tar.gz
curl -L https://github.com/hidracloud/hidra/releases/download/${RELEASE}/hidra_${VERSION}_${OS}_${ARCH}.tar.gz -o /tmp/hidra.tar.gz &> /dev/null

# Extract tar.gz
mkdir -p /tmp/hidra-${VERSION}
tar -xzf /tmp/hidra.tar.gz -C /tmp/hidra-${VERSION}

# Move hidra to /usr/local/bin
mv /tmp/hidra-${VERSION}/hidra /usr/local/bin/hidra

# Remove temp files
rm -rf /tmp/hidra-${VERSION}
rm /tmp/hidra.tar.gz

# Make executable
chmod +x /usr/local/bin/hidra

echo "Installed on /usr/local/bin/hidra"

# If mac, code sign
if [[ "$OS" == "darwin" ]]; then
    codesign -s - /usr/local/bin/hidra
fi

/usr/local/bin/hidra version