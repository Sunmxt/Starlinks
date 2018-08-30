package starlinks

import (
	"crypto/des"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	_ "github.com/go-sql-driver/mysql"
	"net/http"
)

type LinkStorage interface {
	QueryLink(id LinkID) (string, error)
	QueryLinks(id []LinkID) ([]string, error)

	AddLink(url string) (LinkID, error)
	AddLinks(urls []string) ([]LinkID, error)

	RemoveLink(id LinkID) error
	RemoveLinks(ids []LinkID) error
}

type CacheStorage interface {
	QueryLink(id LinkID) (string, error)
	QueryLinks(ids []LinkID) ([]string, error)

	AddLink(id LinkID, url string) error
	AddLinks(url_map map[LinkID]string) error

	RemoveLink(id LinkID) error
	RemoveLinks(ids []LinkID) error
}

type LinkRequestHandler struct {
	redis_path     string
	redis_net_type string
	sql_path       string
	sql_net_type   string
	sql_type       string
	secret         string

	storage LinkStorage
}

type BackendFactory func(string) (LinkStorage, error)

var (
	StorageBackend = make(map[string]BackendFactory)
)

type LinkID uint64 // Node: the zero value indicates invalid. // BigEndian

func (id *LinkID) Transform(map_scecrt []byte) error {
	blk, err := des.NewCipher(map_scecrt)
	if err != nil {
		return err
	}
	mapped := make([]byte, 8)
	src := make([]byte, 8)
	binary.BigEndian.PutUint64(mapped, uint64(*id))
	copy(src, mapped)
	blk.Encrypt(mapped, src)
	*id = LinkID(binary.BigEndian.Uint64(mapped))
	return nil
}

func (id *LinkID) ToString() string {
	return fmt.Sprintf("%v", *id)
}

func (id *LinkID) ReverseTransform(map_secret []byte) error {
	blk, err := des.NewCipher(map_secret)
	if err != nil {
		return err
	}
	mapped := make([]byte, 8)
	src := make([]byte, 8)
	binary.BigEndian.PutUint64(mapped, uint64(*id))
	copy(src, mapped)
	blk.Decrypt(mapped, src)
	*id = LinkID(binary.BigEndian.Uint64(src))
	return nil
}

func (id *LinkID) ToTextEncodedBinary() string {
	bin := make([]byte, 8)
	binary.BigEndian.PutUint64(bin, uint64(*id))
	encoded := base64.StdEncoding.EncodeToString(bin)
	return encoded[:11]
}

func FromTextEncodedBinary(text string) (LinkID, error) {
	ERR_INVALID := "Not a valid text-encoded link id."
	if len(text) != 11 {
		return 0, errors.New(ERR_INVALID)
	}

	decoded, err := base64.StdEncoding.DecodeString(text + "=")
	if err != nil {
		return 0, err
	}

	return LinkID(binary.BigEndian.Uint64(decoded)), nil
}

func NewLinkRequestHandler(redis, sql, sqltype, secret string) (*LinkRequestHandler, error) {
	var err error
	var ok bool
	var factory BackendFactory

	hdr := new(LinkRequestHandler)
	if hdr.redis_path, hdr.redis_net_type, err = parseNetPath(redis); err != nil {
		return nil, err
	}

	hdr.sql_type = sqltype

	if factory, ok = StorageBackend[sqltype]; !ok {
		return nil, fmt.Errorf("Not supported storage backend type: %v", sqltype)
	}
	if hdr.storage, err = factory(sql); err != nil {
		return nil, err
	}

	return hdr, nil
}

func (hdr *LinkRequestHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	var link string

	return_error := func() {
		http.Error(writer, "Server come across a unexpected error.", 500)
	}

	// Try decode
	id, err := FromTextEncodedBinary(request.RequestURI)
	if err != nil {
		log.WithFields(log.Fields{
			"event": "request",
		}).Infof("Invalid Links: %v", request.RequestURI)
		http.NotFound(writer, request)
		return
	}
	if err = id.ReverseTransform([]byte(hdr.secret)); err != nil {
		log.WithFields(log.Fields{
			"event":      "request",
			"encoded_id": id,
		}).Errorf("Cannot decode id with request: %v", request.RequestURI)
		return_error()
		return
	}

	// Query
	if link, err = hdr.storage.QueryLink(id); err != nil {
		log.WithFields(log.Fields{
			"event": "request",
			"id":    id,
		}).Errorf("Storage failure: %v", err.Error())
		return_error()
		return
	}

	http.Redirect(writer, request, link, 302)
}
