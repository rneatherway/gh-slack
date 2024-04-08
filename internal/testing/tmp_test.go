package testing

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestXxx(t *testing.T) {
	f := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World!"))
	})

	server := httptest.NewServer(f)
	t.Cleanup(server.Close)

	recorder := NewRecorder("fixture.json", http.DefaultTransport)

	client := &http.Client{
		Transport: recorder,
	}

	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Body %q", string(bs))

	err = recorder.Close()
	if err != nil {
		t.Fatal(err)
	}

	server.Close()

	replayer, err := NewReplayer("fixture.json")
	if err != nil {
		t.Fatal(err)
	}

	client.Transport = replayer
	resp, err = client.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	bs, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Replay %q", string(bs))

	// also play it back with an incorrect url so that it doesn't match
}
