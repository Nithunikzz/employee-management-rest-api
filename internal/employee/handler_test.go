package employee

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	NewHandler(NewService(NewStore())).Register(mux)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func post(t *testing.T, srv *httptest.Server, body string) *http.Response {
	t.Helper()
	resp, err := http.Post(srv.URL+"/employees", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func list(t *testing.T, srv *httptest.Server, query string) (int, []Employee) {
	t.Helper()
	resp, err := http.Get(srv.URL + "/employees" + query)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var out struct {
		Employees []Employee `json:"employees"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&out)
	return resp.StatusCode, out.Employees
}

func names(es []Employee) string {
	var s []string
	for _, e := range es {
		s = append(s, e.Name)
	}
	return strings.Join(s, ",")
}

func TestCreateValidation(t *testing.T) {
	srv := newServer(t)
	cases := map[string]int{
		`{"name":"Ann","department":"Eng","salary":100}`: http.StatusCreated,
		`{"name":"Ann","department":"Eng","salary":0}`:   http.StatusCreated,
		`{"name":"","department":"Eng","salary":100}`:    http.StatusBadRequest,
		`{"name":"Ann","department":"","salary":100}`:    http.StatusBadRequest,
		`{"name":"Ann","department":"Eng"}`:              http.StatusBadRequest,
		`{"name":"Ann","department":"Eng","salary":-1}`:  http.StatusBadRequest,
		`{"name":"Ann","department":"Eng","salary":"x"}`: http.StatusBadRequest,
		`{"name":"Ann","dept":"Eng","salary":1}`:         http.StatusBadRequest,
		`not json`:                                       http.StatusBadRequest,
	}
	for body, want := range cases {
		resp := post(t, srv, body)
		resp.Body.Close()
		if resp.StatusCode != want {
			t.Errorf("POST %s: got %d, want %d", body, resp.StatusCode, want)
		}
	}
}

func TestListFilterAndSort(t *testing.T) {
	srv := newServer(t)
	for _, b := range []string{
		`{"name":"Carol","department":"Engineering","salary":120000}`,
		`{"name":"alice","department":"Sales","salary":70000}`,
		`{"name":"Bob","department":"Engineering","salary":95000}`,
		`{"name":"Alan","department":"HR","salary":60000}`,
	} {
		post(t, srv, b).Body.Close()
	}

	tests := []struct {
		query string
		want  string
	}{
		{"", "Carol,alice,Bob,Alan"},
		{"?name=al", "alice,Alan"},
		{"?department=engineering", "Carol,Bob"},
		{"?min_salary=65000&max_salary=100000", "alice,Bob"},
		{"?min_salary=95000", "Carol,Bob"},
		{"?sort_by=name", "Alan,alice,Bob,Carol"},
		{"?sort_by=name&order=desc", "Carol,Bob,alice,Alan"},
		{"?sort_by=salary", "Alan,alice,Bob,Carol"},
		{"?sort_by=salary&order=desc", "Carol,Bob,alice,Alan"},
		{"?sort_by=department", "Carol,Bob,Alan,alice"},
		{"?department=Engineering&sort_by=salary", "Bob,Carol"},
	}
	for _, tt := range tests {
		code, got := list(t, srv, tt.query)
		if code != http.StatusOK || names(got) != tt.want {
			t.Errorf("GET %q: got %d %q, want 200 %q", tt.query, code, names(got), tt.want)
		}
	}

	for _, q := range []string{"?sort_by=age", "?order=up", "?min_salary=abc", "?min_salary=10&max_salary=5"} {
		if code, _ := list(t, srv, q); code != http.StatusBadRequest {
			t.Errorf("GET %q: got %d, want 400", q, code)
		}
	}
}
