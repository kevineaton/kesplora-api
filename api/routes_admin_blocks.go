package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

// routeAdminCreateBlock creates a new block id and then processes the content for saving
func routeAdminCreateBlock(w http.ResponseWriter, r *http.Request) {
	blockType := chi.URLParam(r, "blockType")
	if !isValidBlockType(blockType) {
		sendAPIError(w, api_error_block_invalid_type, errors.New("invalid type"), map[string]string{
			"blockType": blockType,
		})
		return
	}

	input := &Block{}
	render.Bind(r, input)

	if input.Name == "" {
		sendAPIError(w, api_error_block_missing_data, errors.New("missing data"), map[string]interface{}{
			"input": input,
		})
		return
	}

	// validate the content before committing to a save
	err := handleBlockRequiredFields(blockType, input.Content)
	if err != nil {
		sendAPIError(w, api_error_block_content_missing_data, err, map[string]interface{}{
			"input": input,
		})
		return
	}

	err = CreateBlock(input)
	if err != nil {
		sendAPIError(w, api_error_block_save_error, err, map[string]interface{}{
			"input": input,
		})
		return
	}

	// now, hand off the save depending on the body of the content
	data, err := handleBlockSave(blockType, input.ID, input.Content)
	if err != nil {
		sendAPIError(w, api_error_block_save_error, err, map[string]interface{}{
			"input": input,
		})
		return
	}
	input.Content = data
	sendAPIJSONData(w, http.StatusCreated, input)
}

// routeAdminGetBlocksOnSite gets the meta data about all modules on the platform
func routeAdminGetBlocksOnSite(w http.ResponseWriter, r *http.Request) {
	blocks, err := GetBlocksForSite()
	if err != nil {
		sendAPIError(w, api_error_block_not_found, err, map[string]interface{}{})
		return
	}
	sendAPIJSONData(w, http.StatusOK, blocks)
}

// routeAdminGetBlocksForModule gets the blocks on a module
func routeAdminGetBlocksForModule(w http.ResponseWriter, r *http.Request) {
	moduleID, moduleIDErr := strconv.ParseInt(chi.URLParam(r, "moduleID"), 10, 64)
	if moduleIDErr != nil {
		sendAPIError(w, api_error_invalid_path, moduleIDErr, map[string]string{})
		return
	}

	_, err := GetModuleByID(moduleID)
	if err != nil {
		sendAPIError(w, api_error_module_not_found, err, map[string]interface{}{})
		return
	}

	blocks, err := GetBlocksForModule(moduleID)
	if err != nil {
		sendAPIError(w, api_error_block_not_found, err, map[string]interface{}{})
		return
	}
	sendAPIJSONData(w, http.StatusOK, blocks)
}

// routeAdminUnlinkAllBlocksFromModule unlinks all blocks from a module
func routeAdminUnlinkAllBlocksFromModule(w http.ResponseWriter, r *http.Request) {
	moduleID, moduleIDErr := strconv.ParseInt(chi.URLParam(r, "moduleID"), 10, 64)
	if moduleIDErr != nil {
		sendAPIError(w, api_error_invalid_path, moduleIDErr, map[string]string{})
		return
	}

	_, err := GetModuleByID(moduleID)
	if err != nil {
		sendAPIError(w, api_error_module_not_found, err, map[string]interface{}{})
		return
	}

	err = UnlinkAllBlocksFromModule(moduleID)
	if err != nil {
		sendAPIError(w, api_error_block_not_found, err, map[string]interface{}{})
		return
	}
	sendAPIJSONData(w, http.StatusOK, map[string]bool{
		"unlinkedAll": true,
	})
}

// routeAdminGetBlock gets the block and content
func routeAdminGetBlock(w http.ResponseWriter, r *http.Request) {
	blockID, blockIDErr := strconv.ParseInt(chi.URLParam(r, "blockID"), 10, 64)
	if blockIDErr != nil {
		sendAPIError(w, api_error_invalid_path, blockIDErr, map[string]string{})
		return
	}

	block, err := GetBlockByID(blockID)
	if err != nil {
		sendAPIError(w, api_error_block_not_found, err, map[string]interface{}{})
		return
	}
	content, err := handleBlockGet(block.BlockType, block.ID)
	if err != nil {
		sendAPIError(w, api_error_block_not_found, err, map[string]interface{}{})
		return
	}
	block.Content = content
	sendAPIJSONData(w, http.StatusOK, block)
}

// routeAdminUpdateBlock updates a block and its content
func routeAdminUpdateBlock(w http.ResponseWriter, r *http.Request) {
	blockID, blockIDErr := strconv.ParseInt(chi.URLParam(r, "blockID"), 10, 64)
	if blockIDErr != nil {
		sendAPIError(w, api_error_invalid_path, blockIDErr, map[string]string{})
		return
	}

	block, err := GetBlockByID(blockID)
	if err != nil {
		sendAPIError(w, api_error_block_not_found, err, map[string]interface{}{})
		return
	}

	input := &Block{}
	render.Bind(r, input)
	if block.Name != "" && block.Name != input.Name {
		block.Name = input.Name
	}
	if block.Summary != "" && block.Summary != input.Summary {
		block.Summary = input.Summary
	}
	err = UpdateBlock(block)
	if err != nil {
		sendAPIError(w, api_error_block_save_error, err, map[string]interface{}{
			"input": input,
		})
		return
	}
	// now, we take the content and send it straight through to saving
	if input.Content != nil {
		content, err := handleBlockSave(block.BlockType, block.ID, input.Content)
		if err != nil {
			sendAPIError(w, api_error_block_save_error, err, map[string]interface{}{
				"input": input,
			})
			return
		}
		block.Content = content
	}
	sendAPIJSONData(w, http.StatusOK, block)
}

// routeAdminDeleteBlock deletes a block and associated content
func routeAdminDeleteBlock(w http.ResponseWriter, r *http.Request) {
	blockID, blockIDErr := strconv.ParseInt(chi.URLParam(r, "blockID"), 10, 64)
	if blockIDErr != nil {
		sendAPIError(w, api_error_invalid_path, blockIDErr, map[string]string{})
		return
	}

	block, err := GetBlockByID(blockID)
	if err != nil {
		sendAPIError(w, api_error_block_not_found, err, map[string]interface{}{})
		return
	}
	err = handleBlockDelete(block.BlockType, blockID)
	if err != nil {
		sendAPIError(w, api_error_block_delete_err, err, map[string]string{})
		return
	}
	err = DeleteBlock(blockID)
	if err != nil {
		sendAPIError(w, api_error_block_delete_err, err, map[string]string{})
		return
	}
	sendAPIJSONData(w, http.StatusOK, map[string]bool{
		"deleted": true,
	})
}

// routeAdminLinkBlockAndModule links a module and a block
func routeAdminLinkBlockAndModule(w http.ResponseWriter, r *http.Request) {
	moduleID, moduleIDErr := strconv.ParseInt(chi.URLParam(r, "moduleID"), 10, 64)
	blockID, blockIDErr := strconv.ParseInt(chi.URLParam(r, "blockID"), 10, 64)
	order, orderErr := strconv.ParseInt(chi.URLParam(r, "order"), 10, 64)
	if blockIDErr != nil || moduleIDErr != nil || orderErr != nil {
		sendAPIError(w, api_error_invalid_path, errors.New("invalid path"), map[string]string{})
		return
	}

	// make sure the module and block both exist
	_, err := GetModuleByID(moduleID)
	if err != nil {
		sendAPIError(w, api_error_module_not_found, err, map[string]interface{}{})
		return
	}
	_, err = GetBlockByID(blockID)
	if err != nil {
		sendAPIError(w, api_error_block_not_found, err, map[string]interface{}{})
		return
	}

	err = LinkBlockAndModule(moduleID, blockID, order)
	if err != nil {
		sendAPIError(w, api_error_block_link_err, err, map[string]int64{
			"moduleID": moduleID,
			"blockID":  blockID,
			"order":    order,
		})
		return
	}
	sendAPIJSONData(w, http.StatusOK, map[string]bool{
		"linked": true,
	})
}

// routeAdminUnlinkBlockAndModule unlinks a module and a block
func routeAdminUnlinkBlockAndModule(w http.ResponseWriter, r *http.Request) {
	moduleID, moduleIDErr := strconv.ParseInt(chi.URLParam(r, "moduleID"), 10, 64)
	blockID, blockIDErr := strconv.ParseInt(chi.URLParam(r, "blockID"), 10, 64)
	if blockIDErr != nil || moduleIDErr != nil {
		sendAPIError(w, api_error_invalid_path, errors.New("invalid path"), map[string]string{})
		return
	}

	err := UnlinkBlockAndModule(moduleID, blockID)
	if err != nil {
		sendAPIError(w, api_error_block_unlink_err, err, map[string]int64{
			"moduleID": moduleID,
			"blockID":  blockID,
		})
		return
	}
	sendAPIJSONData(w, http.StatusOK, map[string]bool{
		"linked": false,
	})
}
