#!/bin/bash

set -euxo pipefail

cd "$( dirname "${BASH_SOURCE[0]}" )/../.."

killall kratos || true