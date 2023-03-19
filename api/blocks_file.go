package api

// BlockFile is the `file` block type
type BlockFile struct {
	BlockID int64 `json:"blockId" db:"blockId"`
	FileID  int64 `json:"fileId" db:"fileId"`
}

// SaveBlockFile saves the block content
func SaveBlockFile(input *BlockFile) error {
	input.processForDB()
	defer input.processForAPI()
	_, err := config.DBConnection.NamedExec(`INSERT INTO BlockFile SET blockId = :blockId, fileId = :fileId
		ON DUPLICATE KEY UPDATE fileId = :fileId`, input)
	return err
}

// GetBlockFileByBlockID gets the block content
func GetBlockFileByBlockID(blockID int64) (*BlockEmbed, error) {
	found := &BlockEmbed{}
	defer found.processForAPI()
	err := config.DBConnection.Get(found, `SELECT * FROM BlockFile WHERE blockId = ?`, blockID)
	return found, err
}

// DeleteBlockFileByBlockID deletes the block content
func DeleteBlockFileByBlockID(blockID int64) error {
	_, err := config.DBConnection.Exec("DELETE FROM BlockFile WHERE blockId = ?", blockID)
	return err
}

func (input *BlockFile) processForDB() {
}

func (input *BlockFile) processForAPI() {

}
