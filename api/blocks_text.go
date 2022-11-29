package api

// BlockText is the `text` block type
type BlockText struct {
	BlockID int64  `json:"blockId" db:"blockId"`
	Text    string `json:"text" db:"text"`
}

// SaveBlockText saves the block content
func SaveBlockText(input *BlockText) error {
	input.processForDB()
	defer input.processForAPI()
	_, err := config.DBConnection.NamedExec(`INSERT INTO BlockText SET blockId = :blockId, text = :text ON DUPLICATE KEY UPDATE text = :text`, input)
	return err
}

// GetBlockTextByBlockID gets the block content
func GetBlockTextByBlockID(blockID int64) (*BlockText, error) {
	found := &BlockText{}
	defer found.processForAPI()
	err := config.DBConnection.Get(found, `SELECT * FROM BlockText WHERE blockId = ?`, blockID)
	return found, err
}

// DeleteBlockTextByBlockID deletes the block content
func DeleteBlockTextByBlockID(blockID int64) error {
	_, err := config.DBConnection.Exec("DELETE FROM BlockText WHERE blockId = ?", blockID)
	return err
}

func (input *BlockText) processForDB() {

}

func (input *BlockText) processForAPI() {

}
