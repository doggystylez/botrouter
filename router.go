package botrouter

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/doggystylez/go-utils"
)

func Parse(c ContractInfo) (routes RoutingTable, err error) {
	for _, info := range c {
		var route Route
		str := regexp.MustCompile("[^a-zA-Z0-9_/]+").ReplaceAllString(string(info.Key), "")
		if strings.HasPrefix(str, "routing_table") {
			if strings.HasPrefix(str, "routing_tableD") {
				route = Route{
					InputDenom:  str[14:82],
					OutputDenom: str[82:],
				}
			} else if strings.HasPrefix(str, "routing_tableuosmo") {
				route = Route{
					InputDenom:  "uosmo",
					OutputDenom: str[18:],
				}
			}
			poolRoutes := []PoolRoute{}
			err = json.Unmarshal(info.Value, &poolRoutes)
			if err != nil {
				return
			}
			route.PoolRoute = append(route.PoolRoute, poolRoutes...)
			routes.Routes = append(routes.Routes, route)
		}
	}
	return
}

// search all routes in reverse
func (rt RoutingTable) Reverse() (newRoutes []RouteAdd) {
	denomRoutes := rt.sortRoutes()
	for in, routes := range denomRoutes {
		for _, out := range routes {
			if !utils.StringInSlice(in, denomRoutes[out]) {
				denomIn, denomOut := out, in
				newRoute := Route{
					InputDenom:  denomIn,
					OutputDenom: denomOut,
				}
				route := rt.GetRoute(in, out)
				var pools, denoms []string
				skip := true
				for i := len(route) - 1; i >= 0; i-- {
					pools = append(pools, route[i].PoolID)
					if !skip {
						denoms = append(denoms, route[i].TokenOutDenom)
					}
					skip = false
				}
				denoms = append(denoms, in)
				for i, denomOut := range denoms {
					newRoute.PoolRoute = append(newRoute.PoolRoute, PoolRoute{
						PoolID:        pools[i],
						TokenOutDenom: denomOut,
					})
				}
				newRoutes = append(newRoutes, RouteAdd{newRoute})
			}
		}
	}
	return
}

// search intermediary routes
func (rt RoutingTable) Fill() (newRoutes []RouteAdd) {
	for _, route := range rt.Routes {
		if len(route.PoolRoute) > 1 {
			var poolRoute []PoolRoute
			denomIn := route.InputDenom
			for _, subroute := range route.PoolRoute {
				if denomIn == subroute.TokenOutDenom {
					continue
				}
				poolRoute = append(poolRoute, subroute)
				if rt.GetRoute(denomIn, subroute.TokenOutDenom) == nil {
					newRoutes = append(newRoutes, RouteAdd{Route{
						InputDenom:  denomIn,
						OutputDenom: subroute.TokenOutDenom,
						PoolRoute:   poolRoute,
					}})
				}
			}
		}
	}
	return
}

// connect routes with common tokens
func (rt RoutingTable) Connect() (newRoutes []RouteAdd) {
	denomRoutes := rt.sortRoutes()
	var denoms = make([]string, len(denomRoutes))
	for in := range denomRoutes {
		denoms = append(denoms, in)
	}
	for _, input := range denoms {
		outputs := rt.GetRoutesByDenom(input)
		for _, output := range denoms {
			if input == output {
				continue
			}
			if utils.StringInSlice(output, outputs) {
				continue
			}
			var connectRoutes [][]PoolRoute
			for _, connector := range denomRoutes[output] {
				if utils.StringInSlice(connector, outputs) {
					route := rt.GetRoute(input, connector)
					route = append(route, rt.GetRoute(connector, output)...)
					connectRoutes = append(connectRoutes, route)
				}
			}
			if len(connectRoutes) != 0 {
				var shortest []PoolRoute
				for _, route := range connectRoutes {
					if len(shortest) == 0 {
						shortest = route
					}
					if len(route) < len(shortest) {
						shortest = route
					}
				}
				newRoutes = append(newRoutes, RouteAdd{Route{
					InputDenom:  input,
					OutputDenom: output,
					PoolRoute:   shortest,
				}})
			}
		}
	}
	return
}

func (rt RoutingTable) GetRoutesByDenom(inDenom string) []string {
	return rt.sortRoutes()[inDenom]
}

func (rt RoutingTable) sortRoutes() map[string][]string {
	sortedRoutes := make(map[string][]string)
	for _, route := range rt.Routes {
		sortedRoutes[route.InputDenom] = append(sortedRoutes[route.InputDenom], route.OutputDenom)
	}
	return sortedRoutes
}

func (rt RoutingTable) GetRoute(inDenom string, outDenom string) []PoolRoute {
	for _, route := range rt.Routes {
		if route.InputDenom == inDenom && route.OutputDenom == outDenom {
			return route.PoolRoute
		}
	}
	return nil
}
