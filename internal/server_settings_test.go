package internal

import (
	"reflect"
	"testing"
)

func TestSelectChangedServers(t *testing.T) {
	local := []Server{
		{Domain: "a.com", Snippets: "foo", Cert: Cert{Certificate: "certA", PrivateKey: "keyA"}},
		{Domain: "b.com", Snippets: "bar", Cert: Cert{Certificate: "certB", PrivateKey: "keyB"}},
	}
	remote := []Server{
		{Domain: "a.com", Snippets: "foo", Cert: Cert{Certificate: "certA", PrivateKey: "keyA"}}, // unchanged
		{Domain: "b.com", Snippets: "baz", Cert: Cert{Certificate: "certB", PrivateKey: "keyB"}}, // changed
		{Domain: "c.com", Snippets: "qux", Cert: Cert{Certificate: "certC", PrivateKey: "keyC"}}, // new
	}

	expected := []Server{
		{Domain: "b.com", Snippets: "baz", Cert: Cert{Certificate: "certB", PrivateKey: "keyB"}},
		{Domain: "c.com", Snippets: "qux", Cert: Cert{Certificate: "certC", PrivateKey: "keyC"}},
	}

	changed := SelectChangedServers(local, remote)
	if !reflect.DeepEqual(changed, expected) {
		t.Errorf("Expected %v, got %v", expected, changed)
	}
}

func TestSelectRemovedServers(t *testing.T) {
	local := []Server{
		{Domain: "a.com"},
		{Domain: "b.com"},
		{Domain: "c.com"},
	}
	remote := []Server{
		{Domain: "a.com"},
		{Domain: "c.com"},
	}

	expected := []Server{
		{Domain: "b.com"},
	}

	removed := SelectRemovedServers(local, remote)
	if !reflect.DeepEqual(removed, expected) {
		t.Errorf("Expected %v, got %v", expected, removed)
	}
}
