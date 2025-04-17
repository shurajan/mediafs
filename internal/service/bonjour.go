package service

import (
	"log"

	"github.com/grandcat/zeroconf"
)

func PublishBonjour() {
	server, err := zeroconf.Register(
		"MediaFS",
		"_http._tcp",
		"local.",
		8000,
		nil,
		nil,
	)
	if err != nil {
		log.Println("❌ Failed to publish Bonjour service:", err)
		return
	}
	log.Println("✅ Bonjour service 'MediaFS._http._tcp.local' published")

	<-make(chan struct{})
	defer server.Shutdown()
}
