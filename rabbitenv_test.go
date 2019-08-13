package rabbitenv

import (
	"testing"
)

// TestGetConfig tests rabbitenv.GetConfig
func TestGetConfig(t *testing.T) {
	congig := GetConfig()

	if congig["queue"] != "test" {
		t.Error("Config is incorrect")
	}

}
