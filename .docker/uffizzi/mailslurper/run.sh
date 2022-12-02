#!/bin/sh

# Enable SSL in ServiceSettings.go
sed -i 's/\/\/ IsSSL/IsSSL/g' /go/src/github.com/mailslurper/mailslurper/pkg/mailslurper/ServiceSettings.go
# Enable SSL in ServiceSettings.go
sed -i 's/\/\/ IsSSL/IsSSL/g' /go/src/github.com/mailslurper/mailslurper/cmd/mailslurper/controllers/AdminController.go

sed -i 's/4437,/443,\n\  "IsSSL\": true,/g' /go/src/github.com/mailslurper/mailslurper/cmd/mailslurper/config.json
sed -i 's/":" + serviceSettings\.servicePort/""/g' /go/src/github.com/mailslurper/mailslurper/cmd/mailslurper/www/mailslurper/js/services/SettingsService.js

go get
go generate
go build

mailslurper