package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLastPositionPage(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	}))
	defer ts.Close()

	client := ts.Client()
	res, err := client.Get(ts.URL)
	require.Nil(t, err)

	greeting, err := io.ReadAll(res.Body)
	res.Body.Close()
	require.Nil(t, err)

	fmt.Printf("%s", greeting)

	require.Equal(t, "Hello, client\n", string(greeting))
}
