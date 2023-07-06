package botrouter

type (
	//
	RoutingTable struct {
		Routes []Route `json:"routes"`
	}

	Route struct {
		DenomIn   string      `json:"denom_in,omitempty"`
		DenomOut  string      `json:"denom_out,omitempty"`
		PoolRoute []PoolRoute `json:"pool_route,omitempty"`
	}

	// contract message for new route
	RouteAdd struct {
		SetRoute `json:"set_route"`
	}

	SetRoute struct {
		InputDenom  string      `json:"input_denom"`
		OutputDenom string      `json:"output_denom"`
		PoolRoute   []PoolRoute `json:"pool_route"`
	}

	// route contract query response
	RouteRes struct {
		PoolRoute []PoolRoute `json:"pool_route"`
	}

	PoolRoute struct {
		PoolID        string `json:"pool_id"`
		TokenOutDenom string `json:"token_out_denom"`
	}

	// general contract info
	ContractInfo []Info

	Info struct {
		Key   []byte
		Value []byte
	}
)
