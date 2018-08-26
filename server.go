package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang/glog"
)

// AuthRequest respresent an authentication request
type AuthRequest struct {
	RemoteAddr string
	UserName   string
	Password   PasswordString
	ClientID   string
	Service    string
	Scopes     []Scope
}

// Scope defined the required resources and actions
type Scope struct {
	Type    string
	ID      string
	Actions []string
}

// PasswordString hides passwords on output
type PasswordString string

func (s PasswordString) String() string {
	if len(s) == 0 {
		return ""
	}
	return "********"
}

// GetAuthRequest gets an authentication request from the http request
func GetAuthRequest(r *http.Request) *AuthRequest {
	ar := AuthRequest{}
	ar.RemoteAddr = r.RemoteAddr
	if username, password, ok := r.BasicAuth(); ok {
		ar.UserName = username
		ar.Password = PasswordString(password)
	} else {
		glog.Errorf("Authentication request didn't have basic authentication")
		return nil
	}
	service := r.FormValue("service")
	if len(service) == 0 {
		glog.Errorf("No service provided")
		return nil
	}
	ar.Service = service
	clientid := r.FormValue("client_id")
	if len(clientid) == 0 {
		glog.Error("No ClientId provided")
		return nil
	}
	scopeString := r.FormValue("scope")
	if len(scopeString) == 0 {
		glog.Infof("No scopes provided")
	}
	scopes := strings.Split(scopeString, " ")
	ar.Scopes = []Scope{}
	for _, v := range scopes {
		scope := GetScope(v)
		if scope != nil {
			ar.Scopes = append(ar.Scopes, *scope)
		} else {
			glog.Errorf("Could not parse scope %s", v)
		}
	}
	return &ar
}

func (ar *AuthRequest) String() string {
	return fmt.Sprintf("%s:%s - ip='%s' client_id='%s' service='%s' scopes=%s", ar.UserName, ar.Password, ar.RemoteAddr, ar.ClientID, ar.Service, ar.Scopes)
}

// GetScope ngets the scope from a string
func GetScope(s string) *Scope {
	scope := Scope{}
	parts := strings.Split(s, ":")
	switch len(parts) {
	case 3:
		scope.Type = parts[0]
		scope.ID = parts[1]
		scope.Actions = strings.Split(parts[2], ",")
	case 4:
		scope.Type = parts[0]
		scope.ID = fmt.Sprintf("%s:%s", parts[1], parts[2])
		scope.Actions = strings.Split(parts[3], ",")
	default:
		return nil
	}
	return &scope
}

// HandleAuth Authenticates and resturns a token.
func HandleAuth(w http.ResponseWriter, r *http.Request) {
	glog.Infof("Call to authentication endpoint")
	// check parameters
	ar := GetAuthRequest(r)
	if ar == nil {
		http.Error(w, "Authentication failed", http.StatusInternalServerError)
	}
	glog.Infof("Authentication Request: %s", ar.String())
	_, ok := Authenticate(ar.UserName, string(ar.Password))
	if !ok {
		glog.Infof("User %s not authenticated return empty token", ar.UserName)
	}
	glog.Infof("User %s authenticated check authorizations", ar.UserName)
	fmt.Fprintf(w, "Test page.")
}
