package api

// User is a person with a login that has permission to "do stuff". This is for researchers, site admins, etc and NOT participants
type User struct {
	ID        int64
	FirstName string
	LastName  string
	Email     string
	Password  string
}
