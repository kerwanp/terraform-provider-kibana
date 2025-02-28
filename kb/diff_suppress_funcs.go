package kb

import (
	"fmt"
	"strings"

	"github.com/elastic/go-ucfg"
	"github.com/elastic/go-ucfg/diff"
	ucfgjson "github.com/elastic/go-ucfg/json"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	log "github.com/sirupsen/logrus"
)

// suppressEquivalentJSON permit to compare json string
func suppressEquivalentJSON(k, old, new string, d *schema.ResourceData) bool {
	if old == "" {
		old = `{}`
	}
	if new == "" {
		new = `{}`
	}
	confOld, err := ucfgjson.NewConfig([]byte(old), ucfg.PathSep("."))
	if err != nil {
		fmt.Printf("[ERR] Error when converting current Json: %s\ndata: %s", err.Error(), old)
		log.Errorf("Error when converting current Json: %s\ndata: %s", err.Error(), old)
		return false
	}
	confNew, err := ucfgjson.NewConfig([]byte(new), ucfg.PathSep("."))
	if err != nil {
		fmt.Printf("[ERR] Error when converting new Json: %s\ndata: %s", err.Error(), new)
		log.Errorf("Error when converting new Json: %s\ndata: %s", err.Error(), new)
		return false
	}

	currentDiff := diff.CompareConfigs(confOld, confNew)
	log.Debugf("Diff\n: %s", currentDiff.GoStringer())

	return !currentDiff.HasChanged()
}

// suppressEquivalentNDJSON permit to compare ndjson string
func suppressEquivalentNDJSON(k, old, new string, d *schema.ResourceData) bool {

	// NDJSON mean sthat each line correspond to JSON struct
	oldSlice := strings.Split(old, "\n")
	newSlice := strings.Split(new, "\n")
	oldObjSlice := make([]*ucfg.Config, len(oldSlice))
	newObjSlice := make([]*ucfg.Config, len(newSlice))
	if len(oldSlice) != len(newSlice) {
		return false
	}

	// Convert string line to JSON
	for i, oldJSON := range oldSlice {
		if oldJSON == "" {
			oldJSON = `{}`
		}
		config, err := ucfgjson.NewConfig([]byte(oldJSON), ucfg.PathSep("."))
		if err != nil {
			fmt.Printf("[ERR] Error when converting current Json: %s\ndata: %s", err.Error(), oldJSON)
			log.Errorf("Error when converting current Json: %s\ndata: %s", err.Error(), oldJSON)
			return false
		}
		config.Remove("version", -1)
		config.Remove("updated_at", -1)

		oldObjSlice[i] = config
	}
	for i, newJSON := range newSlice {
		if newJSON == "" {
			newJSON = `{}`
		}
		config, err := ucfgjson.NewConfig([]byte(newJSON), ucfg.PathSep("."))
		if err != nil {
			fmt.Printf("[ERR] Error when converting new Json: %s\ndata: %s", err.Error(), newJSON)
			log.Errorf("Error when converting new Json: %s\ndata: %s", err.Error(), newJSON)
			return false
		}
		config.Remove("version", -1)
		config.Remove("updated_at", -1)

		newObjSlice[i] = config
	}

	// Compare json obj
	for i, oldConfig := range oldObjSlice {
		isFound := false
		if !oldConfig.HasField("id") {
			return false
		}
		oldId, err := oldConfig.String("id", -1)
		if err != nil {
			log.Errorf("Error when get ID on current Json: %s\ndata: %s", err.Error(), oldSlice[i])
			fmt.Printf("[ERR] Error when get ID on current Json: %s\ndata: %s", err.Error(), oldSlice[i])
			return false
		}
		for j, newConfig := range newObjSlice {
			if !newConfig.HasField("id") {
				return false
			}
			newId, err := newConfig.String("id", -1)
			if err != nil {
				log.Errorf("Error when get ID on new Json: %s\ndata: %s", err.Error(), newSlice[j])
				fmt.Printf("[ERR] Error when get ID on new Json: %s\ndata: %s", err.Error(), newSlice[j])
				return false
			}
			if oldId == newId {
				currentDiff := diff.CompareConfigs(oldConfig, newConfig)
				log.Debugf("Diff\n: %s", currentDiff.GoStringer())

				if currentDiff.HasChanged() {
					return false
				}
				isFound = true
				break
			}
		}

		if isFound == false {
			return false
		}
	}

	return true
}
