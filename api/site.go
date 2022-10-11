package api

// Site is an available location or installation
type Site struct {
	ID          int64
	ShortName   string
	Name        string
	Description string
	Domain      string
	Status      string // pending, active, disabled
}
