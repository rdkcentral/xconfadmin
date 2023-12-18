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
package queries

import (
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/360EntSecGroup-Skylar/excelize"
	"xconfwebconfig/shared/estbfirmware"
)

func nextChar(ch rune) rune {
	if ch += 1; ch > 'z' {
		return 'a'
	}
	return ch
}

func doReport(macAddresses []string) ([]byte, error) {
	sort.Slice(macAddresses, func(i, j int) bool {
		return strings.Compare(strings.ToLower(macAddresses[i]), strings.ToLower(macAddresses[j])) < 0
	})
	xlsx := excelize.NewFile()

	headers := []string{
		"estbMac",
		"env",
		"model",
		"firmwareVersion",
		"time",
		"ipAddress",

		"rule type",
		"rule name",
		"noop",

		"filter name",
		"firmwareVersion(Config)",
		"firmwareFilename",
		"firmwareLocation",
		"firmwareDownloadProtocol",

		"lst chg env",
		"lst chg model",
		"lst chg firmwareVersion",
		"lst chg time",
		"lst chg ipAddress",

		"lst chg rule type",
		"lst chg rule name",
		"lst chg noop",

		"lst chg firmwareVersion(Config)",
		"lst chg firmwareFilename",
		"lst chg firmwareLocation",
		"lst chg firmwareDownloadProtocol",
	}

	data := map[string][]string{}

	for _, ma := range macAddresses {
		ll := estbfirmware.GetLastConfigLog(ma)
		if ll == nil {
			continue
		}

		if ll.Input != nil {
			data["estbMac"] = append(data["estbMac"], ll.Input.EstbMac)
			data["env"] = append(data["env"], ll.Input.Env)
			data["model"] = append(data["model"], ll.Input.Model)
			data["firmwareVersion"] = append(data["firmwareVersion"], ll.Input.FirmwareVersion)
			data["time"] = append(data["time"], ll.Input.Time.String())
			data["ipAddress"] = append(data["ipAddress"], ll.Input.IpAddress)
		}

		if ll.Rule != nil {
			data["rule type"] = append(data["rule type"], ll.Rule.Type)
			data["rule name"] = append(data["rule name"], ll.Rule.Name)
			data["noop"] = append(data["noop"], strconv.FormatBool(ll.Rule.NoOp))
		} else {
			data["rule type"] = append(data["rule type"], "")
			data["rule name"] = append(data["rule name"], "")
			data["noop"] = append(data["noop"], "")
		}

		if len(ll.Filters) > 0 {
			data["filter name"] = append(data["filter name"], ll.Filters[0].Name)
		} else {
			data["filter name"] = append(data["filter name"], "")
		}

		if ll.FirmwareConfig != nil {
			data["firmwareVersion(Config)"] = append(data["firmwareVersion(Config)"], ll.FirmwareConfig.GetFirmwareVersion())
			data["firmwareFilename"] = append(data["firmwareFilename"], ll.FirmwareConfig.GetFirmwareFilename())
			data["firmwareLocation"] = append(data["firmwareLocation"], ll.FirmwareConfig.GetFirmwareLocation())
			data["firmwareDownloadProtocol"] = append(data["firmwareDownloadProtocol"], ll.FirmwareConfig.GetFirmwareDownloadProtocol())
		} else {
			data["firmwareVersion(Config)"] = append(data["firmwareVersion(Config)"], "")
			data["firmwareFilename"] = append(data["firmwareFilename"], "")
			data["firmwareLocation"] = append(data["firmwareLocation"], "")
			data["firmwareDownloadProtocol"] = append(data["firmwareDownloadProtocol"], "")
		}

		chs := estbfirmware.GetConfigChangeLogsOnly(ma)

		if len(chs) == 0 {
			data["lst chg env"] = append(data["lst chg env"], "")
			data["lst chg model"] = append(data["lst chg model"], "")
			data["lst chg firmwareVersion"] = append(data["lst chg firmwareVersion"], "")
			data["lst chg time"] = append(data["lst chg time"], "")
			data["lst chg ipAddress"] = append(data["lst chg ipAddress"], "")

			data["lst chg rule type"] = append(data["lst chg rule type"], "")
			data["lst chg rule name"] = append(data["lst chg rule name"], "")
			data["lst chg noop"] = append(data["lst chg noop"], "")

			data["lst chg firmwareVersion(Config)"] = append(data["lst chg firmwareVersion(Config)"], "")
			data["lst chg firmwareFilename"] = append(data["lst chg firmwareFilename"], "")
			data["lst chg firmwareLocation"] = append(data["lst chg firmwareLocation"], "")
			data["lst chg firmwareDownloadProtocol"] = append(data["lst chg firmwareDownloadProtocol"], "")
		} else {
			cl := chs[0]
			if cl.Input != nil {
				data["lst chg env"] = append(data["lst chg env"], cl.Input.Env)
				data["lst chg model"] = append(data["lst chg model"], cl.Input.Model)
				data["lst chg firmwareVersion"] = append(data["lst chg firmwareVersion"], cl.Input.FirmwareVersion)
				data["lst chg time"] = append(data["lst chg time"], cl.Input.Time.String())
				data["lst chg ipAddress"] = append(data["lst chg ipAddress"], cl.Input.IpAddress)
			}

			if cl.Rule != nil {
				data["lst chg rule type"] = append(data["lst chg rule type"], cl.Rule.Type)
				data["lst chg rule name"] = append(data["lst chg rule name"], cl.Rule.Name)
				data["lst chg noop"] = append(data["lst chg noop"], strconv.FormatBool(cl.Rule.NoOp))
			} else {
				data["lst chg rule type"] = append(data["lst chg rule type"], "")
				data["lst chg rule name"] = append(data["lst chg rule name"], "")
				data["lst chg noop"] = append(data["lst chg noop"], "")
			}

			if cl.FirmwareConfig != nil {
				data["lst chg firmwareVersion(Config)"] = append(data["lst chg firmwareVersion(Config)"], cl.FirmwareConfig.GetFirmwareVersion())
				data["lst chg firmwareFilename"] = append(data["lst chg firmwareFilename"], cl.FirmwareConfig.GetFirmwareFilename())
				data["lst chg firmwareLocation"] = append(data["lst chg firmwareLocation"], cl.FirmwareConfig.GetFirmwareLocation())
				data["lst chg firmwareDownloadProtocol"] = append(data["lst chg firmwareDownloadProtocol"], cl.FirmwareConfig.GetFirmwareDownloadProtocol())
			} else {
				data["lst chg firmwareVersion(Config)"] = append(data["lst chg firmwareVersion(Config)"], "")
				data["lst chg firmwareFilename"] = append(data["lst chg firmwareFilename"], "")
				data["lst chg firmwareLocation"] = append(data["lst chg firmwareLocation"], "")
				data["lst chg firmwareDownloadProtocol"] = append(data["lst chg firmwareDownloadProtocol"], "")
			}
		}
	}

	col := 'A'
	for _, k := range headers {
		vlist := data[k]
		xlsx.SetCellValue("Sheet1", string(col)+"1", k)
		row := 2
		for _, v := range vlist {
			xlsx.SetCellValue("Sheet1", string(col)+strconv.Itoa(row), v)
			row = row + 1
		}
		col = nextChar(col)
	}

	fileName := "/tmp/temp.xlsx"
	err := xlsx.SaveAs(fileName)
	if err != nil {
		return nil, err
	}
	reportBytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	err = os.Remove(fileName)
	if err != nil {
		return nil, err
	}
	return reportBytes, nil
}
