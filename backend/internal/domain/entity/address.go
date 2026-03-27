package entity

// Address represents a physical address (value object, reusable across entities).
type Address struct {
	Street        string `json:"street"`
	City          string `json:"city"`
	StateOrRegion string `json:"state_or_region"`
	PostalCode    string `json:"postal_code,omitempty"`
	Country       string `json:"country"`
}

// NewAddress creates an Address with defaults.
func NewAddress(street, city, stateOrRegion, postalCode, country string) *Address {
	if country == "" {
		country = "Philippines"
	}
	return &Address{
		Street:        street,
		City:          city,
		StateOrRegion: stateOrRegion,
		PostalCode:    postalCode,
		Country:       country,
	}
}

// FullAddress returns the formatted address string.
func (a *Address) FullAddress() string {
	s := a.Street + ", " + a.City + ", " + a.StateOrRegion
	if a.PostalCode != "" {
		s += ", " + a.PostalCode
	}
	s += ", " + a.Country
	return s
}
