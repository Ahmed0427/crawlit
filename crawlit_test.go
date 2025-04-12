package main

import (
	"testing"
	"reflect"
)

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name          string
		inputURL      string
		expected      string
	}{
		{
			name:     "normalize url",
			inputURL: "https://blog.boot.dev/path",
			expected: "blog.boot.dev/path",
		},
		{
			name:     "normalize url",
			inputURL: "https://blog.boot.dev/path/",
			expected: "blog.boot.dev/path",
		},
		{
			name:     "normalize url",
			inputURL: "http://blog.boot.dev/path",
			expected: "blog.boot.dev/path",
		},
		{
			name:     "normalize url",
			inputURL: "http://blog.boot.dev/path/",
			expected: "blog.boot.dev/path",
		},
	}

	for i, tc := range tests {
		actual, err := normalizeURL(tc.inputURL)
		if err != nil {
			t.Errorf("Test %v - '%s' FAIL: unexpected error: %v",
				i, tc.name, err)

		} else if actual != tc.expected {
			t.Errorf("Test %v - '%s' FAIL: expected : %v, output: %v",
				i, tc.name, tc.expected, actual)
		}
	}
}

func TestGetURLsFromHTML(t *testing.T) {
	tests := []struct {
		name          string
		inputURL      string
		inputBody     string
		expected      []string
	}{
		{
			name:     "absolute and relative URLs",
			inputURL: "https://blog.boot.dev",
			inputBody: `
			<html>
				<body>
					<a href="/path/one">
						<span>Boot.dev</span>
					</a>
					<a href="https://other.com/path/one">
						<span>Boot.dev</span>
					</a>
				</body>
			</html>
			`,
			expected: []string{"https://blog.boot.dev/path/one",
				"https://other.com/path/one"},
		},
	}

	for i, tc := range tests {
		actual, err := getURLsFromHTML(tc.inputBody, tc.inputURL)
		if err != nil {
			t.Errorf("Test %v - '%s' FAIL: unexpected error: %v",
				i, tc.name, err)

		} else if !reflect.DeepEqual(actual, tc.expected) {
			t.Errorf("Test %v - '%s' FAIL: expected : %v, output: %v",
				i, tc.name, tc.expected, actual)
		}
	}
}
