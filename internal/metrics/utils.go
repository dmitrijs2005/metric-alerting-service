package metrics

import (
	"regexp"
)

// Metric names and labels
// Every time series is uniquely identified by its metric name and optional key-value pairs called labels.

// Metric names:

// Specify the general feature of a system that is measured (e.g. http_requests_total - the total number of HTTP requests received).
// Metric names may contain ASCII letters, digits, underscores, and colons. It must match the regex [a-zA-Z_:][a-zA-Z0-9_:]*.

func IsMetricNameValid(name string) bool {
	pattern := `^[a-zA-Z_:][a-zA-Z0-9_:]*$`

	// Compile the regular expression
	re := regexp.MustCompile(pattern)

	re.MatchString(name)

	return re.MatchString(name)
}
