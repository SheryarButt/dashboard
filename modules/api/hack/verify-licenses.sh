#!/usr/bin/env bash

# Copyright 2022 The Kubermatic Kubernetes Platform contributors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -euo pipefail

cd $(dirname $0)/../../..
source hack/lib.sh

API=modules/api

CONTAINERIZE_IMAGE=quay.io/kubermatic/build:go-1.24-node-20-6 containerize ./$API/hack/verify-licenses.sh

cd $API
go mod vendor

echodate "Checking licenses..."
wwhrd check -q
echodate "Check successful."
