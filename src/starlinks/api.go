package starlinks

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type APIHandler struct {
	storage LinkStorage
	secret  string
	gin     *gin.Engine
}

const (
	ERR_SUCCEED        = 0
	ERR_INVALID_PARAMS = 0101
	ERR_UNKNOWN        = 0500
)

var (
	ERR_DESP map[uint]string = map[uint]string{
		ERR_SUCCEED:        "Succeed.",
		ERR_INVALID_PARAMS: "Invalid parameters.",
		ERR_UNKNOWN:        "Server encounters unknown issue.",
	}
)

func NewAPIHandler(redis, sql, sqltype, secret string) (*APIHandler, error) {
	var err error
	var ok bool
	//var redis_path, redis_net_type string
	var factory BackendFactory
	var cache CacheStorage

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
	if cache, err = NewRedisLinkCache(redis); err != nil {
		return nil, err
	}
	hdr.storage = NewCachedStorage(cache, hdr.storage)

	hdr.gin = gin.New()
	hdr.route_register()
	hdr.secret = secret

	return hdr, nil
}

func (hdr *APIHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	hdr.gin.ServeHTTP(writer, request)
}

func (hdr *APIHandler) welcome(c *gin.Context) {
	c.String(200, "Starstudio short link service.")
}

func (hdr *APIHandler) add_link(c *gin.Context) {
	var err error
	var id LinkID
	var ok bool
	var response struct {
		ErrorCode  int
		ErrorDesc  string
		ShortRoute string
		ID         string
	}

	// Handle JSON response
	defer func() {
		if err != nil {
			response.ErrorCode = ERR_UNKNOWN
		}
		response.ErrorDesc, ok = ERR_DESP[uint(response.ErrorCode)]
		if !ok {
			c.Error(errors.New("Unknown error code."))
			c.Abort()
		}
		c.JSON(200, response)
	}()

	process_params := func() (string, error) {
		err := c.Request.ParseForm()
		if err != nil {
			return "", nil
		}
		new_links, ok := c.Request.Form["URL"]

		// TODO : Complete URL validity check.
		if len(c.Request.Form) != 1 || len(new_links) != 1 || !ok {
			response.ErrorCode = ERR_INVALID_PARAMS
			return "", errors.New("Too many URLs")
		}
		if new_links[0] != "" {
			return new_links[0], nil
		}
		response.ErrorCode = ERR_INVALID_PARAMS
		return "", errors.New("Links too short.")
	}

	links, err := process_params()
	if err != nil {
		return
	}

	id, err = hdr.storage.AddLink(links)
	if err != nil {
		return
	}
	response.ID = id.ToString()
	response.ErrorCode = 0
	id.Transform([]byte(hdr.secret))
	response.ShortRoute = "/" + id.ToTextEncodedBinary()
}

func (hdr *APIHandler) route_register() {
	hdr.gin.Use(gin.Recovery())
	hdr.gin.GET("/", hdr.welcome)
	hdr.gin.POST("/v1/links", hdr.add_link)
}
