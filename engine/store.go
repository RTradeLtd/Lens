package engine

import "github.com/RTradeLtd/Lens/models"

type DocProps struct {
	Indexed string `json:"indexed"`
}

type DocData struct {
	Content    string             `json:"content"`
	Metadata   *models.MetaDataV2 `json:"metadata"`
	Properties *DocProps          `json:"properties"`
}
