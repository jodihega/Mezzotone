package app

type Route int

const (
	RouteMainMenu Route = iota
)

func (r Route) String() string {
	switch r {
	case RouteMainMenu:
		return "Main Menu"
	}
	return "Screen not found"
}
