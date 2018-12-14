package main

import (
	_ "errors"
	_ "flag"
	"fmt"
	_ "log"
	_ "os"
	_ "os/signal"
	_ "path/filepath"
	_ "syscall"
	_ "text/template"

	_ "github.com/Sirupsen/logrus"

	// _ "github.com/prometheus/client_golang/prometheus"
	// _ "github.com/prometheus/client_golang/prometheus/promhttp"

	_ "github.com/gorilla/mux"
)

func main() {
	fmt.Println("This is used for build caching purposes. Should be replaced.")
}
