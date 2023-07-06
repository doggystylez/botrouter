package botrouter

type (
	RoutingTable struct {
		Routes []Route `json:"routes"`
	}

	PoolRoute struct {
		PoolID        string `json:"pool_id"`
		TokenOutDenom string `json:"token_out_denom"`
	}

	// contract message for new route
	RouteAdd struct {
		Route `json:"set_route"`
	}

	Route struct {
		InputDenom  string      `json:"input_denom"`
		OutputDenom string      `json:"output_denom"`
		PoolRoute   []PoolRoute `json:"pool_route"`
	}

	// general contract info
	ContractInfo []Info

	Info struct {
		Key   []byte
		Value []byte
	}
)
