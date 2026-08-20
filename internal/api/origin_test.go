package api

import "testing"

func TestOriginAddress(t *testing.T) {
	got := originAddress(
		"providencewilmington.classreach.com:443",
		"providencewilmington.classreach.com",
		"classreach.azurewebsites.net",
	)
	if got != "classreach.azurewebsites.net:443" {
		t.Fatalf("origin address = %q", got)
	}
}

func TestOriginAddressPreservesOtherHosts(t *testing.T) {
	const address = "example.com:443"
	got := originAddress(
		address,
		"tenant.classreach.com",
		"classreach.azurewebsites.net",
	)
	if got != address {
		t.Fatalf("origin address = %q", got)
	}
}
