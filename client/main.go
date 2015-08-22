package main

import (
	"bytes"
	"flag"
	"log"
	"strings"
	"xep/c2s/actors"
	"xep/c2s/actors/steps"
	"xep/c2s/stream"
	"xep/entity"
	"xep/units"
)

var user string
var pwd string
var server string
var resource string

func init() {
	flag.StringVar(&user, "u", "goxep", "-u=user")
	flag.StringVar(&server, "s", "xmpp.ru", "-s=server")
	flag.StringVar(&resource, "r", "go", "-r=resource")
	flag.StringVar(&pwd, "p", "GogogOg0", "-p=password")
	log.SetFlags(0)
}

func conv(fn func(entity.Entity)) func(*bytes.Buffer) *bytes.Buffer {
	return func(in *bytes.Buffer) (out *bytes.Buffer) {
		if _e, err := entity.Consume(in); err == nil {
			switch e := _e.(type) {
			case *entity.Message:
				fn(e)
			}
		} else {
			log.Println("ERROR", err)
			log.Println(string(in.Bytes()))
			log.Println()
		}
		return
	}
}

func main() {
	flag.Parse()
	s := &units.Server{Name: server}
	c := &units.Client{Name: user, Server: s}
	st := stream.New(s)
	if err := stream.Dial(st); err == nil {
		errHandler := func(err error) {
			log.Fatal(err)
		}
		neg := &steps.Negotiation{}
		actors.With(st).Do(steps.Starter, errHandler).Do(neg.Act(), errHandler).Run()
		if neg.HasMechanism("PLAIN") {
			auth := &steps.PlainAuth{Client: c, Pwd: pwd}
			neg := &steps.Negotiation{}
			bind := &steps.Bind{Rsrc: resource}
			actors.With(st).Do(auth.Act(), errHandler).Do(steps.Starter).Do(neg.Act()).Do(bind.Act()).Do(steps.Session).Run()
			actors.With(st).Do(steps.InitialPresence).Run()
			actors.With(st).Do(func(st stream.Stream) error {
				actors.With(st).Do(steps.PresenceTo("golang@conference.jabber.ru/xep")).Run()
				for {
					st.Ring(conv(func(_e entity.Entity) {
						switch e := _e.(type) {
						case *entity.Message:
							log.Println(strings.TrimPrefix(e.From, "golang@conference.jabber.ru/"))
							log.Println(e.Body)
						}
					}), 0)
				}
			}).Run()
		}
	} else {
		log.Fatal(err)
	}
}
