# ------------------------------------------------------------------------
#
# Copyright 2019 WSO2, Inc. (http://wso2.com)
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License
#
# ------------------------------------------------------------------------

# set to product base image
FROM wso2/wso2am-analytics-worker:2.6.0

ARG ANALYTICS_FILES=./target/files

COPY --chown=wso2carbon:wso2 ${ANALYTICS_FILES}/lib/*.jar ${WSO2_SERVER_HOME}/lib/
COPY --chown=wso2carbon:wso2 ${ANALYTICS_FILES}/siddhi-files/* ${WSO2_SERVER_HOME}/wso2/worker/deployment/siddhi-files/

# expose ports
EXPOSE 9091 9444 7712 7612 9613 9713 7444 7071 7575 7576 7577
