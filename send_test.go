package amzses

import (
	"flag"
	"testing"
)

var to, from string

func init() {
	flag.StringVar(&to, "to", "", "email recipient")
	flag.StringVar(&from, "from", "", "email sender")
}

func checkFlags(t *testing.T) {
	if len(to) == 0 {
		t.Fatal("must specify recipient via -to flag.")
	}
	if len(from) == 0 {
		t.Fatal("must specify sender via -from flag.")
	}
}

func TestText(t *testing.T) {
	checkFlags(t)
	_, err := SendMail(from, to, "amzses text test", textBody)
	if err != nil {
		t.Fatal(err)
	}
}

func TestHtml(t *testing.T) {
	checkFlags(t)
	_, err := SendMailHTML(from, to, "amzses html test", textBody, htmlBody)
	if err != nil {
		t.Fatal(err)
	}
}

var textBody = `This is an example email body for the amzses go package.`

var htmlBody = `
This is an <b>html email</b>.
<br/>
<br/>
<img src="http://placehold.it/600x200/">
`
