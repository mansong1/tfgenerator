package main

import (
	"regexp"
	"testing"
)

// TestGenerateRandomHexColor tests the generateRandomHexColor function.
func TestGenerateRandomHexColor(t *testing.T) {
	color := generateRandomHexColor()
	match, _ := regexp.MatchString("^#[0-9a-fA-F]{6}$", color)

	if !match {
		t.Errorf("Generated color '%s' is not a valid hex color", color)
	}
}
