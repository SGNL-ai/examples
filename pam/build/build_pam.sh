#!/bin/bash

function get_distro() {
    if [[ -f /etc/os-release ]]
    then
        # On Linux systems
        source /etc/os-release
        echo $ID
    else
        # On systems other than Linux (e.g. Mac or FreeBSD)
        uname
    fi
}

get_distro

case $ID in
        raspbian)
        echo "The SGNL PAM module is not supported in Raspian"
        ;;
        rhel)
        echo "Installing SGNL PAM module on Red Hat Enterprise Linux..."
        gcc -Wall -fPIC -fno-stack-protector -c ../src/sgnl_pam.c
        sudo ld -x --shared -o /lib64/security/sgnl_pam.so ../src/sgnl_pam.o -lcurl -ljson-c
        rm ../src/sgnl_pam.o
        ;;
        ubuntu)
        echo "Installing SGNL PAM module on Ubuntu..."
        gcc -Wall -fPIC -fno-stack-protector -c ../src/sgnl_pam.c -o ../src/sgnl_pam.o
        sudo ld -x --shared -o /lib/x86_64-linux-gnu/security/sgnl_pam.so ../src/sgnl_pam.o -lcurl -ljson-c
        rm ../src/sgnl_pam.o
        sudo mkdir /etc/sgnl
        sudo cp ../config/sgnl_pam.json /etc/sgnl
        echo "Update configuration for sgnl_pam.json in /etc/sgnl"
        ;;
        Darwin)
        echo "The SGNL PAM module is not supported in macOS"
        ;;
esac
