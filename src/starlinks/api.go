package starlinks

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type APIHandler struct {
	storage LinkStorage
	gin     *gin.Engine
}

func NewAPIHandler(redis, sql, sqltype, secret string) (*APIHandler, error) {
	var err error
	var ok bool
	//var redis_path, redis_net_type string
	var factory BackendFactory

	hdr := new(APIHandler)
	//if redis_path, redis_net_type, err = parseNetPath(redis); err != nil {
	//	return nil, err
	//}

	if factory, ok = StorageBackend[sqltype]; !ok {
		return nil, fmt.Errorf("Not supported storage backend type: %v", sqltype)
	}
	if hdr.storage, err = factory(sql); err != nil {
		return nil, err
	}

	hdr.gin = gin.New()
	hdr.route_register()

	return hdr, nil
}

func (hdr *APIHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	hdr.gin.ServeHTTP(writer, request)
}

func (hdr *APIHandler) welcome(c *gin.Context) {
	c.String(200, "Starstudio short link service.")
}

func (hdr *APIHandler) route_register() {
	hdr.gin.Use(gin.Recovery())
	hdr.gin.GET("/", hdr.welcome)
}
