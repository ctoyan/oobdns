package main

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/miekg/dns"
	"github.com/slack-go/slack"
)

func main() {
	domain := flag.String("domain", "", "Your registered domain name")
	webhook := flag.String("webhook", "", "Your Slack webhook URL")
	flag.Parse()

	if *domain == "" {
		fmt.Println("Error: Must supply a domain")
		return
	}

	dns.HandleFunc(".", func(w dns.ResponseWriter, r *dns.Msg) {
		m := new(dns.Msg)
		m.SetReply(r)
		remoteAddr := w.RemoteAddr().String()
		q1 := r.Question[0]
		t := time.Now()

		ns1 := fmt.Sprintf("ns1.%v.", *domain)
		ns2 := fmt.Sprintf("ns2.%v.", *domain)
		mainDomain := fmt.Sprintf("%v.", *domain)

		blacklist := []string{ns1, ns2, mainDomain}
		for _, item := range blacklist {
			if strings.ToLower(q1.Name) == item {
				return
			}
		}

		if !dns.IsSubDomain(*domain+".", q1.Name) {
			return
		}

		addrParts := strings.Split(remoteAddr, ":")

		name := fmt.Sprintf("Lookup Query: `%v`", q1.Name)
		date := fmt.Sprintf("Received At: `%v`", t.Format("Mon Jan _2 15:04:05 2006"))
		from := fmt.Sprintf("Received From: `%v`", addrParts[0])
		queryType := fmt.Sprintf("Query Type: `%v`", dns.TypeToString[q1.Qtype])

		message := fmt.Sprintf("*Received DNS interaction:*\n %v \n %v \n %v \n %v \n", date, from, name, queryType)
		if *webhook != "" {
			sendSlack(message, *webhook)
		} else {
			fmt.Println(message)
		}

		// Server must responsd, because the client keeps making requests
		// and therefor more slack messages are received
		w.WriteMsg(m)
	})
	if err := dns.ListenAndServe("0.0.0.0:53", "udp", nil); err != nil {
		fmt.Println(err.Error())
		return
	}
}

func sendSlack(message string, webhook string) {
	msg := slack.WebhookMessage{
		Text: message,
	}
	_ = slack.PostWebhook(webhook, &msg)
}

func handleInteraction(w dns.ResponseWriter, r *dns.Msg) {
}
