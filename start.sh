#!/usr/bin/bash

cd $(dirname $0)

sf=$1

###

f=conf/app.conf.yaml

while read k v
do
  sed -i "s|{"$k"}|"$v"|"  $f
done < $sf

/opt/app/cla-server
