#!/usr/bin/env bash
set -euo pipefail

REPO="Ben8t/grimmoir"
BINARY_NAME="grim"

detect_os() {
  case "$(uname -s)" in
    Linux) echo "linux" ;;
    Darwin) echo "darwin" ;;
    *)
      echo "Unsupported OS: $(uname -s)" >&2
      exit 1
      ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64) echo "amd64" ;;
    arm64|aarch64) echo "arm64" ;;
    *)
      echo "Unsupported architecture: $(uname -m)" >&2
      exit 1
      ;;
  esac
}

resolve_version() {
  if [[ -n "${VERSION:-}" ]]; then
    echo "${VERSION}"
    return
  fi

  curl -fsSLI -o /dev/null -w '%{url_effective}' "https://github.com/${REPO}/releases/latest" | sed 's#.*/tag/##'
}

choose_install_dir() {
  if [[ -n "${INSTALL_DIR:-}" ]]; then
    echo "${INSTALL_DIR}"
    return
  fi

  if [[ -w "/usr/local/bin" ]]; then
    echo "/usr/local/bin"
    return
  fi

  echo "${HOME}/.local/bin"
}

main() {
  local os arch version install_dir asset url tmpdir
  os="$(detect_os)"
  arch="$(detect_arch)"
  version="$(resolve_version)"
  install_dir="$(choose_install_dir)"

  asset="${BINARY_NAME}-${os}-${arch}.tar.gz"
  url="https://github.com/${REPO}/releases/download/${version}/${asset}"

  tmpdir="$(mktemp -d)"
  trap 'rm -rf "${tmpdir}"' EXIT

  mkdir -p "${install_dir}"

  echo "Downloading ${url}"
  curl -fsSL "${url}" -o "${tmpdir}/${asset}"

  tar -xzf "${tmpdir}/${asset}" -C "${tmpdir}"
  install -m 755 "${tmpdir}/${BINARY_NAME}" "${install_dir}/${BINARY_NAME}"

  echo "Installed ${BINARY_NAME} to ${install_dir}/${BINARY_NAME}"
  if [[ ":$PATH:" != *":${install_dir}:"* ]]; then
    echo "Warning: ${install_dir} is not in PATH"
  fi
  echo "Run: ${BINARY_NAME} --help"
}

main "$@"
