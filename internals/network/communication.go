package network

import (
   
    "log"
	"net"

    "google.golang.org/grpc"
)

type CommunicationService struct {
    server *grpc.Server
}

func NewCommunicationService() *CommunicationService {
    return &CommunicationService{
        server: grpc.NewServer(),
    }
}

func (cs *CommunicationService) Start(address string) error {
    lis, err := net.Listen("tcp", address)
    if err != nil {
        return err
    }

    log.Printf("Server listening on %v", address)
    return cs.server.Serve(lis)
}