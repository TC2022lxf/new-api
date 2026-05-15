/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
package model

import (
	"github.com/QuantumNous/new-api/common"
)

// migrateLegacyBillingGroupOptions copies billing-related config from legacy keys
// vip/svip into limited_offer/free when the new keys are absent, so ratios and
// usable-group lists stay aligned with updated channel/model group slugs.
// Legacy keys are kept for backward compatibility with existing data.
func migrateLegacyBillingGroupOptions() {
	migrateLegacyGroupRatio()
	migrateLegacyUserUsableGroups()
	migrateLegacyTopupGroupRatio()
	migrateLegacyGroupGroupRatio()
}

func migrateLegacyGroupRatio() {
	common.OptionMapRWMutex.RLock()
	raw := common.OptionMap["GroupRatio"]
	common.OptionMapRWMutex.RUnlock()
	if raw == "" {
		return
	}
	var m map[string]float64
	if err := common.UnmarshalJsonStr(raw, &m); err != nil || len(m) == 0 {
		return
	}
	changed := false
	if _, has := m["limited_offer"]; !has {
		if v, ok := m["vip"]; ok {
			m["limited_offer"] = v
			changed = true
		}
	}
	if _, has := m["free"]; !has {
		if v, ok := m["svip"]; ok {
			m["free"] = v
			changed = true
		} else {
			m["free"] = 1
			changed = true
		}
	}
	if !changed {
		return
	}
	b, err := common.Marshal(m)
	if err != nil {
		return
	}
	if err := UpdateOption("GroupRatio", string(b)); err != nil {
		common.SysLog("migrate GroupRatio: " + err.Error())
	}
}

func migrateLegacyUserUsableGroups() {
	common.OptionMapRWMutex.RLock()
	raw := common.OptionMap["UserUsableGroups"]
	common.OptionMapRWMutex.RUnlock()
	if raw == "" {
		return
	}
	var m map[string]string
	if err := common.UnmarshalJsonStr(raw, &m); err != nil || len(m) == 0 {
		return
	}
	changed := false
	if _, has := m["limited_offer"]; !has {
		if d, ok := m["vip"]; ok {
			m["limited_offer"] = d
		} else {
			m["limited_offer"] = "限时优惠"
		}
		changed = true
	}
	if _, has := m["free"]; !has {
		if d, ok := m["svip"]; ok {
			m["free"] = d
		} else {
			m["free"] = "免费"
		}
		changed = true
	}
	if !changed {
		return
	}
	b, err := common.Marshal(m)
	if err != nil {
		return
	}
	if err := UpdateOption("UserUsableGroups", string(b)); err != nil {
		common.SysLog("migrate UserUsableGroups: " + err.Error())
	}
}

func migrateLegacyTopupGroupRatio() {
	common.OptionMapRWMutex.RLock()
	raw := common.OptionMap["TopupGroupRatio"]
	common.OptionMapRWMutex.RUnlock()
	if raw == "" {
		return
	}
	var m map[string]float64
	if err := common.UnmarshalJsonStr(raw, &m); err != nil || len(m) == 0 {
		return
	}
	changed := false
	if _, has := m["limited_offer"]; !has {
		if v, ok := m["vip"]; ok {
			m["limited_offer"] = v
			changed = true
		}
	}
	if _, has := m["free"]; !has {
		if v, ok := m["svip"]; ok {
			m["free"] = v
		} else {
			m["free"] = 1
		}
		changed = true
	}
	if !changed {
		return
	}
	b, err := common.Marshal(m)
	if err != nil {
		return
	}
	if err := UpdateOption("TopupGroupRatio", string(b)); err != nil {
		common.SysLog("migrate TopupGroupRatio: " + err.Error())
	}
}

func migrateLegacyGroupGroupRatio() {
	common.OptionMapRWMutex.RLock()
	raw := common.OptionMap["GroupGroupRatio"]
	common.OptionMapRWMutex.RUnlock()
	if raw == "" {
		return
	}
	var m map[string]map[string]float64
	if err := common.UnmarshalJsonStr(raw, &m); err != nil || len(m) == 0 {
		return
	}
	changed := mergeNestedGroupRatioKeys(m)
	if !changed {
		return
	}
	b, err := common.Marshal(m)
	if err != nil {
		return
	}
	s := string(b)
	if err := UpdateOption("GroupGroupRatio", s); err != nil {
		common.SysLog("migrate GroupGroupRatio: " + err.Error())
	}
}

func mergeNestedGroupRatioKeys(m map[string]map[string]float64) bool {
	changed := false
	if inner, ok := m["vip"]; ok {
		if _, has := m["limited_offer"]; !has {
			m["limited_offer"] = cloneFloatMap(inner)
			changed = true
		}
	}
	if inner, ok := m["svip"]; ok {
		if _, has := m["free"]; !has {
			m["free"] = cloneFloatMap(inner)
			changed = true
		}
	}
	for _, inner := range m {
		if mergeInnerUsingGroupKeys(inner) {
			changed = true
		}
	}
	return changed
}

func mergeInnerUsingGroupKeys(inner map[string]float64) bool {
	changed := false
	if _, has := inner["limited_offer"]; !has {
		if v, ok := inner["vip"]; ok {
			inner["limited_offer"] = v
			changed = true
		}
	}
	if _, has := inner["free"]; !has {
		if v, ok := inner["svip"]; ok {
			inner["free"] = v
			changed = true
		}
	}
	return changed
}

func cloneFloatMap(src map[string]float64) map[string]float64 {
	if len(src) == 0 {
		return map[string]float64{}
	}
	out := make(map[string]float64, len(src))
	for k, v := range src {
		out[k] = v
	}
	_ = mergeInnerUsingGroupKeys(out)
	return out
}
