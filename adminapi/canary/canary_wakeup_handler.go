/**
 * Copyright 2025 Comcast Cable Communications Management, LLC
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
package canary

import (
	"fmt"
	"net/http"

	"github.com/rdkcentral/xconfadmin/adminapi/auth"
	"github.com/rdkcentral/xconfadmin/common"
	xhttp "github.com/rdkcentral/xconfadmin/http"
	"github.com/rdkcentral/xconfadmin/taggingapi/tag"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"

	log "github.com/sirupsen/logrus"
)

// RemoveCanaryWakeupTagHandler removes all the data from XDAS associated with the t_canary_wakeup tag
func RemoveCanaryWakeupTagHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanWrite(r, auth.COMMON_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}
	tagName := common.WakeupPoolTagName
	// Get all members of the tag
	members, err := tag.GetTagMembers(tagName)
	if err != nil {
		log.Errorf("Error getting members from %s tag: %s", tagName, err.Error())
		xwhttp.WriteXconfResponse(w, http.StatusInternalServerError, []byte(fmt.Sprintf("Error getting members: %s", err.Error())))
		return
	}

	if len(members) == 0 {
		xwhttp.WriteXconfResponse(w, http.StatusOK, []byte(fmt.Sprintf("No members found in %s tag", tagName)))
		return
	}

	// Remove each member from the tag
	count, err := tag.RemoveMembersFromTag(tagName, members)
	if err != nil {
		log.Errorf("Error removing members from %s tag: %s", tagName, err.Error())
		xwhttp.WriteXconfResponse(w, http.StatusInternalServerError, []byte(fmt.Sprintf("Error removing members: %s", err.Error())))
		return
	}

	response := fmt.Sprintf("Successfully removed %d members from %s tag", count, tagName)
	xwhttp.WriteXconfResponse(w, http.StatusOK, []byte(response))
}
