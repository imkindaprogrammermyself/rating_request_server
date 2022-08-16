package dat

type ExpectedValue struct {
	AverageDamageDealt float64 `json:"average_damage_dealt"`
	AverageFrags       float64 `json:"average_frags"`
	WinRate            float64 `json:"win_rate"`
}

type ExpectedValues struct {
	Time int                    `json:"time"`
	Data map[uint64]interface{} `json:"data"`
}

type PlayerShipsStatsMeta struct {
	Count  int         `json:"count"`
	Hidden interface{} `json:"hidden"`
}

type PlayerShipStats struct {
	DamageDealt uint64 `json:"damage_dealt"`
	Wins        int    `json:"wins"`
	Frags       int    `json:"frags"`
	Battles     int    `json:"battles"`
}

type PlayerShipInfo struct {
	ShipId  uint64          `json:"ship_id"`
	PvpSolo PlayerShipStats `json:"pvp_solo"`
}

type ApiError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    int    `json:"code"`
	Value   string `json:"value"`
}

type PlayerShipsStats struct {
	Status string                      `json:"status"`
	Meta   PlayerShipsStatsMeta        `json:"meta"`
	Data   map[uint64][]PlayerShipInfo `json:"data"`
	Error  ApiError                    `json:"error"`
}

type PlayerRating struct {
	ID     uint64  `json:"id"`
	Rating float64 `json:"rating"`
}

type PlayerRatings []PlayerRating

type PlayerInfo struct {
	ID        uint64 `json:"id"`
	Realm     string `json:"realm"`
	AccountID uint64 `json:"account_id"`
}

type RequestPayload []PlayerInfo
