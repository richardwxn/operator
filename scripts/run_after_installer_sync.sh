#!/bin/bash

# Copyright 2019 Istio Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e

WD=$(dirname "$0")
WD=$(cd "$WD"; pwd)
ROOT=$(dirname "$WD")

cd "${ROOT}"

# sync charts with installer and update vfs files(for charts)
# Note: local change you have would be compiled in vfs, move unnecessary changes before.
make generate-vfs

# update golden test files
make update-goldens

# Update default profile if needed based on the output
"${WD}"/run_migrate_profile.sh

# Update values schema
make generate-values

# update vfs files again(for possible profiles update)
make generate-vfs
