package main

import (
	"regexp"
	"strings"

	"github.com/golang/glog"
)

// AuthzRequest is an authorization request
type AuthzRequest struct {
	User   string
	Groups []string
}

// Access represents a access authorization
/*
type Access struct {
	Type    string   `json:"type"`
	Name    string   `json:"name"`
	Actions []string `json:"actions"`
}
*/

/*
func (a Access) String() string {
	return fmt.Sprintf("%s:%s:%s", a.Type, a.Name, strings.Join(a.Actions, ","))
}
*/

// Accesses represents a set of access
/*
type Accesses []Access
*/

/*
func (as Accesses) String() string {
	accesses := ""
	for _, a := range as {
		accesses = accesses + " " + a.String()
	}
	return strings.Trim(accesses, " ")
}
*/

// GetAccess gets the scope from a string
/*
func GetAccess(s string) *Access {
	access := Access{}
	parts := strings.Split(s, ":")
	switch len(parts) {
	case 3:
		access.Type = parts[0]
		access.Name = parts[1]
		access.Actions = strings.Split(parts[2], ",")
	case 4:
		access.Type = parts[0]
		access.Name = fmt.Sprintf("%s:%s", parts[1], parts[2])
		access.Actions = strings.Split(parts[3], ",")
	default:
		return nil
	}
	return &access
}
*/

// GetAccesses gets scopes from a string
/*
func GetAccesses(s string) *Accesses {
	ss := strings.Split(s, " ")
	accesses := Accesses{}
	for _, v := range ss {
		if len(s) == 0 {
			continue
		}
		access := GetAccess(v)
		if access != nil {
			accesses = append(accesses, *access)
		} else {
			glog.Errorf("Could not parse scope %s", v)
		}
	}
	return &accesses
}
*/

// Eval evaluates a rule
func (r *Rule) Eval(user string, group string, scope Scope, access *Scope) {
	if scope.Type != "repository" {
		glog.Errorf("Requested scope is not repository (%s)", scope.Type)
		return
	}
	match := r.Match
	if strings.Contains(match, "${user}") {
		match = strings.Replace(match, "${user}", user, -1)
	}
	if strings.Contains(match, "${group}") {
		match = strings.Replace(match, "${fgroup}", group, -1)
	}
	matched, err := regexp.MatchString(match, scope.Name)
	if err != nil {
		glog.Error("Error matching rule")
		return
	}
	if !matched {
		return
	}
	if len(r.User) > 0 && r.User != user {
		return
	}
	if len(r.Group) > 0 && r.Group != group {
		return
	}
	for _, a := range scope.Actions {
		for _, ra := range r.Actions {
			if a == ra {
				found := false
				for _, aa := range access.Actions {
					if a == aa {
						found = true
						break
					}
				}
				if !found {
					access.Actions = append(access.Actions, a)
				}
			}
		}
	}

}

func checkAccess(request AuthzRequest, scope Scope) *Scope {
	access := Scope{
		Type:    scope.Type,
		Name:    scope.Name,
		Actions: []string{},
	}
	for _, rule := range AuthConfig.Rules {
		if len(rule.Group) == 0 && !strings.Contains(rule.Match, "${group}") {
			// group evaluation not necessary
			rule.Eval(request.User, "", scope, &access)
			continue
		}
		for _, g := range request.Groups {
			// evaluate for each group
			rule.Eval(request.User, g, scope, &access)
		}
	}
	glog.Infof("checked access %s@%s: %s", request.User, scope.Name, access.Actions)
	return &access
}

// Authorize check authorization of a user for the given scopes
func Authorize(request AuthzRequest, scopes []Scope) Scopes {
	accesses := Scopes{}
	for _, scope := range scopes {
		access := checkAccess(request, scope)
		if access != nil {
			accesses = append(accesses, *access)
		}
	}
	return accesses
}
