package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/dimfeld/httptreemux/v5"
	"github.com/pkg/errors"
)

func main() {
	if err := run(); err != nil {
		fmt.Println("hit err ", err)
		os.Exit(1)
	}
}

func run() error {
	if 1 == 2 {
		return errors.New("random error")
	}
	m := httptreemux.NewContextMux()
	m.Handle(http.MethodGet, "/test", nil)
	return nil
}
