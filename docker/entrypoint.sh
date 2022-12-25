#! /bin/sh
cd /home/echoserver/
./create_directories.sh echoserver
port=${1:-12345}
echo "use port $port"
echoserver $port
