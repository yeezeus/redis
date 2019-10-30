#!/bin/bash

# Copyright The KubeDB Authors.
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


set -e

# needs args in following manner:
#    src_node_ip
#    src_node_id
#    dst_node_ip
#    dst_node_id
#    slot_start
#    slot_end
reshard() {
    for i in `seq $5 $6`; do
        redis-cli -c -h $3 cluster setslot ${i} importing $2
        redis-cli -c -h $1 cluster setslot ${i} migrating $4
        while true; do
            key=`redis-cli -c -h $1 cluster getkeysinslot ${i} 1`
            if [ "" = "$key" ]; then
                echo "there are no key in this slot ${i}"
                break
            fi
            redis-cli -h $1 migrate $3 6379 ${key} 0 5000
        done
        redis-cli -c -h $1 cluster setslot ${i} node $4
        redis-cli -c -h $3 cluster setslot ${i} node $4
    done
}

"$@"
