package api

// BlockEmbed is the `presentation` block type
type BlockEmbed struct {
	BlockID   int64  `json:"blockId" db:"blockId"`
	EmbedLink string `json:"embedLink" db:"embedLink"`
	EmbedType string `json:"embedType" db:"embedType"`
	FileID    int64  `json:"fileId" db:"fileId"`
}

const (
	BlockEmbedTypeExternalPDF = "external_pdf"
	BlockEmbedTypeInternalPDF = "internal_pdf"
	BlockEmbedTypeYoutube     = "youtube"
)

// SaveBlockEmbed saves the block content
func SaveBlockEmbed(input *BlockEmbed) error {
	input.processForDB()
	defer input.processForAPI()
	_, err := config.DBConnection.NamedExec(`INSERT INTO BlockEmbed SET blockId = :blockId, embedLink = :embedLink, embedType = :embedType, fileId = :fileId 
		ON DUPLICATE KEY UPDATE embedLink = :embedLink, embedType = :embedType, fileId = :fileId`, input)
	return err
}

// GetBlockEmbedByBlockID gets the block content
func GetBlockEmbedByBlockID(blockID int64) (*BlockEmbed, error) {
	found := &BlockEmbed{}
	defer found.processForAPI()
	err := config.DBConnection.Get(found, `SELECT * FROM BlockEmbed WHERE blockId = ?`, blockID)
	return found, err
}

// DeleteBlockEmbedByBlockID deletes the block content
func DeleteBlockEmbedByBlockID(blockID int64) error {
	_, err := config.DBConnection.Exec("DELETE FROM BlockEmbed WHERE blockId = ?", blockID)
	return err
}

func (input *BlockEmbed) processForDB() {
	if input.EmbedType == "" {
		input.EmbedType = BlockEmbedTypeExternalPDF
	}
}

func (input *BlockEmbed) processForAPI() {
	if input.FileID != 0 {
		input.EmbedLink = ""
	}
}
