package main

import (
	"fmt"
	"net/url"
	"os"

	influx "github.com/influxdb/influxdb/client"
)

type Backend interface {
}

type influxBackend struct {
	Host string
	Port int
	Conn *influx.Client
}

func (b *influxBackend) Connect() (err error) {
	var (
		u *url.URL
	)

	if u, err = url.Parse(fmt.Sprintf("http://%s:%d", b.Host, b.Port)); err != nil {
		return
	}

	conf := influx.Config{
		URL:      *u,
		Username: os.Getenv("INFLUX_USER"),
		Password: os.Getenv("INFLUX_PWD"),
	}

	if b.Conn, err = influx.NewClient(conf); err != nil {
		return
	}

	return
}

func NewInfluxBackend(host string, port int) Backend {
	b := &influxBackend{host, port, nil}
	return b
}
