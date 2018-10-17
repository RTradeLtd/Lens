package utils_test

import (
	"testing"

	"github.com/RTradeLtd/Lens/utils"
)

var (
	data = []string{"v", "v", "c", "c"}
)

func TestUnique(t *testing.T) {
	uniqued := utils.Unique(data)
	if len(uniqued) > 2 || len(uniqued) < 2 {
		t.Fatal("failed to uniquefy data")
	}
}
