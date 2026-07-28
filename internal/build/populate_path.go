package build

import (
	"fmt"

	"github.com/DhimasYulian/hush/internal/query"
)

// PopulatePath holds the parsed components of a populate parameter path.
type PopulatePath struct {
	Relations  []string
	Option     string
	OptionPath []string
}

// ParsePopulatePath parses a populate parameter path into its components.
//
//	For example: ["populate", "author", "fields", "0"] → PopulatePath{
//	  Relations: ["author"], Option: "fields", OptionPath: ["0"]
//	}
func ParsePopulatePath(path query.Path) (PopulatePath, error) {
	if len(path) < 3 || path[0] != "populate" {
		return PopulatePath{}, query.PathError(query.ErrInvalidPopulate, path, "invalid populate path")
	}

	var result PopulatePath

	i := 1

	for i < len(path) {
		relation := path[i]
		if relation == "" {
			return PopulatePath{}, query.PathError(query.ErrInvalidPopulate, path, "empty relation")
		}

		result.Relations = append(result.Relations, relation)
		i++

		if i >= len(path) {
			return result, nil
		}

		switch path[i] {
		case "populate":
			i++

		case "fields", "sort", "filters":
			result.Option = path[i]
			result.OptionPath = append(result.OptionPath, path[i+1:]...)
			return result, nil

		default:
			return PopulatePath{}, query.PathError(query.ErrInvalidPopulate, path, fmt.Sprintf("unexpected segment %q", path[i]))
		}
	}

	return result, nil
}
