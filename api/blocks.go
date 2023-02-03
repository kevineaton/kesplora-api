package api

import (
	"encoding/json"
	"errors"
	"net/http"
)

// Block is a block of content, which has the details filled out in linked tables
type Block struct {
	ID           int64       `json:"id" db:"id"`
	Name         string      `json:"name" db:"name"`
	Summary      string      `json:"summary" db:"summary"`
	BlockType    string      `json:"blockType" db:"blockType"`
	Content      interface{} `json:"content,omitempty"`
	FoundInFlows int64       `json:"foundInFlows" db:"foundInFlows"`
	AllowReset   string      `json:"allowReset" db:"allowReset"`

	UserStatus    string `json:"userStatus,omitempty" db:"userStatus"`
	LastUpdatedOn string `json:"lastUpdatedOn,omitempty" db:"lastUpdatedOn"`
}

const (
	BlockTypeExternal     = "external"
	BlockTypePresentation = "presentation"
	BlockTypeSurvey       = "survey"
	BlockTypeText         = "text"
)

var blockTypes = []string{
	BlockTypeExternal,
	BlockTypePresentation,
	BlockTypeSurvey,
	BlockTypeText,
}

func isValidBlockType(search string) bool {
	for i := range blockTypes {
		if blockTypes[i] == search {
			return true
		}
	}
	return false
}

// CreateBlock creates a new block
func CreateBlock(input *Block) error {
	input.processForDB()
	defer input.processForAPI()
	res, err := config.DBConnection.NamedExec(`INSERT INTO Blocks SET name = :name, summary = :summary, blockType = :blockType, allowReset = :allowReset`, input)
	if err != nil {
		return err
	}
	input.ID, _ = res.LastInsertId()
	return nil
}

// GetBlocksForSite gets the blocks for the site, usually used in admin views for linking and setting up flows
func GetBlocksForSite() ([]Block, error) {
	blocks := []Block{}
	err := config.DBConnection.Select(&blocks, `SELECT b.*, COUNT(bmf.moduleId) AS foundInFlows FROM
	Blocks b
	LEFT JOIN BlockModuleFlows bmf ON bmf.blockId = b.id 
	GROUP BY b.id, b.name, b.summary, b.blockType
	ORDER BY b.name`)
	for i := range blocks {
		blocks[i].processForAPI()
	}
	return blocks, err
}

// GetBlockByID gets a single block by id
func GetBlockByID(blockID int64) (*Block, error) {
	block := &Block{}
	defer block.processForAPI()
	err := config.DBConnection.Get(block, `SELECT b.*, COUNT(bmf.moduleId) AS foundInFlows FROM
	Blocks b
	LEFT JOIN BlockModuleFlows bmf ON bmf.blockId = b.id 
	WHERE b.id = ?
	GROUP BY b.id, b.name, b.summary, b.blockType`, blockID)
	return block, err
}

// GetModuleBlockForparticipant gets a single block for a participant; we take in all three levels to
// ensure that the permissions are correct
func GetModuleBlockForparticipant(participantID, projectID, moduleID, blockID int64) (*Block, error) {
	block := &Block{}
	defer block.processForAPI()
	err := config.DBConnection.Get(block, `SELECT b.*, 
	IFNULL(bus.status, 'not_started') AS userStatus,
	IFNULL(bus.lastUpdatedOn, NOW()) AS lastUpdatedOn
	FROM Blocks b, BlockModuleFlows bmf, Flows f
	LEFT JOIN BlockUserStatus bus ON bus.userId = ? AND bus.blockId = ?
	WHERE b.id = ? AND
	b.id = bmf.blockId AND
	bmf.moduleId = ? AND
	bmf.moduleId = f.moduleId AND
	f.projectId = ?;`, participantID, blockID, blockID, moduleID, projectID)
	return block, err
}

// UpdateBlock updates a block
func UpdateBlock(input *Block) error {
	input.processForDB()
	defer input.processForAPI()
	_, err := config.DBConnection.NamedExec(`UPDATE Blocks SET name = :name, blockType = :blockType, summary = :summary, allowReset = :allowReset WHERE id = :id`, input)
	return err
}

// GetBlocksForModule gets all of the blocks for a module
func GetBlocksForModule(moduleID int64) ([]Block, error) {
	blocks := []Block{}
	err := config.DBConnection.Select(&blocks, `SELECT b.* FROM Blocks b, BlockModuleFlows bmf WHERE bmf.moduleId = ? AND bmf.blockId = b.id ORDER BY bmf.flowOrder`, moduleID)
	for i := range blocks {
		blocks[i].processForAPI()
	}
	return blocks, err
}

// DeleteBlock deletes a block
func DeleteBlock(blockID int64) error {
	_, err := config.DBConnection.Exec(`DELETE FROM Blocks WHERE id = ?`, blockID)
	if err != nil {
		return err
	}

	_, err = config.DBConnection.Exec(`DELETE FROM BlockModuleFlows WHERE blockId = ?`, blockID)
	if err != nil {
		return err
	}

	return nil
}

// LinkBlockAndModule links a block and a module
func LinkBlockAndModule(moduleID int64, blockID int64, order int64) error {
	_, err := config.DBConnection.Exec(`INSERT INTO BlockModuleFlows SET 
	moduleId = ?, blockId = ?, flowOrder = ? ON DUPLICATE KEY UPDATE flowOrder = ?`, moduleID, blockID, order, order)
	return err
}

// UnlinkBlockAndModule unlinks a block and a module
func UnlinkBlockAndModule(moduleID int64, blockID int64) error {
	_, err := config.DBConnection.Exec(`DELETE FROM BlockModuleFlows WHERE moduleId = ? AND blockId = ?`, moduleID, blockID)
	return err
}

// UnlinkAllBlocksFromModule unlinks all blocks from a module
func UnlinkAllBlocksFromModule(moduleID int64) error {
	_, err := config.DBConnection.Exec(`DELETE FROM BlockModuleFlows WHERE moduleId = ?`, moduleID)
	return err
}

//
// these handles probably need to be refactored; we want to make sure the API can take in the arbitrary data, but converting
// using the json mapping seems... overkill
//

// handleBlockRequiredFields is a quick helper to validate data before a block is saved
func handleBlockRequiredFields(blockType string, rawData interface{}) error {
	str, _ := json.Marshal(rawData)
	switch blockType {
	case BlockTypeExternal:
		content := &BlockExternal{}
		err := json.Unmarshal(str, content)
		if err != nil || content.ExternalLink == "" {
			return errors.New("invalid")
		}
	case BlockTypePresentation:
		content := &BlockPresentation{}
		err := json.Unmarshal(str, content)
		if err != nil || content.EmbedLink == "" {
			return errors.New("invalid")
		}
	case BlockTypeText:
		content := &BlockText{}
		err := json.Unmarshal(str, content)
		if err != nil || content.Text == "" {
			return errors.New("invalid")
		}
	default:
		return errors.New("invalid type")
	}
	return nil
}

// handleBlockSave is a helper for creating and updating block content types
func handleBlockSave(blockType string, blockID int64, rawData interface{}) (interface{}, error) {
	str, _ := json.Marshal(rawData)
	switch blockType {
	case BlockTypeExternal:
		content := &BlockExternal{}
		err := json.Unmarshal(str, content)
		if err != nil {
			return content, errors.New("could not convert")
		}
		content.BlockID = blockID
		err = SaveBlockExternal(content)
		return content, err
	case BlockTypePresentation:
		content := &BlockPresentation{}
		err := json.Unmarshal(str, content)
		if err != nil {
			return content, errors.New("could not convert")
		}
		content.BlockID = blockID
		err = SaveBlockPresentation(content)
		return content, err
	case BlockTypeSurvey:
		return rawData, errors.New("not implemented")
	case BlockTypeText:
		content := &BlockText{}
		err := json.Unmarshal(str, content)
		if err != nil {
			return content, errors.New("could not convert")
		}
		content.BlockID = blockID
		err = SaveBlockText(content)
		return content, err
	}
	return rawData, errors.New("unsupported type")
}

// handleBlockGet is a helper for getting the content for a block
func handleBlockGet(blockType string, blockID int64) (interface{}, error) {
	switch blockType {
	case BlockTypeExternal:
		found, err := GetBlockExternalByBlockID(blockID)
		return found, err
	case BlockTypePresentation:
		found, err := GetBlockPresentationByBlockID(blockID)
		return found, err
	case BlockTypeSurvey:
		return map[string]string{}, errors.New("not implemented")
	case BlockTypeText:
		found, err := GetBlockTextByBlockID(blockID)
		return found, err
	}
	return map[string]string{}, errors.New("unsupported type")
}

// handleBlockDelete is a helper for deleting a block and its content
func handleBlockDelete(blockType string, blockID int64) error {
	switch blockType {
	case BlockTypeExternal:
		err := DeleteBlockExternalByBlockID(blockID)
		return err
	case BlockTypePresentation:
		err := DeleteBlockPresentationByBlockID(blockID)
		return err
	case BlockTypeSurvey:
		return errors.New("not implemented")
	case BlockTypeText:
		err := DeleteBlockTextByBlockID(blockID)
		return err
	}
	return errors.New("unsupported type")
}

func (input *Block) processForDB() {
	if input.BlockType == "" {
		input.BlockType = BlockTypeText
	}
	if input.AllowReset == "" {
		input.AllowReset = Yes
	}
}

func (input *Block) processForAPI() {
	if input.Content == nil {
		input.Content = map[string]string{}
	}
	if input.LastUpdatedOn != "" {
		input.LastUpdatedOn, _ = parseTimeToTimeFormat(input.LastUpdatedOn, timeFormatAPI)
	}
}

// Bind binds the data for the HTTP
func (data *Block) Bind(r *http.Request) error {
	return nil
}
