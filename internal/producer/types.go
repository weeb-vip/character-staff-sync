package producer

type DataType = string

const (
	// DataTypeImage represents an image data type
	DataTypeAnime     DataType = "Anime"
	DataTypeCharacter DataType = "Character"
	DataTypeStaff     DataType = "Staff"
)

type ImageSchema struct {
	Name string   `json:"name"`
	URL  string   `json:"url"`
	Type DataType `json:"type"`
}
type ImagePayload struct {
	Data ImageSchema `json:"data"`
}
