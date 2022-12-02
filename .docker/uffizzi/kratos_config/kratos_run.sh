#!/bin/sh
export UFFIZZI_URL_WITH_ESCAPE_CHAR=$(echo $UFFIZZI_URL | sed "s/\//\\\\\//g")
sed -i "s/app/${UFFIZZI_URL_WITH_ESCAPE_CHAR}/g" /etc/config/kratos/kratos.yml

cd /usr/src/app;

kratos -c /etc/config/kratos/kratos.yml migrate sql -e --yes
kratos serve -c /etc/config/kratos/kratos.yml --dev --watch-courier