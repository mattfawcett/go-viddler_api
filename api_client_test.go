package viddler_api

import (
  "testing"
  "net/http"
  "net/http/httptest"
  "fmt"
  "net/url"
)

const ApiKey string = "my-api-key"

var (
  mux *http.ServeMux
  client *Client
  server *httptest.Server
)

func setup() {
  mux    = http.NewServeMux()
  server = httptest.NewServer(mux)

  client = NewClient(ApiKey)
  urlObj, _ := url.Parse(server.URL)
  client.BaseDomain = urlObj.String()
}

func TestNewClient(t *testing.T) {
  setup()

  if client.ApiKey != ApiKey {
    t.Errorf("API key was not set properly")
  }
}

func TestGet(t *testing.T) {
  setup()

  mux.HandleFunc("/api/v2/viddler.api.echo.json",
    func(w http.ResponseWriter, r *http.Request) {
      testMethod(t, r, "GET")
      testFormValues(t, r, map[string]string{"message": "hello there"})

      fmt.Fprint(w, `{"echo_response":{"message":"hello there"}}`)
    },
  )

  params := map[string]string{
    "message": "hello there",
  }
  response, err := client.Get("viddler.api.echo", params)

  if err != nil {
    t.Errorf("Post returned an unexpected error")
  }

  message, _ := response.Get("echo_response").Get("message").String()
  if message != "hello there" {
    t.Errorf("API did not echo back the response")
  }
}

func TestShouldIncludeSessionIdIfSet(t *testing.T) {
  setup()

  mux.HandleFunc("/api/v2/viddler.api.echo.json",
    func(w http.ResponseWriter, r *http.Request) {
      testFormValues(t, r, map[string]string{"message": "hello there", "sessionid": "a-session-id"})

      fmt.Fprint(w, `{"echo_response":{"message":"hello there"}}`)
    },
  )

  params := map[string]string{
    "message": "hello there",
  }
  client.SessionId = "a-session-id"
  response, err := client.Get("viddler.api.echo", params)

  if err != nil {
    t.Errorf("Post returned an unexpected error")
  }

  message, _ := response.Get("echo_response").Get("message").String()
  if message != "hello there" {
    t.Errorf("API did not echo back the response")
  }
}

func TestPost(t *testing.T) {
  setup()

  mux.HandleFunc("/api/v2/viddler.users.setSettings.json",
    func(w http.ResponseWriter, r *http.Request) {
      testMethod(t, r, "POST")
      testFormValues(t, r, map[string]string{"name": "bob", "sessionid": "a-session-id"})

      fmt.Fprint(w, `{"success":"true"}`)
    },
  )

  params := map[string]string{
    "name": "bob",
  }
  client.SessionId = "a-session-id"
  response, err := client.Post("viddler.users.setSettings", params)

  if err != nil {
    t.Errorf("Post returned an unexpected error")
  }

  success, _ := response.Get("success").String()
  if success != "true" {
    t.Errorf("API did not echo back the response")
  }
}

func TestShouldHandleErrors(t *testing.T) {
  setup()

  mux.HandleFunc("/api/v2/viddler.api.echo.json",
    func(w http.ResponseWriter, r *http.Request) {
      fmt.Fprint(w, `{"error":{"code":"4","description":"missing required parameter","details":"message"}}`)
    },
  )

  params := map[string]string{}
  _, err := client.Get("viddler.api.echo", params)

  if err.Error() != "missing required parameter" {
    t.Errorf("Post returned an unexpected error")
  }
}

func TestAuthenticate(t *testing.T) {
  setup()

  mux.HandleFunc("/api/v2/viddler.users.auth.json",
    func(w http.ResponseWriter, r *http.Request) {
      testMethod(t, r, "GET")
      testFormValues(t, r, map[string]string{"username": "auser", "password": "apassword"})

      fmt.Fprint(w, `{"auth":{"sessionid":"thesessionid"}}`)
    },
  )

  success := client.Authenticate("auser", "apassword")
  if !success {
    t.Errorf("Authentication failed")
  }

  if client.SessionId != "thesessionid" {
    t.Errorf("The session id was not set correctly")
  }
}

// copied from https://github.com/google/go-github/blob/3bb8a96d4846d1bef2f45e0b27eef4bcbbca2df0/github/github_test.go
type values map[string]string
func testFormValues(t *testing.T, r *http.Request, values values) {
  for key, want := range values {
    if v := r.FormValue(key); v != want {
      t.Errorf("Request parameter %v = %v, want %v", key, v, want)
    }
  }
}

func testMethod(t *testing.T, r *http.Request, want string) {
  if want != r.Method {
    t.Errorf("Request method = %v, want %v", r.Method, want)
  }
}
