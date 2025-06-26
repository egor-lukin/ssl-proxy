package internal

// import (
// 	"fmt"
// )

type Cert struct {
	Domain      string
	PrivateKey  string
	Certificate string
}

type Server struct {
	Domain   string
	Snippets string
	Cert     Cert
}

func (s Server) Equal(other Server) bool {
	return s.Cert.Certificate == other.Cert.Certificate &&
		s.Cert.PrivateKey == other.Cert.PrivateKey &&
		s.Snippets == other.Snippets
}

func SelectRemovedServers(localServers, remoteServers []Server) []Server {
	newDomains := make(map[string]struct{})
	for _, s := range remoteServers {
		newDomains[s.Domain] = struct{}{}
	}
	var removedServers []Server
	for _, s := range localServers {
		if _, exists := newDomains[s.Domain]; !exists {
			removedServers = append(removedServers, s)
		}
	}

	return removedServers
}

func SelectChangedServers(localServers, removeServers []Server) []Server {
	oldMap := make(map[string]Server)
	for _, s := range localServers {
		oldMap[s.Domain] = s
	}

	return filter(removeServers, func(s Server) bool {
		old, exists := oldMap[s.Domain]
		return !exists || !s.Equal(old)
	})
}

func filter(servers []Server, predicate func(Server) bool) []Server {
	var filtered []Server
	for _, s := range servers {
		if predicate(s) {
			filtered = append(filtered, s)
		}
	}
	return filtered
}
