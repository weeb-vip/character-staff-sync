package staff_processor

type Action = string

const (
	CreateAction Action = "create"
	UpdateAction Action = "update"
	DeleteAction Action = "delete"
)

type Schema struct {
	Id         string  `json:"id"`
	Language   *string `json:"language"`
	GivenName  *string `json:"given_name"`
	FamilyName *string `json:"family_name"`
	Image      *string `json:"image"`
	Birthday   *string `json:"birthday"`
	BirthPlace *string `json:"birth_place"`
	BloodType  *string `json:"blood_type"`
	Hobbies    *string `json:"hobbies"`
	Summary    *string `json:"summary"`
	CreatedAt  *int64  `json:"created_at"`
	UpdatedAt  *int64  `json:"updated_at"`
}

type Source struct {
	Version   string      `json:"version"`
	Connector string      `json:"connector"`
	Name      string      `json:"name"`
	TsMs      int64       `json:"ts_ms"`
	Snapshot  string      `json:"snapshot"`
	Db        string      `json:"db"`
	Sequence  string      `json:"sequence"`
	Schema    string      `json:"schema"`
	Table     string      `json:"table"`
	TxId      int         `json:"txId"`
	Lsn       int         `json:"lsn"`
	Xmin      interface{} `json:"xmin"`
}

type Payload struct {
	Before *Schema `json:"before"`
	After  *Schema `json:"after"`
	Source Source  `json:"source"`
}

type ProducerPayload struct {
	Action string  `json:"action"`
	Data   *Schema `json:"data"`
}
