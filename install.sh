#!/bin/sh

set -e

os=$(uname -s)
arch=$(uname -m)
base_url="https://github.com/SekyrOrg/forge/releases/latest/download"

case "$os" in
  Darwin*) os="Darwin" ;;
  Linux*) os="Linux" ;;
  CYGWIN*|MINGW*|MSYS*) os="Windows" ;;
  *)
    echo "Unsupported OS: $os"
    exit 1
    ;;
esac

case "$arch" in
  x86_64) arch="x86_64" ;;
  arm64) arch="arm64" ;;
  i386) arch="i386" ;;
  *)
    echo "Unsupported architecture: $arch"
    exit 1
    ;;
esac

binary_name="forge_${os}_${arch}"
archive_ext=".tar.gz"
[ "$os" = "Windows" ] && archive_ext=".zip"

download_url="${base_url}/${binary_name}${archive_ext}"

echo "Downloading binary for ${os}-${arch}..."
curl -fsSL -o "${binary_name}${archive_ext}" "${download_url}"

if [ "$os" = "Windows" ]; then
  echo "Extracting ${binary_name}${archive_ext}..."
  unzip "${binary_name}${archive_ext}"
else
  echo "Extracting ${binary_name}${archive_ext}..."
  tar -xzf "${binary_name}${archive_ext}"
fi

echo "Installation complete!"