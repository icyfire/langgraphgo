package graph

import (
	"fmt"
	"maps"
	"reflect"
)

// StateSchema defines the structure and update logic for the graph state with type safety.
type StateSchema[S any] interface {
	// Init returns the initial state.
	Init() S

	// Update merges the new state into the current state.
	Update(current, new S) (S, error)
}

// StructSchema implements StateSchema for struct-based states.
// It provides a simple and type-safe way to manage struct states.
//
// Example:
//
//	type MyState struct {
//	    Count int
//	    Logs []string
//	}
//
//	schema := graph.NewStructSchema(
//	    MyState{Count: 0},
//	    func(current, new MyState) (MyState, error) {
//	        // Merge logs (append)
//	        current.Logs = append(current.Logs, new.Logs...)
//	        // Add counts
//	        current.Count += new.Count
//	        return current, nil
//	    },
//	)
type StructSchema[S any] struct {
	InitialValue S
	MergeFunc    func(S, S) (S, error)
}

// NewStructSchema creates a new StructSchema with the given initial value and merge function.
// If merge function is nil, a default merge function will be used that overwrites non-zero fields.
func NewStructSchema[S any](initial S, merge func(S, S) (S, error)) *StructSchema[S] {
	if merge == nil {
		merge = DefaultStructMerge[S]
	}
	return &StructSchema[S]{
		InitialValue: initial,
		MergeFunc:    merge,
	}
}

// Init returns the initial state.
func (s *StructSchema[S]) Init() S {
	return s.InitialValue
}

// Update merges the new state into the current state using the merge function.
func (s *StructSchema[S]) Update(current S, new S) (S, error) {
	if s.MergeFunc != nil {
		return s.MergeFunc(current, new)
	}
	// Default: return new state
	return new, nil
}

// DefaultStructMerge provides a default merge function for struct states.
// It uses reflection to merge non-zero fields from new into current.
// This is a sensible default for most struct types.
func DefaultStructMerge[S any](current S, new S) (S, error) {
	// Create a zero-initialized result of type S
	resultType := reflect.TypeOf(current)
	resultValue := reflect.New(resultType).Elem()

	currentVal := reflect.ValueOf(current)
	newVal := reflect.ValueOf(new)

	// Check if S is a struct
	if currentVal.Kind() != reflect.Struct {
		// For non-struct types, just return new
		return new, nil
	}

	// Copy non-zero fields from current to result
	for i := 0; i < currentVal.NumField(); i++ {
		fieldCurrent := currentVal.Field(i)
		resultField := resultValue.Field(i)
		if !fieldCurrent.IsZero() && resultField.CanSet() {
			resultField.Set(fieldCurrent)
		}
	}

	// Copy non-zero fields from new to result (overwrites if non-zero)
	for i := 0; i < newVal.NumField(); i++ {
		fieldNew := newVal.Field(i)
		resultField := resultValue.Field(i)
		if !fieldNew.IsZero() && resultField.CanSet() {
			resultField.Set(fieldNew)
		}
	}

	return resultValue.Interface().(S), nil
}

// OverwriteStructMerge is a merge function that completely replaces the current state with the new state.
func OverwriteStructMerge[S any](current S, new S) (S, error) {
	return new, nil
}

// FieldMerger provides fine-grained control over how individual struct fields are merged.
type FieldMerger[S any] struct {
	InitialValue  S
	FieldMergeFns map[string]func(currentVal, newVal reflect.Value) reflect.Value
}

// NewFieldMerger creates a new FieldMerger with the given initial value.
func NewFieldMerger[S any](initial S) *FieldMerger[S] {
	return &FieldMerger[S]{
		InitialValue:  initial,
		FieldMergeFns: make(map[string]func(currentVal, newVal reflect.Value) reflect.Value),
	}
}

// RegisterFieldMerge registers a custom merge function for a specific field.
func (fm *FieldMerger[S]) RegisterFieldMerge(fieldName string, mergeFn func(currentVal, newVal reflect.Value) reflect.Value) {
	fm.FieldMergeFns[fieldName] = mergeFn
}

// Init returns the initial state.
func (fm *FieldMerger[S]) Init() S {
	return fm.InitialValue
}

// Update merges the new state into the current state using registered field merge functions.
func (fm *FieldMerger[S]) Update(current S, new S) (S, error) {
	// Create a zero-initialized result of type S
	resultType := reflect.TypeOf(current)
	resultValue := reflect.New(resultType).Elem()

	currentVal := reflect.ValueOf(current)
	newVal := reflect.ValueOf(new)

	if currentVal.Kind() != reflect.Struct {
		return new, fmt.Errorf("FieldMerger only works with struct types")
	}

	structType := currentVal.Type()

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		fieldName := field.Name

		currentFieldVal := currentVal.Field(i)
		newFieldVal := newVal.Field(i)
		resultFieldVal := resultValue.Field(i)

		// Check if there's a custom merge function for this field
		if mergeFn, ok := fm.FieldMergeFns[fieldName]; ok {
			if resultFieldVal.CanSet() {
				mergedVal := mergeFn(currentFieldVal, newFieldVal)
				resultFieldVal.Set(mergedVal)
			}
		} else {
			// Default: overwrite if new value is non-zero, otherwise keep current
			if !newFieldVal.IsZero() && resultFieldVal.CanSet() {
				resultFieldVal.Set(newFieldVal)
			} else if !currentFieldVal.IsZero() && resultFieldVal.CanSet() {
				resultFieldVal.Set(currentFieldVal)
			}
		}
	}

	return resultValue.Interface().(S), nil
}

// Common merge helpers for FieldMerger

// AppendSliceMerge appends new slice to current slice.
func AppendSliceMerge(current, new reflect.Value) reflect.Value {
	if current.Kind() != reflect.Slice || new.Kind() != reflect.Slice {
		return new
	}
	return reflect.AppendSlice(current, new)
}

// SumIntMerge adds two integer values.
func SumIntMerge(current, new reflect.Value) reflect.Value {
	if current.Kind() == reflect.Int && new.Kind() == reflect.Int {
		return reflect.ValueOf(current.Int() + new.Int()).Convert(current.Type())
	}
	return new
}

// OverwriteMerge always uses the new value.
func OverwriteMerge(current, new reflect.Value) reflect.Value {
	return new
}

// KeepCurrentMerge always keeps the current value (ignores new).
func KeepCurrentMerge(current, new reflect.Value) reflect.Value {
	return current
}

// MaxIntMerge takes the maximum of two integer values.
func MaxIntMerge(current, new reflect.Value) reflect.Value {
	if current.Kind() == reflect.Int && new.Kind() == reflect.Int {
		if current.Int() > new.Int() {
			return current
		}
	}
	return new
}

// MinIntMerge takes the minimum of two integer values.
func MinIntMerge(current, new reflect.Value) reflect.Value {
	if current.Kind() == reflect.Int && new.Kind() == reflect.Int {
		if current.Int() < new.Int() {
			return current
		}
	}
	return new
}

// Reducer defines how a state value should be updated.
// It takes the current value and the new value, and returns a merged value.
type Reducer func(current, new any) (any, error)

// MapSchema implements StateSchema for map[string]any.
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
func (s *MapSchema) Init() map[string]any {
	return make(map[string]any)
}

// sameValue returns true if a and b point to the same underlying object.
// This is needed because when a node modifies its input state and returns it,
// values in current and new maps may point to the same objects (e.g., slices).
func sameValue(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Use reflect to compare pointers for reference types
	va := reflect.ValueOf(a)
	vb := reflect.ValueOf(b)

	// For slices, check if they point to the same underlying array
	if va.Kind() == reflect.Slice && vb.Kind() == reflect.Slice {
		// Get data pointer using reflect.Value.Pointer()
		// For slices, Pointer() returns address of first element
		dataPtrA := va.Pointer()
		dataPtrB := vb.Pointer()
		// Also check lengths - if lengths differ, they're definitely different
		if va.Len() != vb.Len() {
			return false
		}
		return dataPtrA == dataPtrB
	}

	// For maps, check if they point to the same underlying hash table
	if va.Kind() == reflect.Map && vb.Kind() == reflect.Map {
		// For maps, Pointer() returns hash table address
		return va.Pointer() == vb.Pointer()
	}

	// For other types, just compare values
	return reflect.DeepEqual(a, b)
}

// Update merges the new map into the current map using registered reducers.
func (s *MapSchema) Update(current, new map[string]any) (map[string]any, error) {
	if current == nil {
		current = make(map[string]any)
	}

	// Check if current and new are the same map object
	// This can happen when a node modifies its input state in-place and returns it
	if sameValue(current, new) {
		// They're the same object, so new already contains all the updates
		return new, nil
	}

	// Create a copy of the current map to avoid mutating it directly
	result := make(map[string]any, len(current))
	maps.Copy(result, current)

	for k, v := range new {
		if reducer, ok := s.Reducers[k]; ok {
			// Use reducer
			currVal := result[k]
			// Check if current and new values are the same reference
			// This can happen when a node modifies its input state and returns it
			if sameValue(currVal, v) {
				// They're the same object - skip reducer to avoid self-append
				continue
			}
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
func OverwriteReducer(current, new any) (any, error) {
	return new, nil
}

// AppendReducer appends the new value to the current slice.
// It supports appending a slice to a slice, or a single element to a slice.
func AppendReducer(current, new any) (any, error) {
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
		if currVal.Type().Elem() != newVal.Type().Elem() {
			// Types don't match, convert both to []any
			result := make([]any, 0, currVal.Len()+newVal.Len())
			for i := 0; i < currVal.Len(); i++ {
				result = append(result, currVal.Index(i).Interface())
			}
			for i := 0; i < newVal.Len(); i++ {
				result = append(result, newVal.Index(i).Interface())
			}
			return result, nil
		}
		return reflect.AppendSlice(currVal, newVal).Interface(), nil
	}

	// Append single element
	return reflect.Append(currVal, newVal).Interface(), nil
}
