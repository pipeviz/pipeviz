package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/spf13/cobra"
	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/unrolled/secure"
	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/zenazn/goji/graceful"
	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/zenazn/goji/web"
	"github.com/tag1consulting/pipeviz/ingest"
	"github.com/tag1consulting/pipeviz/log"
)

func main() {
	s := &srv{}
	root := &cobra.Command{
		Use:   "pvproxy",
		Short: "pvproxy transforms native payloads into pipeviz-compatible messages.",
		Run:   s.Run,
	}

	root.Flags().StringVarP(&s.target, "target", "t", "http://localhost:2309", "Address of the target pipeviz daemon. Default to http://localhost:2309")
	root.Flags().IntVarP(&s.port, "port", "p", 2906, "Port to listen on") // 2906, because Viv
	root.Flags().StringVarP(&s.bind, "bind", "b", "127.0.0.1", "Address to bind on")
	root.Flags().BoolVar(&s.useSyslog, "syslog", false, "Write log output to syslog.")
	root.Flags().StringVar(&s.key, "tls-key", "", "Path to an x509 key to use for TLS. If no cert is provided, unsecured HTTP will be used.")
	root.Flags().StringVar(&s.cert, "tls-cert", "", "Path to an x508 certificate to use for TLS. If key is provided, will try to find a certificate of the same name plus .crt extension.")
	root.Flags().StringVar(&s.syslogAddr, "syslog-addr", "localhost:514", "The address of the syslog server with which to communicate.")
	root.Flags().StringVar(&s.syslogProto, "syslog-proto", "udp", "The protocol over which to send syslog messages.")

	root.Flags().String("github-oauth", "", "OAuth token for retrieving data from Github")
	root.Execute()
}

type srv struct {
	port                    int
	bind, target            string
	key, cert               string
	useSyslog               bool
	syslogAddr, syslogProto string
}

type client struct {
	target string
	c      http.Client
	h      http.Header
}

func newClient(target string, timeout time.Duration) client {
	return client{
		target: target,
		c:      http.Client{Timeout: timeout},
		h:      make(http.Header),
	}
}

// TODO return msgid sent back from pipeviz backend as uint64
func (c client) send(m *ingest.Message) error {
	j, err := json.Marshal(m)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.target, bytes.NewReader(j))
	// TODO is it safe to reuse the header map like this?
	req.Header = c.h
	if err != nil {
		return err
	}

	resp, err := c.c.Post(c.target, "application/json", bytes.NewReader(j))
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"system": "pvproxy",
			"err":    err,
		}).Warn("Error returned from backend")
		return err
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return fmt.Errorf("Pipeviz backend rejected message with code %d", resp.StatusCode)
	}

	return nil
}

// Run sets up and runs the proxying HTTP server, then blocks.
func (s *srv) Run(cmd *cobra.Command, args []string) {
	setUpLogging(s)

	mux := web.New()
	cl := newClient(s.target, 5*time.Second)

	mux.Use(log.NewHttpLogger("pvproxy"))

	if s.key != "" && s.cert == "" {
		s.cert = s.key + ".crt"
	}
	useTLS := s.key != "" && s.cert != ""
	if useTLS {
		sec := secure.New(secure.Options{
			AllowedHosts:         nil,                                             // TODO allow a way to declare these
			SSLRedirect:          false,                                           // we have just one port to work with, so an internal redirect can't work
			SSLTemporaryRedirect: false,                                           // Use 301, not 302
			SSLProxyHeaders:      map[string]string{"X-Forwarded-Proto": "https"}, // list of headers that indicate we're using TLS (which would have been set by TLS-terminating proxy)
			STSSeconds:           315360000,                                       // 1yr HSTS time, as is generally recommended
			STSIncludeSubdomains: false,                                           // don't include subdomains; it may not be correct in general case TODO allow config
			STSPreload:           false,                                           // can't know if this is correct for general case TODO allow config
			FrameDeny:            true,                                            // proxy is write-only, no reason this should ever happen
			ContentTypeNosniff:   true,                                            // again, write-only
			BrowserXssFilter:     true,                                            // again, write-only
		})

		mux.Use(sec.Handler)
	}

	mux.Post("/github/push", githubIngestor(cl, cmd))

	if useTLS {
		graceful.ListenAndServeTLS(s.bind+":"+strconv.Itoa(s.port), s.cert, s.key, mux)
	} else {
		graceful.ListenAndServe(s.bind+":"+strconv.Itoa(s.port), mux)
	}
}

func statusIsOK(r *http.Response) bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}