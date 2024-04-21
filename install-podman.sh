#!/bin/bash

set -e

# Function to detect the distribution and version
detect_distribution() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        DISTRO=$ID
        VERSION=$VERSION_ID
    elif type lsb_release >/dev/null 2>&1; then
        DISTRO=$(lsb_release -si | tr '[:upper:]' '[:lower:]')
        VERSION=$(lsb_release -sr)
    elif [ -f /etc/lsb-release ]; then
        . /etc/lsb-release
        DISTRO=${DISTRIB_ID,,}
        VERSION=${DISTRIB_RELEASE}
    else
        echo "Your distribution is not supported by this script."
        exit 1
    fi
}

# Function to install Podman
install_podman() {
    case $DISTRO in
        arch|manjaro)
            sudo pacman -S --noconfirm podman
            ;;
        alpine)
            sudo apk add podman
            ;;
        debian|ubuntu|linuxmint)
            if [[ $DISTRO == "ubuntu" && $(echo $VERSION | cut -d '.' -f1) -ge 20 ]] || [[ $DISTRO == "debian" && $(echo $VERSION | cut -d '.' -f1) -ge 11 ]]; then
                sudo apt-get update
                sudo apt-get install -y podman
            else
                echo "Podman is not available in the official repositories for this version of $DISTRO."
                exit 1
            fi
            ;;
        centos)
            if [[ $(echo $VERSION | cut -d '.' -f1) -eq 7 ]]; then
                sudo yum update -y
                sudo yum -y install podman
            fi
            ;;
        fedora)
            sudo dnf -y install podman
            ;;
        opensuse*|sles)
            sudo zypper install -y podman
            ;;
        *)
            echo "Your distribution ($DISTRO) is not supported by this script."
            exit 1
            ;;
    esac
}

# Detect the distribution
detect_distribution

# Install Podman
install_podman

echo "Podman installation complete."
