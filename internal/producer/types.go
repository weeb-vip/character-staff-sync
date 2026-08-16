package producer

type DataType = string

const (
	// DataTypeImage represents an image data type
	DataTypeAnime     DataType = "Anime"
	DataTypeCharacter DataType = "Character"
	DataTypeStaff     DataType = "Staff"
)

type ImageSchema struct {
	// ID is what image-sync keys the object by. Name is still sent so a
	// consumer that has not picked up the id yet keeps working.
	ID   string   `json:"id"`
	Name string   `json:"name"`
	URL  string   `json:"url"`
	Type DataType `json:"type"`
}
type ImagePayload struct {
	Data ImageSchema `json:"data"`
}
