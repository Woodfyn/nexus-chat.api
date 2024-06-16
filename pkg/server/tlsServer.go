package server

import (
	"context"
	"net/http"
)

type TLSServer struct {
	muxServer *http.ServeMux
}

func (s *TLSServer) RunTLSServer(port string, handler http.Handler) error {
	return nil
}

func (s *TLSServer) ShutdownTLSServer(ctx context.Context) error {
	return nil
}
