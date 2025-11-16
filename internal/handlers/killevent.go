package handlers

import (
	"encoding/json"
	"log"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
)

// ensures that the Killevent struct satisfy the core.RecordProxy interface
var _ core.RecordProxy = (*Killevent)(nil)

type Killevent struct {
	core.BaseRecordProxy
	cachedData map[string]interface{} // Cache the parsed data field
	dataParsed bool                   // Track if data has been parsed
}

type Killer struct {
	Name    string `json:"Name"`
	SteamID string `json:"SteamID"`
	Team    int    `json:"Team"`
}

// All event data is stored in the 'data' JSON field of the record.
func (k *Killevent) getDataMap() map[string]interface{} {
	// Return cached data if already parsed
	if k.dataParsed {
		return k.cachedData
	}

	var data map[string]interface{}
	raw := k.GetString("data")
	if raw == "" {
		k.dataParsed = true
		return nil
	}
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		log.Printf("[Killevent] Failed to unmarshal data field: %v", err)
		k.dataParsed = true
		return nil
	}

	// Cache the result
	k.cachedData = data
	k.dataParsed = true
	return data
}

func (k *Killevent) Killers() []Killer {
	data := k.getDataMap()
	var killers []Killer
	if data == nil {
		return killers
	}
	raw, ok := data["killers"]
	if !ok || raw == nil {
		return killers
	}
	// Efficient unmarshaling: marshal the raw interface{} and unmarshal to typed slice
	if b, err := json.Marshal(raw); err == nil {
		if err := json.Unmarshal(b, &killers); err == nil {
			return killers
		}
	}
	log.Printf("[Killevent] Failed to unmarshal killers")
	return nil
}

// Victim returns the victim as a Killer struct (Name, SteamID, Team)
func (k *Killevent) Victim() Killer {
	data := k.getDataMap()
	var victim Killer
	if data == nil {
		return victim
	}
	raw, ok := data["victim"]
	if !ok || raw == nil {
		return victim
	}
	// Efficient unmarshaling
	if b, err := json.Marshal(raw); err == nil {
		if err := json.Unmarshal(b, &victim); err == nil {
			return victim
		}
	}
	log.Printf("[Killevent] Failed to unmarshal victim")
	return Killer{}
}

func (k *Killevent) VictimSteamID() string {
	return k.Victim().SteamID
}

func (k *Killevent) VictimIsPlayer() bool {
	steamID := k.VictimSteamID()
	return steamID != "" && steamID != "INVALID"
}

func (k *Killevent) VictimName() string {
	return k.Victim().Name
}

func (k *Killevent) VictimTeam() int {
	return k.Victim().Team
}

func (k *Killevent) Weapon() string {
	data := k.getDataMap()
	if data == nil {
		return ""
	}
	v, _ := data["weapon"].(string)
	return v
}

func (k *Killevent) IsHeadshot() bool {
	data := k.getDataMap()
	if data == nil {
		return false
	}
	v, _ := data["is_headshot"].(bool)
	return v
}

func (k *Killevent) IsCatchup() bool {
	data := k.getDataMap()
	if data == nil {
		return false
	}
	v, _ := data["is_catchup"].(bool)
	return v
}

func (k *Killevent) Created() types.DateTime {
	return k.GetDateTime("created")
}

func (k *Killevent) Updated() types.DateTime {
	return k.GetDateTime("updated")
}
