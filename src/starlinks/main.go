package starlinks

import (
	log "github.com/Sirupsen/logrus"
	"net/http"
	"time"
)

func init() {
	StorageBackend["mysql"] = NewMySQLLinkStorage
}

func Main() {
	var link_handler *LinkRequestHandler
	var api_handler *APIHandler

	opts, err := parseArgs()
	if err != nil || opts == nil {
		return
	}

	defer func() {
		if err != nil {
			log.WithFields(log.Fields{
				"event": "init",
			}).Error(err.Error())
		}
	}()

	if link_handler, err = NewLinkRequestHandler(opts.RedisDail, opts.SQLDail, opts.SQLType, opts.Secret); err != nil {
		return
	}

	link_server := http.Server{
		Addr:           opts.Listen,
		Handler:        link_handler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	go func() error {
		return link_server.ListenAndServe()
	}()

	log.WithFields(log.Fields{
		"event": "init",
	}).Infof("Serve link requests at (%v) %v", opts.NetType, opts.Listen)

	api_handler, err = NewAPIHandler(opts.RedisDail, opts.SQLDail, opts.SQLType, opts.Secret)
	api_server := http.Server{
		Addr:           opts.APIListen,
		Handler:        api_handler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.WithFields(log.Fields{
		"event": "init",
	}).Infof("API Serve at (%v) %v", opts.APINetType, opts.APIListen)

	err = api_server.ListenAndServe()
}
