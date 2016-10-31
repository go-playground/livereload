package livereload

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/go-playground/log"

	"github.com/gorilla/websocket"
)

// NOTES:
// - Run "go test" to run tests
// - Run "gocov test | gocov report" to report on test converage by file
// - Run "gocov test | gocov annotate -" to report on all code and functions, those ,marked with "MISS" were never called
//
// or
//
// -- may be a good idea to change to output path to somewherelike /tmp
// go test -coverprofile cover.out && go tool cover -html=cover.out -o cover.html
//

const (
	port = 3005
)

var (
	cssFile = `body {
		background-color:#000;
	}`
	jsFile      = `function(test){}`
	txtFile     = "test"
	clientHello = struct {
		Command   string   `json:"command"`
		Protocols []string `json:"protocols"`
	}{
		"hello",
		[]string{
			"http://livereload.com/protocols/official-7",
			"http://livereload.com/protocols/official-8",
			"http://livereload.com/protocols/2.x-origin-version-negotiation",
		},
	}
)

func testFnOK(name string) (bool, error) {
	return true, nil
}

func testFnBAD(name string) (bool, error) {
	return true, errors.New("Bad Error")
}

func testFnOKNoReload(name string) (bool, error) {
	return true, nil
}

func TestMain(m *testing.M) {

	paths := []string{
		"testfiles",
	}

	mappings := ReloadMapping{
		".css": testFnOK,
		".js":  testFnBAD,
		".txt": testFnOKNoReload,
	}

	done, err := ListenAndServe(port, paths, mappings)
	defer close(done)

	if err != nil {
		log.Fatal("could not setup server")
	}

	os.Exit(m.Run())
}

func TestLivereload(t *testing.T) {

	dialer := new(websocket.Dialer)

	conn, resp, err := dialer.Dial(
		fmt.Sprintf("ws://%s:%d/livereload", "127.0.0.1", port),
		http.Header{},
	)
	if err != nil {
		t.Fatal(err)
	}

	defer conn.Close()

	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Fatal("Bad status code, expected 101")
	}

	m := make(map[string]interface{})

	err = conn.ReadJSON(&m)
	if err != nil {
		t.Fatal(err)
	}

	if m["command"] != "hello" {
		log.Fatal("command should have been 'hello'")
	}

	err = conn.WriteJSON(clientHello)
	if err != nil {
		t.Fatal(err)
	}

	// CSS Test
	go func() {
		time.Sleep(time.Second * 1)
		err = ioutil.WriteFile("testfiles/test.css", []byte(cssFile), os.FileMode(0777))
		if err != nil {
			t.Fatal(err)
		}
	}()

	// don't care about error
	defer os.Remove("testfiles/test.css")

	m = make(map[string]interface{})

	err = conn.ReadJSON(&m)
	if err != nil {
		t.Fatal(err)
	}

	expected := "reload"
	if m["command"] != expected {
		log.Fatalf("command should have been '%s'\n", expected)
	}

	expected = "testfiles/test.css"
	if m["path"] != expected {
		log.Fatalf("path should have been '%s'\n", expected)
	}

	// JS Test
	go func() {
		time.Sleep(time.Second * 1)
		err = ioutil.WriteFile("testfiles/test.js", []byte(jsFile), os.FileMode(0777))
		if err != nil {
			t.Fatal(err)
		}
	}()

	// don't care about error
	defer os.Remove("testfiles/test.js")

	// there will be no response for JS

	// .txt Test
	go func() {
		time.Sleep(time.Second * 1)
		err = ioutil.WriteFile("testfiles/test.txt", []byte(jsFile), os.FileMode(0777))
		if err != nil {
			t.Fatal(err)
		}
	}()

	// don't care about error
	defer os.Remove("testfiles/test.txt")

	m = make(map[string]interface{})

	err = conn.ReadJSON(&m)
	if err != nil {
		t.Fatal(err)
	}

	expected = "reload"
	if m["command"] != expected {
		log.Fatalf("command should have been '%s'\n", expected)
	}

	expected = "testfiles/test.txt"
	if m["path"] != expected {
		log.Fatalf("path should have been '%s'\n", expected)
	}
}
