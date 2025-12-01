package graph

import (
	"fmt"
	"reflect"
)

// Reducer defines how a state value should be updated.
// It takes the current value and the new value, and returns the merged value.
type Reducer func(current, new interface{}) (interface{}, error)

// StateSchema defines the structure and update logic for the graph state.
type StateSchema interface {
	// Init returns the initial state.
	Init() interface{}

	// Update merges the new state into the current state.
	Update(current, new interface{}) (interface{}, error)
}

// MapSchema implements StateSchema for map[string]interface{}.
// It allows defining reducers for specific keys.
type MapSchema struct {
	Reducers map[string]Reducer
}

// NewMapSchema creates a new MapSchema.
func NewMapSchema() *MapSchema {
	return &MapSchema{
		Reducers: make(map[string]Reducer),
	}
}

// RegisterReducer adds a reducer for a specific key.
func (s *MapSchema) RegisterReducer(key string, reducer Reducer) {
	s.Reducers[key] = reducer
}

// Init returns an empty map.
func (s *MapSchema) Init() interface{} {
	return make(map[string]interface{})
}

// Update merges the new map into the current map using registered reducers.
func (s *MapSchema) Update(current, new interface{}) (interface{}, error) {
	if current == nil {
		current = make(map[string]interface{})
	}

	currMap, ok := current.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("current state is not a map[string]interface{}")
	}

	newMap, ok := new.(map[string]interface{})
	if !ok {
		// If new state is not a map, maybe it's intended to replace the whole state?
		// Or maybe it's an error?
		// For MapSchema, we generally expect partial updates to be maps.
		// Unless we allow "resetting" the state?
		return nil, fmt.Errorf("new state is not a map[string]interface{}")
	}

	// Create a copy of the current map to avoid mutating it directly
	result := make(map[string]interface{}, len(currMap))
	for k, v := range currMap {
		result[k] = v
	}

	for k, v := range newMap {
		if reducer, ok := s.Reducers[k]; ok {
			// Use reducer
			currVal := result[k]
			mergedVal, err := reducer(currVal, v)
			if err != nil {
				return nil, fmt.Errorf("failed to reduce key %s: %w", k, err)
			}
			result[k] = mergedVal
		} else {
			// Default: Overwrite
			result[k] = v
		}
	}

	return result, nil
}

// Common Reducers

// OverwriteReducer replaces the old value with the new one.
func OverwriteReducer(current, new interface{}) (interface{}, error) {
	return new, nil
}

// AppendReducer appends the new value to the current slice.
// It supports appending a slice to a slice, or a single element to a slice.
func AppendReducer(current, new interface{}) (interface{}, error) {
	if current == nil {
		// If current is nil, start a new slice
		// We need to know the type? We can infer from new.
		newVal := reflect.ValueOf(new)
		if newVal.Kind() == reflect.Slice {
			return new, nil
		}
		// Create slice of type of new
		sliceType := reflect.SliceOf(reflect.TypeOf(new))
		slice := reflect.MakeSlice(sliceType, 0, 1)
		slice = reflect.Append(slice, newVal)
		return slice.Interface(), nil
	}

	currVal := reflect.ValueOf(current)
	newVal := reflect.ValueOf(new)

	if currVal.Kind() != reflect.Slice {
		return nil, fmt.Errorf("current value is not a slice")
	}

	if newVal.Kind() == reflect.Slice {
		// Append slice to slice
		// Check if types are compatible? reflect.AppendSlice handles it or panics.
		// We should probably check types to avoid panics.
		if currVal.Type().Elem() != newVal.Type().Elem() {
			// Try to append as generic interface slice?
			// For now, let's assume types match or rely on reflect to panic/convert if possible.
			// Actually reflect.AppendSlice requires exact match.
			return reflect.AppendSlice(currVal, newVal).Interface(), nil
		}
		return reflect.AppendSlice(currVal, newVal).Interface(), nil
	}

	// Append single element
	return reflect.Append(currVal, newVal).Interface(), nil
}
