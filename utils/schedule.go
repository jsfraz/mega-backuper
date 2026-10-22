package utils

import (
	"fmt"
	"strings"

	"github.com/robfig/cron/v3"
)

// ParseSchedule validates a cron expression the same way the scheduler will
// consume it and reports whether the expression uses a leading seconds field.
//
// Supported expressions:
//   - 5 fields (minute, hour, day of month, month, day of week)
//   - 6 fields with a leading seconds field
//   - descriptors (`@daily`, `@every 1h`, ...)
//
// Expressions with a year field (7 fields) are rejected, the scheduler cannot
// run them.
//
//	@param expression
//	@return bool True when the expression uses a seconds field.
//	@return error
func ParseSchedule(expression string) (bool, error) {
	expr := strings.TrimSpace(expression)
	if expr == "" {
		return false, fmt.Errorf("cron expression is empty")
	}

	if strings.HasPrefix(expr, "@") {
		if _, err := cron.ParseStandard(expr); err != nil {
			return false, fmt.Errorf("invalid cron expression %q: %w", expression, err)
		}
		return false, nil
	}

	fields := strings.Fields(expr)
	switch len(fields) {
	case 5:
		if _, err := cron.ParseStandard(expr); err != nil {
			return false, fmt.Errorf("invalid cron expression %q: %w", expression, err)
		}
		return false, nil
	case 6:
		parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
		if _, err := parser.Parse(expr); err != nil {
			return false, fmt.Errorf("invalid cron expression %q: %w", expression, err)
		}
		return true, nil
	default:
		return false, fmt.Errorf("invalid cron expression %q: expected 5 or 6 fields, got %d", expression, len(fields))
	}
}
