#!/bin/bash

# This file is part of MinIO Console Server
# Copyright (c) 2023 MinIO, Inc.
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

add_alias() {
    for i in $(seq 1 4); do
        echo "... attempting to add alias $i"
        until (mc alias set minio http://127.0.0.1:9000 minioadmin minioadmin); do
            echo "...waiting... for 5secs" && sleep 5
        done
    done
}

remove_users() {
  mc admin user remove minio bucketassignpolicy-$TIMESTAMP
  mc admin user remove minio bucketread-$TIMESTAMP
  mc admin user remove minio bucketwrite-$TIMESTAMP
  mc admin user remove minio bucketobjecttags-$TIMESTAMP
  mc admin user remove minio bucketcannottag-$TIMESTAMP
  mc admin user remove minio dashboard-$TIMESTAMP
  mc admin user remove minio diagnostics-$TIMESTAMP
  mc admin user remove minio groups-$TIMESTAMP
  mc admin user remove minio heal-$TIMESTAMP
  mc admin user remove minio iampolicies-$TIMESTAMP
  mc admin user remove minio logs-$TIMESTAMP
  mc admin user remove minio notificationendpoints-$TIMESTAMP
  mc admin user remove minio settings-$TIMESTAMP
  mc admin user remove minio tiers-$TIMESTAMP
  mc admin user remove minio trace-$TIMESTAMP
  mc admin user remove minio users-$TIMESTAMP
  mc admin user remove minio watch-$TIMESTAMP
  mc admin user remove minio inspect-allowed-$TIMESTAMP
  mc admin user remove minio inspect-not-allowed-$TIMESTAMP
  mc admin user remove minio prefix-policy-ui-crash-$TIMESTAMP
  mc admin user remove minio conditions-$TIMESTAMP
  mc admin user remove minio conditions-2-$TIMESTAMP
}

remove_policies() {
  mc admin policy remove minio bucketassignpolicy-$TIMESTAMP
  mc admin policy remove minio bucketread-$TIMESTAMP
  mc admin policy remove minio bucketwrite-$TIMESTAMP
  mc admin policy remove minio bucketcannottag-$TIMESTAMP
  mc admin policy remove minio dashboard-$TIMESTAMP
  mc admin policy remove minio diagnostics-$TIMESTAMP
  mc admin policy remove minio groups-$TIMESTAMP
  mc admin policy remove minio heal-$TIMESTAMP
  mc admin policy remove minio iampolicies-$TIMESTAMP
  mc admin policy remove minio logs-$TIMESTAMP
  mc admin policy remove minio notificationendpoints-$TIMESTAMP
  mc admin policy remove minio settings-$TIMESTAMP
  mc admin policy remove minio tiers-$TIMESTAMP
  mc admin policy remove minio trace-$TIMESTAMP
  mc admin policy remove minio users-$TIMESTAMP
  mc admin policy remove minio watch-$TIMESTAMP
  mc admin policy remove minio inspect-allowed-$TIMESTAMP
  mc admin policy remove minio inspect-not-allowed-$TIMESTAMPmc
  mc admin policy remove minio fix-prefix-policy-ui-crash-$TIMESTAMP
  mc admin policy remove minio conditions-policy-$TIMESTAMP
  mc admin policy remove minio conditions-policy-2-$TIMESTAMP
}

__init__() {
  export TIMESTAMP="$(cat web-app/tests/constants/timestamp.txt)"
  export GOPATH=/tmp/gopath
  export PATH=${PATH}:${GOPATH}/bin

  wget https://dl.min.io/client/mc/release/linux-amd64/mc
  chmod +x mc

  add_alias
}

main() {
  remove_users
  remove_policies
}

( __init__ "$@" && main "$@" )