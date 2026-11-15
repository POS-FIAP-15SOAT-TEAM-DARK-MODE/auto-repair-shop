// Package match resolves an HTTP method and path to a gateway route.
//
// Routes are tried in the order they were given (the first match wins). A
// route pattern is a "/"-separated list of segments where:
//
//   - a plain segment must be equal to the request segment;
//   - ":name" matches exactly one non-empty request segment;
//   - a trailing "*" on the last segment matches any suffix of that segment
//     and any further segments ("/v1/auth/*", "/v1/customers*").
//
// The package has no dependencies outside the standard library so that the
// routing rules can be tested without HTTP machinery.
package match

import (
	"errors"
	"fmt"
	"strings"
)

// Route is one entry of the routing table.
type Route struct {
	// Methods lists the accepted HTTP methods. "*" accepts any method.
	Methods []string
	// Path is the route pattern, see the package documentation.
	Path string
	// Upstream is the name of the upstream that receives matching requests.
	// It is empty for blocked routes.
	Upstream string
	// Block answers 404 without proxying. It hides paths that must never be exposed.
	Block bool
}

// Matcher holds a compiled routing table.
type Matcher struct {
	entries []entry
}

type entry struct {
	route     Route
	anyMethod bool
	methods   []string
	segments  []segment
	prefix    bool
}

type segment struct {
	literal string
	name    string
	param   bool
}

// New compiles the routing table. It fails on malformed patterns, so a bad
// routes file stops the gateway at startup instead of misrouting requests.
func New(routes []Route) (*Matcher, error) {
	m := &Matcher{entries: make([]entry, 0, len(routes))}
	for i, r := range routes {
		e, err := compile(r)
		if err != nil {
			return nil, fmt.Errorf("route %d (%s): %w", i, r.Path, err)
		}
		m.entries = append(m.entries, e)
	}
	return m, nil
}

// Match returns the first route that accepts the method and path.
// Paths containing "." or ".." segments never match.
func (m *Matcher) Match(method, path string) (Route, bool) {
	if !strings.HasPrefix(path, "/") {
		return Route{}, false
	}
	parts := strings.Split(path[1:], "/")
	for _, p := range parts {
		if p == "." || p == ".." {
			return Route{}, false
		}
	}

	method = strings.ToUpper(method)
	for i := range m.entries {
		e := &m.entries[i]
		if e.acceptsMethod(method) && e.matches(parts) {
			return e.route, true
		}
	}
	return Route{}, false
}

func compile(r Route) (entry, error) {
	if !strings.HasPrefix(r.Path, "/") {
		return entry{}, errors.New("path must start with /")
	}
	if len(r.Methods) == 0 {
		return entry{}, errors.New("at least one method is required")
	}
	if r.Block && r.Upstream != "" {
		return entry{}, errors.New("a blocked route has no upstream")
	}
	if !r.Block && r.Upstream == "" {
		return entry{}, errors.New("a route needs an upstream or block: true")
	}

	e := entry{route: r}
	for _, m := range r.Methods {
		m = strings.ToUpper(strings.TrimSpace(m))
		if m == "*" {
			e.anyMethod = true
			e.methods = nil
			break
		}
		if m == "" {
			return entry{}, errors.New("empty method")
		}
		e.methods = append(e.methods, m)
	}

	parts := strings.Split(strings.TrimPrefix(r.Path, "/"), "/")
	for i, p := range parts {
		last := i == len(parts)-1
		if strings.Contains(p, "*") {
			if !last || !strings.HasSuffix(p, "*") || strings.Count(p, "*") > 1 {
				return entry{}, errors.New("'*' is only allowed as the last character of the last segment")
			}
			e.prefix = true
			p = strings.TrimSuffix(p, "*")
			if strings.HasPrefix(p, ":") {
				return entry{}, errors.New("a parameter cannot be combined with '*'")
			}
			e.segments = append(e.segments, segment{literal: p})
			continue
		}
		if strings.HasPrefix(p, ":") {
			name := strings.TrimPrefix(p, ":")
			if name == "" {
				return entry{}, errors.New("parameter needs a name")
			}
			e.segments = append(e.segments, segment{name: name, param: true})
			continue
		}
		if p == "" && !last {
			return entry{}, errors.New("empty path segment")
		}
		e.segments = append(e.segments, segment{literal: p})
	}
	return e, nil
}

func (e *entry) acceptsMethod(method string) bool {
	if e.anyMethod {
		return true
	}
	for _, m := range e.methods {
		if m == method {
			return true
		}
	}
	return false
}

func (e *entry) matches(parts []string) bool {
	if e.prefix {
		if len(parts) < len(e.segments) {
			return false
		}
	} else if len(parts) != len(e.segments) {
		return false
	}

	for i, s := range e.segments {
		part := parts[i]
		switch {
		case s.param:
			if part == "" {
				return false
			}
		case e.prefix && i == len(e.segments)-1:
			if !strings.HasPrefix(part, s.literal) {
				return false
			}
		case part != s.literal:
			return false
		}
	}
	return true
}
