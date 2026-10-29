package liquid

import (
	"reflect"
	"strings"
)

// MethodLiteral represents a method literal (blank, empty).
type MethodLiteral struct {
	MethodName string
	ToString   string
}

var (
	conditionMethodLiterals = map[string]*MethodLiteral{
		"blank": {MethodName: "blank", ToString: ""},
		"empty": {MethodName: "empty", ToString: ""},
	}
)

// ConditionOperator is a function type for condition operators.
type ConditionOperator func(cond *Condition, left, right interface{}) (bool, error)

// Condition represents a Liquid condition.
type Condition struct {
	left           interface{}
	right          interface{}
	attachment     interface{}
	childCondition *Condition
	operator       string
	childRelation  string
}

// NewCondition creates a new Condition.
func NewCondition(left interface{}, operator string, right interface{}) *Condition {
	return &Condition{
		left:     left,
		operator: operator,
		right:    right,
	}
}

// Left returns the left side of the condition.
func (c *Condition) Left() interface{} {
	return c.left
}

// Operator returns the operator.
func (c *Condition) Operator() string {
	return c.operator
}

// Right returns the right side of the condition.
func (c *Condition) Right() interface{} {
	return c.right
}

// ChildCondition returns the child condition.
func (c *Condition) ChildCondition() *Condition {
	return c.childCondition
}

// Attachment returns the attachment.
func (c *Condition) Attachment() interface{} {
	return c.attachment
}

// Or chains this condition with another using OR.
func (c *Condition) Or(condition *Condition) {
	c.childRelation = "or"
	c.childCondition = condition
}

// And chains this condition with another using AND.
func (c *Condition) And(condition *Condition) {
	c.childRelation = "and"
	c.childCondition = condition
}

// Attach attaches an attachment to this condition.
func (c *Condition) Attach(attachment interface{}) {
	c.attachment = attachment
}

// Else returns true if this is an else condition.
func (c *Condition) Else() bool {
	return false
}

// ConditionContext provides context for evaluating conditions.
type ConditionContext interface {
	Evaluate(expr interface{}) interface{}
}

// Evaluate evaluates the condition in the given context.
func (c *Condition) Evaluate(context ConditionContext) (bool, error) {
	condition := c
	var result bool
	var err error

	for {
		result, err = c.interpretCondition(condition.left, condition.right, condition.operator, context)
		if err != nil {
			return false, err
		}

		resultVal := ToLiquidValue(result)
		shouldContinue := false
		switch condition.childRelation {
		case "or":
			if !IsTruthy(resultVal) {
				shouldContinue = true
			}
		case "and":
			if IsTruthy(resultVal) {
				shouldContinue = true
			}
		}

		if !shouldContinue || condition.childCondition == nil {
			break
		}
		condition = condition.childCondition
	}

	return result, nil
}

func (c *Condition) interpretCondition(left, right interface{}, op string, context ConditionContext) (bool, error) {
	// If operator is empty, just evaluate the left side
	if op == "" {
		result := context.Evaluate(left)
		resultVal := ToLiquidValue(result)
		// Convert to bool
		if !IsTruthy(resultVal) {
			return false, nil
		}
		return true, nil
	}

	leftVal := ToLiquidValue(context.Evaluate(left))
	rightVal := ToLiquidValue(context.Evaluate(right))

	operator := getConditionOperator(op)
	if operator == nil {
		return false, NewArgumentError("Unknown operator " + op)
	}

	return operator(c, leftVal, rightVal)
}

func getConditionOperator(op string) ConditionOperator {
	switch op {
	case "==":
		return func(cond *Condition, left, right interface{}) (bool, error) {
			return cond.equalVariables(left, right), nil
		}
	case "!=", "<>":
		return func(cond *Condition, left, right interface{}) (bool, error) {
			return !cond.equalVariables(left, right), nil
		}
	case "<":
		return func(_ *Condition, left, right interface{}) (bool, error) {
			// Handle nil comparisons: all comparisons with nil evaluate to false (matching Shopify Liquid)
			if left == nil || right == nil {
				return false, nil
			}
			res, err := compareValues(left, right)
			if err != nil {
				return false, err
			}
			return res < 0, nil
		}
	case ">":
		return func(_ *Condition, left, right interface{}) (bool, error) {
			// Handle nil comparisons: all comparisons with nil evaluate to false (matching Shopify Liquid)
			if left == nil || right == nil {
				return false, nil
			}
			res, err := compareValues(left, right)
			if err != nil {
				return false, err
			}
			return res > 0, nil
		}
	case ">=":
		return func(_ *Condition, left, right interface{}) (bool, error) {
			// Handle nil comparisons: all comparisons with nil evaluate to false (matching Shopify Liquid)
			if left == nil || right == nil {
				return false, nil
			}
			res, err := compareValues(left, right)
			if err != nil {
				return false, err
			}
			return res >= 0, nil
		}
	case "<=":
		return func(_ *Condition, left, right interface{}) (bool, error) {
			// Handle nil comparisons: all comparisons with nil evaluate to false (matching Shopify Liquid)
			if left == nil || right == nil {
				return false, nil
			}
			res, err := compareValues(left, right)
			if err != nil {
				return false, err
			}
			return res <= 0, nil
		}
	case "contains":
		return func(_ *Condition, left, right interface{}) (bool, error) {
			return containsOperator(left, right), nil
		}
	default:
		return nil
	}
}

func (c *Condition) equalVariables(left, right interface{}) bool {
	if ml, ok := left.(*MethodLiteral); ok {
		return checkMethodLiteral(ml, right)
	}

	if ml, ok := right.(*MethodLiteral); ok {
		return checkMethodLiteral(ml, left)
	}

	return rubyEqual(left, right)
}

// rubyEqual is Ruby's ==: numerics compare across Integer/Float (1 == 1.0),
// arrays and hashes compare structurally, everything else (drops, pointers)
// by Go == when comparable. Never panics on uncomparable values.
func rubyEqual(left, right interface{}) bool {
	if lf, lok := numericValue(left); lok {
		if rf, rok := numericValue(right); rok {
			return lf == rf
		}
	}
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	lv, rv := reflect.ValueOf(left), reflect.ValueOf(right)
	lk, rk := lv.Kind(), rv.Kind()
	switch {
	case (lk == reflect.Slice || lk == reflect.Array) && (rk == reflect.Slice || rk == reflect.Array):
		if lv.Len() != rv.Len() {
			return false
		}
		for i := 0; i < lv.Len(); i++ {
			if !rubyEqual(lv.Index(i).Interface(), rv.Index(i).Interface()) {
				return false
			}
		}
		return true
	case lk == reflect.Map && rk == reflect.Map:
		if lv.Len() != rv.Len() || lv.Type().Key() != rv.Type().Key() {
			return false
		}
		iter := lv.MapRange()
		for iter.Next() {
			r := rv.MapIndex(iter.Key())
			if !r.IsValid() || !rubyEqual(iter.Value().Interface(), r.Interface()) {
				return false
			}
		}
		return true
	}
	if lv.Comparable() && rv.Comparable() {
		return left == right
	}
	return reflect.DeepEqual(left, right)
}

// isBlank is ActiveSupport's blank?, which Shopify's `blank` literal uses:
// nil, false, whitespace-only strings and empty arrays/hashes.
func isBlank(obj interface{}) bool {
	switch v := obj.(type) {
	case nil:
		return true
	case bool:
		return !v
	case string:
		return strings.TrimSpace(v) == ""
	}
	rv := reflect.ValueOf(obj)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map:
		return rv.Len() == 0
	}
	return false
}

// IsTruthy reports Ruby truthiness: only nil and false are falsy. Empty
// strings, 0 and empty arrays are all truthy in Ruby, unlike in many languages.
func IsTruthy(v interface{}) bool {
	if v == nil {
		return false
	}
	b, isBool := v.(bool)
	return !isBool || b
}

// numericValue returns the float value of any Go numeric type.
func numericValue(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case int:
		return float64(n), true
	case int8:
		return float64(n), true
	case int16:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint:
		return float64(n), true
	case uint8:
		return float64(n), true
	case uint16:
		return float64(n), true
	case uint32:
		return float64(n), true
	case uint64:
		return float64(n), true
	case float32:
		return float64(n), true
	case float64:
		return n, true
	}
	return 0, false
}

func checkMethodLiteral(ml *MethodLiteral, obj interface{}) bool {
	switch ml.MethodName {
	case "blank":
		return isBlank(obj)
	case "empty":
		if str, ok := obj.(string); ok {
			return str == ""
		}
		if arr, ok := obj.([]interface{}); ok {
			return len(arr) == 0
		}
		if m, ok := obj.(map[string]interface{}); ok {
			return len(m) == 0
		}
		// Reflection fallback for typed slices/arrays/maps
		// This matches Ruby's duck-typing behavior: objects respond to .empty?
		if obj != nil {
			v := reflect.ValueOf(obj)
			kind := v.Kind()
			if kind == reflect.Slice || kind == reflect.Array || kind == reflect.Map {
				return v.Len() == 0
			}
		}
		// Ruby: nil and other values without empty? never equal `empty`.
		return false
	default:
		return false
	}
}

func compareValues(left, right interface{}) (int, error) {
	// Handle nil comparisons gracefully (matching Shopify Liquid behavior)
	// When comparing nil with a number, the comparison should evaluate to false
	// To achieve this, we return -1 for nil comparisons, which makes:
	// - "nil > number" → res > 0 → -1 > 0 → false ✓
	// - "number > nil" → res > 0 → -1 > 0 → false ✓
	// - "nil < number" → res < 0 → -1 < 0 → true (but condition will be false due to nil falsiness)
	// - "number < nil" → res < 0 → -1 < 0 → true (but condition will be false due to nil falsiness)
	if left == nil && right != nil {
		// Check if right is a number
		_, rightOk := toNumber(right)
		if rightOk {
			return -1, nil
		}
	}
	if right == nil && left != nil {
		// Check if left is a number
		_, leftOk := toNumber(left)
		if leftOk {
			return -1, nil
		}
	}
	if left == nil && right == nil {
		// Both are nil, they're equal
		return 0, nil
	}

	// Simple numeric comparison
	leftNum, leftOk := toNumber(left)
	rightNum, rightOk := toNumber(right)
	if leftOk && rightOk {
		if leftNum < rightNum {
			return -1, nil
		} else if leftNum > rightNum {
			return 1, nil
		}
		return 0, nil
	}

	// If one is a number and the other is not (and neither is nil), it's a type mismatch for comparison
	if leftOk || rightOk {
		// Format types nicely for error message
		leftType := "nil"
		if left != nil {
			leftType = reflect.TypeOf(left).Name()
			if leftType == "" {
				leftType = reflect.TypeOf(left).String()
			}
		}
		rightType := "nil"
		if right != nil {
			rightType = reflect.TypeOf(right).Name()
			if rightType == "" {
				rightType = reflect.TypeOf(right).String()
			}
		}
		return 0, NewArgumentError("comparison of " + leftType + " with " + rightType + " failed")
	}

	// String comparison
	leftStr := ToS(left, nil)
	rightStr := ToS(right, nil)
	if leftStr < rightStr {
		return -1, nil
	} else if leftStr > rightStr {
		return 1, nil
	}
	return 0, nil
}

func toNumber(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case float64:
		return n, true
	}
	return 0, false
}

// itemContains reports whether an array element satisfies `contains`.
//
// Ruby's Array#include? is exact equality, so a string element must match
// exactly: ["about-us"] contains "about" is false. Non-string elements (structs
// and drops) keep this package's substring match over their rendered form,
// which lets `posts contains "Alice"` find a field value.
func itemContains(item interface{}, rightStr string) bool {
	itemStr := ToS(item, nil)
	if _, isString := item.(string); isString {
		return itemStr == rightStr
	}
	return itemStr == rightStr || strings.Contains(itemStr, rightStr)
}

func containsOperator(left, right interface{}) bool {
	if left == nil || right == nil {
		return false
	}

	rightStr := ToS(right, nil)

	// Check if left is a string
	if leftStr, ok := left.(string); ok {
		return strings.Contains(leftStr, rightStr)
	}

	// Check if left is a slice/array
	if arr, ok := left.([]interface{}); ok {
		for _, item := range arr {
			if itemContains(item, rightStr) {
				return true
			}
		}
	} else if left != nil {
		// Reflection fallback for typed slices ([]BlogPost, []string, []int, etc.)
		// This matches Ruby's duck-typing behavior: arrays respond to include?
		v := reflect.ValueOf(left)
		if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
			for i := 0; i < v.Len(); i++ {
				if itemContains(v.Index(i).Interface(), rightStr) {
					return true
				}
			}
			return false
		}
	}

	// Check if left is a map
	if m, ok := left.(map[string]interface{}); ok {
		_, exists := m[rightStr]
		return exists
	}

	// Use reflection to check typed maps (map[string]string, map[string]int, etc.)
	if left != nil {
		v := reflect.ValueOf(left)
		if v.Kind() == reflect.Map {
			// For maps, check if the key exists
			keyVal := reflect.ValueOf(rightStr)
			mapKey := v.MapIndex(keyVal)
			if mapKey.IsValid() {
				return true
			}
			// Also try to find the key by iterating (in case key types don't match exactly)
			iter := v.MapRange()
			for iter.Next() {
				keyStr := ToS(iter.Key().Interface(), nil)
				if keyStr == rightStr {
					return true
				}
			}
		}
	}

	return false
}

// ParseConditionExpression parses an expression for use in conditions.
func ParseConditionExpression(parseContext ParseContextInterface, markup string, safe bool) interface{} {
	if ml, ok := conditionMethodLiterals[markup]; ok {
		return ml
	}

	// Check if markup contains filter syntax (|)
	// If so, parse as Variable to support filters in conditions
	if strings.Contains(markup, "|") {
		// Use Variable which supports filter syntax
		return NewVariable(markup, parseContext)
	}

	if safe {
		// For safe parsing, we'd use SafeParseExpression if available
		return parseContext.ParseExpression(markup)
	}
	return parseContext.ParseExpression(markup)
}

// ElseCondition represents an else condition.
type ElseCondition struct {
	*Condition
}

// NewElseCondition creates a new ElseCondition.
func NewElseCondition() *ElseCondition {
	return &ElseCondition{
		Condition: NewCondition(nil, "", nil),
	}
}

// Else returns true for else conditions.
func (e *ElseCondition) Else() bool {
	return true
}

// Evaluate always returns true for else conditions.
func (e *ElseCondition) Evaluate(context ConditionContext) (bool, error) {
	return true, nil
}
