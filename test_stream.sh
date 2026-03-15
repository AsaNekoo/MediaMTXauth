#!/bin/sh

if [ $# -lt 1 ]; then
  echo "Usage: $0 PATH"
  exit 1
fi

host="${MEDIAMTX_HOST:-localhost}"
port="${MEDIAMTX_RTMP_PORT:-1935}"
path=`echo "$1" | sed "s#^/*##"`
shift

exec ffmpeg -hide_banner -re \
  -f lavfi -i testsrc2=size=320x240:rate=30 \
  -f lavfi -i sine=frequency=1000:sample_rate=48000 \
  -c:v libx264 -preset veryfast -tune zerolatency -pix_fmt yuv420p -g 60 -c:a aac -b:a 128k \
  -f flv "rtmp://$host:$port/$path"
