#! /bin/sh
mkdir /var/log/$1
chown $(id -g):$(id -u) /var/log/$1
chmod 744 /var/log/$1