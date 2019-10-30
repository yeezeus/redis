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

FROM redis:5.0.3-alpine

RUN mkdir -p /conf
COPY fix-ip.sh /conf/fix-ip.sh
# ENTRYPOINT ["/conf/fix-ip.sh"]

CMD ["/conf/fix-ip.sh", "redis-server"]
