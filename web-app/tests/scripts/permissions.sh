#!/bin/bash

# This file is part of MinIO Console Server
# Copyright (c) 2022 MinIO, Inc.
# # This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU Affero General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
# # This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU Affero General Public License for more details.
# # You should have received a copy of the GNU Affero General Public License
# along with this program.  If not, see <http://www.gnu.org/licenses/>.

SCRIPT_DIR=$(dirname "$0")
export SCRIPT_DIR
source "${SCRIPT_DIR}/common.sh"

remove_users() {
  mc admin user remove minio bucketassignpolicy-"$TIMESTAMP"
  mc admin user remove minio bucketread-"$TIMESTAMP"
  mc admin user remove minio bucketwrite-"$TIMESTAMP"
  mc admin user remove minio dashboard-"$TIMESTAMP"
  mc admin user remove minio diagnostics-"$TIMESTAMP"
  mc admin user remove minio groups-"$TIMESTAMP"
  mc admin user remove minio heal-"$TIMESTAMP"
  mc admin user remove minio iampolicies-"$TIMESTAMP"
  mc admin user remove minio logs-"$TIMESTAMP"
  mc admin user remove minio notificationendpoints-"$TIMESTAMP"
  mc admin user remove minio settings-"$TIMESTAMP"
  mc admin user remove minio tiers-"$TIMESTAMP"
  mc admin user remove minio trace-"$TIMESTAMP"
  mc admin user remove minio users-"$TIMESTAMP"
  mc admin user remove minio watch-"$TIMESTAMP"
  mc admin user remove minio bucketwriteprefixonlypolicy-"$TIMESTAMP"
  mc admin user remove minio inspect-allowed-"$TIMESTAMP"
  mc admin user remove minio inspect-not-allowed-"$TIMESTAMP"
  mc admin user remove minio prefix-policy-ui-crash-"$TIMESTAMP"
  mc admin user remove minio delete-object-with-prefix-"$TIMESTAMP"
  mc admin user remove minio conditions-"$TIMESTAMP"
  mc admin user remove minio conditions-2-"$TIMESTAMP"
}

remove_policies() {
  mc admin policy remove minio bucketassignpolicy-"$TIMESTAMP"
  mc admin policy remove minio bucketread-"$TIMESTAMP"
  mc admin policy remove minio bucketwrite-"$TIMESTAMP"
  mc admin policy remove minio dashboard-"$TIMESTAMP"
  mc admin policy remove minio diagnostics-"$TIMESTAMP"
  mc admin policy remove minio groups-"$TIMESTAMP"
  mc admin policy remove minio heal-"$TIMESTAMP"
  mc admin policy remove minio iampolicies-"$TIMESTAMP"
  mc admin policy remove minio logs-"$TIMESTAMP"
  mc admin policy remove minio notificationendpoints-"$TIMESTAMP"
  mc admin policy remove minio settings-"$TIMESTAMP"
  mc admin policy remove minio tiers-"$TIMESTAMP"
  mc admin policy remove minio trace-"$TIMESTAMP"
  mc admin policy remove minio users-"$TIMESTAMP"
  mc admin policy remove minio watch-"$TIMESTAMP"
  mc admin policy remove minio bucketwriteprefixonlypolicy-"$TIMESTAMP"
  mc admin policy remove minio inspect-allowed-"$TIMESTAMP"
  mc admin policy remove minio inspect-not-allowed-"$TIMESTAMP"
  mc admin policy remove minio fix-prefix-policy-ui-crash-"$TIMESTAMP"
  mc admin policy remove minio delete-object-with-prefix-"$TIMESTAMP"
  mc admin policy remove conditions-policy-"$TIMESTAMP"
  mc admin policy remove conditions-policy-2-"$TIMESTAMP"
}

remove_buckets() {
  mc rm minio/testcafe/write/test.txt && mc rm minio/testcafe
  mc rm minio/test/test.txt && mc rm minio/test/digitalinsights/xref_cust_guid_actd-v1.txt && mc rm minio/test/digitalinsights/test.txt && mc rm minio/test
  mc rm minio/testcondition/test.txt && mc rm minio/testcondition/firstlevel/xref_cust_guid_actd-v1.txt && mc rm minio/testcondition/firstlevel/test.txt && mc rm minio/testcondition/firstlevel/secondlevel/test.txt && mc rm minio/testcondition/firstlevel/secondlevel/thirdlevel/test.txt && mc rm minio/testcondition
}

cleanup() {
  remove_users
  remove_policies
  remove_buckets
}

__init__() {
  TIMESTAMP=$(date "+%s")
  echo "$TIMESTAMP" > web-app/tests/constants/timestamp.txt
  export GOPATH=/tmp/gopath
  export PATH=${PATH}:${GOPATH}/bin

  go install github.com/minio/mc@latest

  add_alias

  create_policies
  create_users
  assign_policies
  create_buckets
}

main() {
  (yarn start &> /dev/null) & (./console server &> /dev/null) & (testcafe "chrome:headless" "$1" -q --skip-js-errors -c 3)
  cleanup
}

( __init__ "$@" && main "$@" )