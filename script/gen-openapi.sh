#!/bin/bash
API_USER=./pkg/api/user

# Generate static assets for OpenAPI UI
# rm -rf $API_USER/statik
statik -m -f -src $API_USER/third_party/OpenAPI/ --dest $API_USER