package pulsar_anime_character_postgres_processor

type Action = string

const (
	CreateAction Action = "create"
	UpdateAction Action = "update"
	DeleteAction Action = "delete"
)

type Schema struct {
	Id            string  `json:"id"`
	AnimeID       *string `json:"anime_id"`
	Name          *string `json:"name"`
	Role          *string `json:"role"`
	Birthday      *string `json:"birthday"`
	Zodiac        *string `json:"zodiac"`
	Gender        *string `json:"gender"`
	Race          *string `json:"race"`
	Height        *string `json:"height"`
	Weight        *string `json:"weight"`
	Title         *string `json:"title"`
	MartialStatus *string `json:"martial_status"`
	Summary       *string `json:"summary"`
	Image         *string `json:"image"`
	CreatedAt     *int64  `json:"created_at"`
	UpdatedAt     *int64  `json:"updated_at"`
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
