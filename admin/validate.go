package admin

import (
	"fmt"

	"github.com/edfungus/conduction/messenger"
	"github.com/edfungus/conduction/storage"
)

var (
	ErrFlowMissingName        error = fmt.Errorf("Flow is missing field: name")
	ErrFlowMissingDescription error = fmt.Errorf("Flow is missing field: description")
	ErrFlowMissingPath        error = fmt.Errorf("Flow is missing field: path")
	ErrPathMissingRoute       error = fmt.Errorf("Path is missing field: route")
	ErrPathMissingType        error = fmt.Errorf("Path is missing field: type")
)

func validateFlow(flow storage.Flow) error {
	if flow.Name == "" {
		return ErrFlowMissingName
	}
	if flow.Description == "" {
		return ErrFlowMissingDescription
	}
	if flow.Path == nil {
		return ErrFlowMissingPath
	}
	return validatePath(*flow.Path)
}

func validatePath(path messenger.Path) error {
	if path.Route == "" {
		return ErrPathMissingRoute
	}
	if path.Type == "" {
		return ErrPathMissingType
	}
	return nil
}
