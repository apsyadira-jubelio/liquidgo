package liquid

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// StrainerTemplate is the computed class for the filters system.
// New filters are mixed into the strainer class which is then instantiated for each liquid template render run.
type StrainerTemplate struct {
	context         interface{ Context() interface{} }
	filterMethods   map[string]bool
	filterInstances map[string]interface{}
	filterOrder     []interface{} // Maintains registration order for method precedence
	strictFilters   bool
}

// StrainerTemplateClass represents a strainer template class that can have filters added.
type StrainerTemplateClass struct {
	filterMethods map[string]bool
}

// NewStrainerTemplateClass creates a new strainer template class.
func NewStrainerTemplateClass() *StrainerTemplateClass {
	return &StrainerTemplateClass{
		filterMethods: make(map[string]bool),
	}
}

// AddFilter adds a filter module to the strainer template class.
func (stc *StrainerTemplateClass) AddFilter(filter interface{}) error {
	// Get type of the filter
	filterType := reflect.TypeOf(filter)

	// Handle functions as filters (e.g., func(interface{}) interface{})
	if filterType.Kind() == reflect.Func {
		// For functions, we'll treat the function itself as a callable filter
		// We don't need to register method names since the function is the filter
		// This will be handled during invocation
		return nil
	}

	// Handle structs with methods as filters
	if filterType.Kind() != reflect.Ptr {
		return fmt.Errorf("filter must be a pointer to a struct or a function")
	}

	// Get methods from the filter
	for i := 0; i < filterType.NumMethod(); i++ {
		method := filterType.Method(i)
		// Only include exported methods
		if method.PkgPath == "" {
			stc.filterMethods[method.Name] = true
		}
	}

	return nil
}

// Invokable checks if a method name is invokable.
func (stc *StrainerTemplateClass) Invokable(method string) bool {
	return stc.filterMethods[method]
}

// FilterMethodNames returns all filter method names.
func (stc *StrainerTemplateClass) FilterMethodNames() []string {
	names := make([]string, 0, len(stc.filterMethods))
	for name := range stc.filterMethods {
		names = append(names, name)
	}
	return names
}

// NewStrainerTemplate creates a new strainer template instance.
func NewStrainerTemplate(class *StrainerTemplateClass, context interface{ Context() interface{} }, strictFilters bool) *StrainerTemplate {
	st := &StrainerTemplate{
		context:         context,
		filterMethods:   class.filterMethods,
		strictFilters:   strictFilters,
		filterInstances: make(map[string]interface{}),
		filterOrder:     make([]interface{}, 0),
	}

	// Always add StandardFilters as a base filter with context
	var ctx *Context
	if context != nil {
		if c, ok := context.Context().(*Context); ok {
			ctx = c
		}
	}
	sf := &StandardFilters{context: ctx}
	st.filterInstances["*liquid.StandardFilters"] = sf
	st.filterOrder = append(st.filterOrder, sf)

	return st
}

var errorInterface = reflect.TypeOf((*error)(nil)).Elem()

// ContextualFilter is implemented by filters that need the render context,
// as Ruby filters read @context (registers, environments). WithContext is called
// once per render and returns the instance to use for it; it must return the
// same pointer type as the registered filter.
type ContextualFilter interface {
	WithContext(ctx *Context) interface{}
}

// NewStrainerTemplateWithFilters creates a new strainer template instance with additional filters.
func NewStrainerTemplateWithFilters(class *StrainerTemplateClass, context interface{ Context() interface{} }, strictFilters bool, filters []interface{}) *StrainerTemplate {
	st := NewStrainerTemplate(class, context, strictFilters)

	var ctx *Context
	if context != nil {
		ctx, _ = context.Context().(*Context)
	}

	// Add additional filter instances
	for _, filter := range filters {
		if cf, ok := filter.(ContextualFilter); ok && ctx != nil {
			filter = cf.WithContext(ctx)
		}
		filterType := reflect.TypeOf(filter)
		if filterType.Kind() == reflect.Ptr {
			st.filterInstances[filterType.String()] = filter
			st.filterOrder = append(st.filterOrder, filter)
		}
	}

	return st
}

// Invoke invokes a filter method.
func (st *StrainerTemplate) Invoke(method string, args ...interface{}) (interface{}, error) {
	// Check if method is invokable (try both lowercase and capitalized)
	methodInvokable := st.filterMethods[method]
	if !methodInvokable && len(method) > 0 {
		// Try CamelCase version (converts snake_case to CamelCase for Go method names)
		// e.g., find_index -> FindIndex, sort_natural -> SortNatural
		camelMethod := snakeToCamelCase(method)
		methodInvokable = st.filterMethods[camelMethod]
		if methodInvokable {
			// Use CamelCase version for lookup
			method = camelMethod
		} else {
			// Try case-insensitive match for acronyms (e.g., StripHtml -> StripHTML)
			// This handles cases where the method uses uppercase acronyms like HTML, XML, etc.
			for registeredMethod := range st.filterMethods {
				if strings.EqualFold(registeredMethod, camelMethod) {
					methodInvokable = true
					method = registeredMethod
					break
				}
			}
			if !methodInvokable {
				// Fallback: try simple capitalization (for single-word filters)
				capitalizedMethod := strings.ToUpper(method[:1]) + method[1:]
				methodInvokable = st.filterMethods[capitalizedMethod]
				if methodInvokable {
					// Use capitalized version for lookup
					method = capitalizedMethod
				}
			}
		}
	}
	if !methodInvokable {
		// Before failing, try property access on the first argument
		// This enables patterns like: {{ posts | first | title }}
		if len(args) > 0 && args[0] != nil {
			// Try to access property using InvokeDropOn
			result := InvokeDropOn(args[0], method)
			if result != nil {
				return result, nil
			}
		}

		if st.strictFilters {
			return nil, NewUndefinedFilter("undefined filter " + method)
		}
		// In non-strict mode, return first arg
		if len(args) > 0 {
			return args[0], nil
		}
		return nil, nil
	}

	// Method is invokable - use reflection to find and call it on filter instances
	// Iterate in reverse order so later-registered filters take precedence
	for i := len(st.filterOrder) - 1; i >= 0; i-- {
		filterInstance := st.filterOrder[i]
		filterValue := reflect.ValueOf(filterInstance)

		// Look for the method - try both original case and capitalized version
		// Go method names are capitalized, but Liquid filter names are lowercase
		idx := filterMethodIndex(filterValue.Type(), method)
		if idx < 0 {
			continue
		}
		methodValue := filterValue.Method(idx)

		// Check if method signature matches (first arg is input, rest are filter args)
		methodType := methodValue.Type()
		if methodType.NumIn() < 1 {
			continue
		}

		// Prepare arguments
		// First arg is the input (from args[0])
		if len(args) == 0 {
			continue
		}

		// Check if method is variadic (has ...interface{} parameter)
		isVariadic := methodType.IsVariadic()
		numIn := methodType.NumIn()
		minRequired := numIn
		if isVariadic {
			minRequired = numIn - 1 // Variadic param is optional
		}

		// Build call arguments - convert all args to reflect.Value
		callArgs := make([]reflect.Value, len(args))
		for i := 0; i < len(args); i++ {
			if args[i] == nil {
				// For nil values, use zero value of the expected type
				var paramType reflect.Type
				if isVariadic && i >= minRequired {
					// For variadic params beyond fixed params, use element type
					paramType = methodType.In(numIn - 1).Elem()
				} else {
					paramType = methodType.In(i)
				}
				callArgs[i] = reflect.Zero(paramType)
			} else {
				callArgs[i] = reflect.ValueOf(args[i])
			}
		}

		// Pad missing required (non-variadic) arguments with zero values
		for i := len(callArgs); i < minRequired; i++ {
			paramType := methodType.In(i)
			callArgs = append(callArgs, reflect.Zero(paramType))
		}

		// Call the method - use Call for both variadic and non-variadic
		// Go's reflect.Call handles variadic functions automatically
		results := methodValue.Call(callArgs)
		// A filter that returns an error raises it, as a Ruby filter would;
		// rendering turns it into "Liquid error: ...".
		if n := len(results); n > 1 && results[n-1].Type().Implements(errorInterface) && !results[n-1].IsNil() {
			panic(results[n-1].Interface().(error))
		}
		if len(results) > 0 {
			return results[0].Interface(), nil
		}
		return nil, nil
	}

	// Method not found in any filter - this shouldn't happen if filterMethods is correct
	// but handle gracefully
	if len(args) > 0 {
		return args[0], nil
	}
	return nil, nil
}

// snakeToCamelCase converts snake_case to CamelCase.
// e.g., find_index -> FindIndex, sort_natural -> SortNatural, strip_html -> StripHTML
func snakeToCamelCase(s string) string {
	if s == "" {
		return ""
	}

	// Split by underscore
	parts := strings.Split(s, "_")

	// Common acronyms that should be uppercase
	acronyms := map[string]string{
		"html": "HTML",
		"xml":  "XML",
		"json": "JSON",
		"url":  "URL",
		"id":   "ID",
		"api":  "API",
		"css":  "CSS",
		"js":   "JS",
	}

	// Capitalize each part
	for i, part := range parts {
		if len(part) > 0 {
			lowerPart := strings.ToLower(part)
			if acronym, ok := acronyms[lowerPart]; ok {
				// Use uppercase acronym
				parts[i] = acronym
			} else {
				// Capitalize first letter
				parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
			}
		}
	}

	return strings.Join(parts, "")
}

type filterMethodKey struct {
	t    reflect.Type
	name string
}

// filterMethodIndexCache memoises filterMethodIndex. Keys are registered
// filter names per filter type, so it stays bounded. Type.MethodByName
// allocates a func type on every call, which dominated filter invocation.
var filterMethodIndexCache sync.Map // map[filterMethodKey]int

// filterMethodIndex finds method (or its capitalized form) on t, -1 if absent.
func filterMethodIndex(t reflect.Type, method string) int {
	key := filterMethodKey{t, method}
	if cached, ok := filterMethodIndexCache.Load(key); ok {
		return cached.(int)
	}
	idx := -1
	if m, ok := t.MethodByName(method); ok {
		idx = m.Index
	} else if len(method) > 0 {
		// Try capitalized version (first letter uppercase)
		if m, ok := t.MethodByName(strings.ToUpper(method[:1]) + method[1:]); ok {
			idx = m.Index
		}
	}
	filterMethodIndexCache.Store(key, idx)
	return idx
}
