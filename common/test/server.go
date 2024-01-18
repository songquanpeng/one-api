package test

import (
	"log"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
)

const testAPI = "this-is-my-secure-token-do-not-steal!!"

func GetTestToken() string {
	return testAPI
}

type ServerTest struct {
	handlers map[string]handler
}
type handler func(w http.ResponseWriter, r *http.Request)

func NewTestServer() *ServerTest {
	return &ServerTest{handlers: make(map[string]handler)}
}

func OpenAICheck(w http.ResponseWriter, r *http.Request) bool {
	if r.Header.Get("Authorization") != "Bearer "+GetTestToken() && r.Header.Get("api-key") != GetTestToken() {
		w.WriteHeader(http.StatusUnauthorized)
		return false
	}
	return true
}

func (ts *ServerTest) RegisterHandler(path string, handler handler) {
	// to make the registered paths friendlier to a regex match in the route handler
	// in OpenAITestServer
	path = strings.ReplaceAll(path, "*", ".*")
	ts.handlers[path] = handler
}

// OpenAITestServer Creates a mocked OpenAI server which can pretend to handle requests during testing.
func (ts *ServerTest) TestServer(headerCheck func(w http.ResponseWriter, r *http.Request) bool) *httptest.Server {
	return httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("received a %s request at path %q\n", r.Method, r.URL.Path)

		// check auth
		if headerCheck != nil && !headerCheck(w, r) {
			return
		}

		// Handle /path/* routes.
		// Note: the * is converted to a .* in register handler for proper regex handling
		for route, handler := range ts.handlers {
			// Adding ^ and $ to make path matching deterministic since go map iteration isn't ordered
			pattern, _ := regexp.Compile("^" + route + "$")
			if pattern.MatchString(r.URL.Path) {
				handler(w, r)
				return
			}
		}
		http.Error(w, "the resource path doesn't exist", http.StatusNotFound)
	}))
}
