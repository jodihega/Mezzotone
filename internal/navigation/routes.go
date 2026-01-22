package navigation

type Route int

const (
	RouteMainMenu Route = iota
	RouteConvertImageMenu
)

func (r Route) String() string {
	switch r {
	case RouteMainMenu:
		return "Main Menu"
	case RouteConvertImageMenu:
		return "Convert Image Menu"
	}
	return "Screen not found"
}
