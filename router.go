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
					DenomIn:  str[14:82],
					DenomOut: str[82:],
				}
			} else if strings.HasPrefix(str, "routing_tableuosmo") {
				route = Route{
					DenomIn:  "uosmo",
					DenomOut: str[18:],
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

// search all routes in reverse, including intermediary routes
func ReverseAndFill(routingTable RoutingTable) (newRoutes []RouteAdd, err error) {
	matched := make(map[string]string)
	denomRoutes := sortRoutes(routingTable)
	for in, routes := range denomRoutes {
		for _, out := range routes {
			if !utils.StringInSlice(in, denomRoutes[out]) {
				denomIn, denomOut := out, in
				matched[denomIn] = denomOut
				newRoute := SetRoute{
					InputDenom:  denomIn,
					OutputDenom: denomOut,
				}
				route := getRoutes(in, out, routingTable)
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
					if !utils.StringInSlice(denomOut, denomRoutes[denomIn]) && matched[denomIn] != denomOut {
						newRoutes = append(newRoutes, RouteAdd{SetRoute{
							InputDenom:  denomIn,
							OutputDenom: denomOut,
							PoolRoute: []PoolRoute{{
								PoolID:        pools[i],
								TokenOutDenom: denomOut,
							}}},
						})
					}
					newRoute.PoolRoute = append(newRoute.PoolRoute, PoolRoute{
						PoolID:        pools[i],
						TokenOutDenom: denomOut,
					})
					matched[denomIn] = denomOut
					denomIn = denomOut
				}
				newRoutes = append(newRoutes, RouteAdd{newRoute})
			}
		}
	}
	return
}

func GetRoutesByDenom(inDenom string, routingTable RoutingTable) (outDenoms []string) {
	return sortRoutes(routingTable)[inDenom]
}

func sortRoutes(routingTable RoutingTable) (sortedRoutes map[string][]string) {
	sortedRoutes = make(map[string][]string)
	for _, route := range routingTable.Routes {
		sortedRoutes[route.DenomIn] = append(sortedRoutes[route.DenomIn], route.DenomOut)
	}
	return
}

func getRoutes(inDenom string, outDenom string, routingTable RoutingTable) []PoolRoute {
	for _, route := range routingTable.Routes {
		if route.DenomIn == inDenom && route.DenomOut == outDenom {
			return route.PoolRoute
		}
	}
	return []PoolRoute{}
}
