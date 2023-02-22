package api

// holds helper structs and funcs that don't have a home elsewhere

// CountReturn is a helpers for DB calls that just need a count
type CountReturn struct {
	Count int64 `json:"count" db:"count"`
}
