// internal/models/service.go
package models

type Concentration struct {
	Mg int
	Ml int
}
type Service struct {
	ID            int
	Title         string
	Concentration Concentration
	AdultDose     int
	Description   string
	ImageKey      string
	VideoKey      string
	Likes         []int
	Status        string
}
