package api

// BlockPresentation is the `presentation` block type
type BlockPresentation struct {
	BlockID          int64  `json:"blockId" db:"blockId"`
	EmbedLink        string `json:"embedLink" db:"embedLink"`
	PresentationType string `json:"presentationType" db:"presentationType"`
}

const (
	BlockPresentationTypePDF     = "pdf"
	BlockPresentationTypeYoutube = "youtube"
)

// SaveBlockPresentation saves the block content
func SaveBlockPresentation(input *BlockPresentation) error {
	input.processForDB()
	defer input.processForAPI()
	_, err := config.DBConnection.NamedExec(`INSERT INTO BlockPresentation SET blockId = :blockId, embedLink = :embedLink, presentationType = :presentationType 
		ON DUPLICATE KEY UPDATE embedLink = :embedLink, presentationType = :presentationType`, input)
	return err
}

// GetBlockPresentationByBlockID gets the block content
func GetBlockPresentationByBlockID(blockID int64) (*BlockPresentation, error) {
	found := &BlockPresentation{}
	defer found.processForAPI()
	err := config.DBConnection.Get(found, `SELECT * FROM BlockPresentation WHERE blockId = ?`, blockID)
	return found, err
}

// DeleteBlockPresentationByBlockID deletes the block content
func DeleteBlockPresentationByBlockID(blockID int64) error {
	_, err := config.DBConnection.Exec("DELETE FROM BlockPresentation WHERE blockId = ?", blockID)
	return err
}

func (input *BlockPresentation) processForDB() {
	if input.PresentationType == "" {
		input.PresentationType = "pdf"
	}
}

func (input *BlockPresentation) processForAPI() {

}
