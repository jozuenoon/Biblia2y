package models

type User struct {
	ID        string `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Timezone  int    `json:"timezone,omitempty"`
}
