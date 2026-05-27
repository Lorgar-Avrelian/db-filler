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

// Новые глобальные карты для хранения данных OID в памяти
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
)

func LoadRegistries(accessMap map[int16]string, varTypeMap map[int16]string, pollMap map[int16]string) {
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
}

// LoadOidRegistries динамически инициализирует справочники OID при старте
func LoadOidRegistries(asn1 map[int16]string, status map[int16]string, access map[int16]string, vendors []map[string]interface{}) {
	mu.Lock()
	defer mu.Unlock()

	for id, val := range asn1 {
		asn1TypeStrings[Asn1Type(id)] = val
		asn1TypeIds[strings.ToUpper(val)] = Asn1Type(id)
	}
	for id, val := range status {
		oidStatusStrings[OidStatus(id)] = val
		oidStatusIds[strings.ToUpper(val)] = OidStatus(id)
	}
	for id, val := range access {
		oidAccessStrings[OidAccess(id)] = val
		oidAccessIds[strings.ToUpper(val)] = OidAccess(id)
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

func (t Asn1Type) String() string {
	mu.RLock()
	defer mu.RUnlock()
	return asn1TypeStrings[t]
}

func (s OidStatus) String() string {
	mu.RLock()
	defer mu.RUnlock()
	return oidStatusStrings[s]
}

func (a OidAccess) String() string {
	mu.RLock()
	defer mu.RUnlock()
	return oidAccessStrings[a]
}

func (t Asn1Type) MarshalJSON() ([]byte, error)  { return []byte(fmt.Sprintf("%q", t.String())), nil }
func (s OidStatus) MarshalJSON() ([]byte, error) { return []byte(fmt.Sprintf("%q", s.String())), nil }
func (a OidAccess) MarshalJSON() ([]byte, error) { return []byte(fmt.Sprintf("%q", a.String())), nil }

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

func (a Access) String() string {
	mu.RLock()
	defer mu.RUnlock()
	return accessStrings[a]
}

func (v VarType) String() string {
	mu.RLock()
	defer mu.RUnlock()
	return varTypeStrings[v]
}

func (p PollingFrequency) String() string {
	mu.RLock()
	defer mu.RUnlock()
	return pollingFrequencyStrings[p]
}

func (a Access) MarshalJSON() ([]byte, error) { return []byte(fmt.Sprintf("%q", a.String())), nil }
func (a *Access) UnmarshalJSON(b []byte) error {
	*a = ParseAccess(strings.Trim(string(b), `"`))
	return nil
}
func (v VarType) MarshalJSON() ([]byte, error) { return []byte(fmt.Sprintf("%q", v.String())), nil }
func (v *VarType) UnmarshalJSON(b []byte) error {
	*v = ParseVarType(strings.Trim(string(b), `"`))
	return nil
}

func (p PollingFrequency) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", p.String())), nil
}

func (p *PollingFrequency) UnmarshalJSON(b []byte) error {
	*p = ParsePollingFrequency(strings.Trim(string(b), `"`))
	return nil
}
