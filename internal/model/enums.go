package model

import (
	"fmt"
	"strings"
	"sync"
)

type Access int16
type VarType int16
type PollingFrequency int16
type Asn1Type int16
type OidStatus int16
type OidAccess int16
type VendorID int64
type LogicOperator int16
type AlarmLevel int16

// Глобальные карты для хранения данных в памяти
var (
	mu                      sync.RWMutex
	accessStrings           = make(map[Access]string)
	accessIds               = make(map[string]Access)
	varTypeStrings          = make(map[VarType]string)
	varTypeIds              = make(map[string]VarType)
	pollingFrequencyStrings = make(map[PollingFrequency]string)
	pollingFrequencyIds     = make(map[string]PollingFrequency)
	asn1TypeStrings         = make(map[Asn1Type]string)
	asn1TypeIds             = make(map[string]Asn1Type)
	oidStatusStrings        = make(map[OidStatus]string)
	oidStatusIds            = make(map[string]OidStatus)
	oidAccessStrings        = make(map[OidAccess]string)
	oidAccessIds            = make(map[string]OidAccess)
	vendorNames             = make(map[string]VendorID)
	vendorDirectories       = make(map[string]VendorID)
	logicOperatorStrings    = make(map[LogicOperator]string)
	logicOperatorIds        = make(map[string]LogicOperator)
	alarmLevelStrings       = make(map[AlarmLevel]string)
	alarmLevelIds           = make(map[string]AlarmLevel)
)

// LoadRegistries — единственный централизованный метод инициализации всех справочников системы при старте
func LoadRegistries(
	accessMap map[int16]string,
	varTypeMap map[int16]string,
	pollMap map[int16]string,
	asn1Map map[int16]string,
	statusMap map[int16]string,
	oidAccessMap map[int16]string,
	logicMap map[int16]string,
	alarmMap map[int16]string,
	vendors []map[string]interface{},
) {
	mu.Lock()
	defer mu.Unlock()
	for id, val := range accessMap {
		accessStrings[Access(id)] = val
		accessIds[strings.ToUpper(val)] = Access(id)
	}
	for id, val := range varTypeMap {
		varTypeStrings[VarType(id)] = val
		varTypeIds[strings.ToUpper(val)] = VarType(id)
	}
	for id, val := range pollMap {
		pollingFrequencyStrings[PollingFrequency(id)] = val
		pollingFrequencyIds[strings.ToUpper(val)] = PollingFrequency(id)
	}
	for id, val := range asn1Map {
		asn1TypeStrings[Asn1Type(id)] = val
		asn1TypeIds[strings.ToUpper(val)] = Asn1Type(id)
	}
	for id, val := range statusMap {
		oidStatusStrings[OidStatus(id)] = val
		oidStatusIds[strings.ToUpper(val)] = OidStatus(id)
	}
	for id, val := range oidAccessMap {
		oidAccessStrings[OidAccess(id)] = val
		oidAccessIds[strings.ToUpper(val)] = OidAccess(id)
	}
	for id, val := range logicMap {
		logicOperatorStrings[LogicOperator(id)] = val
		logicOperatorIds[strings.ToUpper(val)] = LogicOperator(id)
	}
	for id, val := range alarmMap {
		alarmLevelStrings[AlarmLevel(id)] = val
		alarmLevelIds[strings.ToUpper(val)] = AlarmLevel(id)
	}
	for _, v := range vendors {
		vID := VendorID(v["id"].(int64))
		name := strings.ToUpper(v["name"].(string))
		vendorNames[name] = vID
		if dir, ok := v["directory"].(string); ok && dir != "" {
			vendorDirectories[strings.ToUpper(dir)] = vID
		}
	}
}

func LookupVendorID(identity string) (VendorID, bool) {
	mu.RLock()
	defer mu.RUnlock()
	clean := strings.ToUpper(strings.TrimSpace(identity))
	if id, ok := vendorNames[clean]; ok {
		return id, true
	}
	if id, ok := vendorDirectories[clean]; ok {
		return id, true
	}
	return 0, false
}

func (a Access) String() string  { mu.RLock(); defer mu.RUnlock(); return accessStrings[a] }
func (v VarType) String() string { mu.RLock(); defer mu.RUnlock(); return varTypeStrings[v] }
func (p PollingFrequency) String() string {
	mu.RLock()
	defer mu.RUnlock()
	return pollingFrequencyStrings[p]
}
func (t Asn1Type) String() string  { mu.RLock(); defer mu.RUnlock(); return asn1TypeStrings[t] }
func (s OidStatus) String() string { mu.RLock(); defer mu.RUnlock(); return oidStatusStrings[s] }
func (a OidAccess) String() string { mu.RLock(); defer mu.RUnlock(); return oidAccessStrings[a] }
func (l LogicOperator) String() string {
	mu.RLock()
	defer mu.RUnlock()
	return logicOperatorStrings[l]
}
func (a AlarmLevel) String() string            { mu.RLock(); defer mu.RUnlock(); return alarmLevelStrings[a] }
func (a Access) MarshalJSON() ([]byte, error)  { return []byte(fmt.Sprintf("%q", a.String())), nil }
func (v VarType) MarshalJSON() ([]byte, error) { return []byte(fmt.Sprintf("%q", v.String())), nil }
func (p PollingFrequency) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", p.String())), nil
}
func (t Asn1Type) MarshalJSON() ([]byte, error)  { return []byte(fmt.Sprintf("%q", t.String())), nil }
func (s OidStatus) MarshalJSON() ([]byte, error) { return []byte(fmt.Sprintf("%q", s.String())), nil }
func (a OidAccess) MarshalJSON() ([]byte, error) { return []byte(fmt.Sprintf("%q", a.String())), nil }
func (l LogicOperator) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", l.String())), nil
}
func (a AlarmLevel) MarshalJSON() ([]byte, error) { return []byte(fmt.Sprintf("%q", a.String())), nil }

func (a *Access) UnmarshalJSON(b []byte) error {
	*a = ParseAccess(strings.Trim(string(b), `"`))
	return nil
}
func (v *VarType) UnmarshalJSON(b []byte) error {
	*v = ParseVarType(strings.Trim(string(b), `"`))
	return nil
}
func (p *PollingFrequency) UnmarshalJSON(b []byte) error {
	*p = ParsePollingFrequency(strings.Trim(string(b), `"`))
	return nil
}
func (l *LogicOperator) UnmarshalJSON(b []byte) error {
	*l = ParseLogicOperator(strings.Trim(string(b), `"`))
	return nil
}
func (a *AlarmLevel) UnmarshalJSON(b []byte) error {
	*a = ParseAlarmLevel(strings.Trim(string(b), `"`))
	return nil
}
func (t *Asn1Type) UnmarshalJSON(b []byte) error {
	*t = ParseAsn1Type(strings.Trim(string(b), `"`))
	return nil
}
func (s *OidStatus) UnmarshalJSON(b []byte) error {
	*s = ParseOidStatus(strings.Trim(string(b), `"`))
	return nil
}
func (a *OidAccess) UnmarshalJSON(b []byte) error {
	*a = ParseOidAccess(strings.Trim(string(b), `"`))
	return nil
}

func ParseAccess(s string) Access {
	mu.RLock()
	defer mu.RUnlock()
	s = strings.ToUpper(strings.TrimSpace(s))
	if id, ok := accessIds[s]; ok {
		return id
	}
	return 1
}
func ParseVarType(s string) VarType {
	mu.RLock()
	defer mu.RUnlock()
	s = strings.ToUpper(strings.TrimSpace(s))
	if id, ok := varTypeIds[s]; ok {
		return id
	}
	return 1
}
func ParsePollingFrequency(s string) PollingFrequency {
	mu.RLock()
	defer mu.RUnlock()
	s = strings.ToUpper(strings.TrimSpace(s))
	if id, ok := pollingFrequencyIds[s]; ok {
		return id
	}
	return 1
}
func ParseLogicOperator(s string) LogicOperator {
	mu.RLock()
	defer mu.RUnlock()
	s = strings.ToUpper(strings.TrimSpace(s))
	if id, ok := logicOperatorIds[s]; ok {
		return id
	}
	return 1
}
func ParseAlarmLevel(s string) AlarmLevel {
	mu.RLock()
	defer mu.RUnlock()
	s = strings.ToUpper(strings.TrimSpace(s))
	if id, ok := alarmLevelIds[s]; ok {
		return id
	}
	return 1
}
func ParseAsn1Type(s string) Asn1Type {
	mu.RLock()
	defer mu.RUnlock()
	s = strings.ToUpper(strings.TrimSpace(s))
	if id, ok := asn1TypeIds[s]; ok {
		return id
	}
	return 1
}
func ParseOidStatus(s string) OidStatus {
	mu.RLock()
	defer mu.RUnlock()
	s = strings.ToUpper(strings.TrimSpace(s))
	if id, ok := oidStatusIds[s]; ok {
		return id
	}
	return 1
}
func ParseOidAccess(s string) OidAccess {
	mu.RLock()
	defer mu.RUnlock()
	s = strings.ToUpper(strings.TrimSpace(s))
	if id, ok := oidAccessIds[s]; ok {
		return id
	}
	return 1
}
