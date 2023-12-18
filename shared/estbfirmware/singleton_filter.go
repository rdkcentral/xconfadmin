/**
 * Copyright 2023 Comcast Cable Communications Management, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package estbfirmware

import (
	"fmt"
	"strings"
	"xconfwebconfig/shared"
	coreef "xconfwebconfig/shared/estbfirmware"
)

func GetRoundRobinIdByApplication(applicationType string) string {
	if shared.STB == applicationType {
		return coreef.ROUND_ROBIN_FILTER_SINGLETON_ID
	}
	return fmt.Sprintf("%s_%s", strings.ToUpper(applicationType), coreef.ROUND_ROBIN_FILTER_SINGLETON_ID)
}
