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
	matched := make(map[string]string)
	denomRoutes := rt.sortRoutes()
	for in, routes := range denomRoutes {
		for _, out := range routes {
			if !utils.StringInSlice(in, denomRoutes[out]) {
				denomIn, denomOut := out, in
				matched[denomIn] = denomOut
				newRoute := Route{
					InputDenom:  denomIn,
					OutputDenom: denomOut,
				}
				route := rt.getRoute(in, out)
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
					matched[denomIn] = denomOut
					denomIn = denomOut
				}
				newRoutes = append(newRoutes, RouteAdd{newRoute})
			}
		}
	}
	return
}

// search intermediary routes
func (rt RoutingTable) Fill() (newRoutes []RouteAdd) {
	matched := make(map[string]string)
	for _, route := range rt.Routes {
		if len(route.PoolRoute) > 1 {
			var poolRoute []PoolRoute
			denomIn := route.InputDenom
			for _, subroute := range route.PoolRoute {
				if denomIn == subroute.TokenOutDenom {
					continue
				}
				poolRoute = append(poolRoute, subroute)
				if rt.getRoute(denomIn, subroute.TokenOutDenom) == nil {
					newRoutes = append(newRoutes, RouteAdd{Route{
						InputDenom:  denomIn,
						OutputDenom: subroute.TokenOutDenom,
						PoolRoute:   poolRoute,
					}})
					matched[denomIn] = subroute.TokenOutDenom
				}
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

func (rt RoutingTable) getRoute(inDenom string, outDenom string) []PoolRoute {
	for _, route := range rt.Routes {
		if route.InputDenom == inDenom && route.OutputDenom == outDenom {
			return route.PoolRoute
		}
	}
	return nil
}
