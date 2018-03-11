#!/usr/bin/with-contenv sh

echo "Using Architect: ${ARCHITECT_ENDPOINT}"
sed -i -e "s#__ARCHITECT_ADDRESS__#${ARCHITECT_ENDPOINT}#g" /usr/share/nginx/html/static/js/*.js
