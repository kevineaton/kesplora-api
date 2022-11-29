package api

// BlockExternal is the `external` block type
type BlockExternal struct {
	BlockID      int64  `json:"blockId" db:"blockId"`
	ExternalLink string `json:"externalLink" db:"externalLink"`
}

// SaveBlockExternal saves the block content
func SaveBlockExternal(input *BlockExternal) error {
	input.processForDB()
	defer input.processForAPI()
	_, err := config.DBConnection.NamedExec(`INSERT INTO BlockExternal SET blockId = :blockId, externalLink = :externalLink ON DUPLICATE KEY UPDATE externalLink = :externalLink`, input)
	return err
}

// GetBlockExternalByBlockID gets the block content
func GetBlockExternalByBlockID(blockID int64) (*BlockExternal, error) {
	found := &BlockExternal{}
	defer found.processForAPI()
	err := config.DBConnection.Get(found, `SELECT * FROM BlockExternal WHERE blockId = ?`, blockID)
	return found, err
}

// DeleteBlockExternalByBlockID deletes the block content
func DeleteBlockExternalByBlockID(blockID int64) error {
	_, err := config.DBConnection.Exec("DELETE FROM BlockExternal WHERE blockId = ?", blockID)
	return err
}

func (input *BlockExternal) processForDB() {

}

func (input *BlockExternal) processForAPI() {

}
