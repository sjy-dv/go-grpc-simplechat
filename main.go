package main

import (
	"context"
	"gigchatr/chat_v1"
	"log"
	"net"
	"sync"

	"google.golang.org/grpc"
)

type Connection struct {
	stream chat_v1.Broadcast_CreateStreamServer
	id     string
	active bool
	error  chan error
}

type Server struct {
	Connection []*Connection
}


func (s *Server) CreateStream(pconn *chat_v1.Connect, stream chat_v1.Broadcast_CreateStreamServer) error {
	conn := &Connection{
		stream: stream,
		id : pconn.User.Id,
		active : true,
		error : make(chan error),
	}

	s.Connection = append(s.Connection, conn)

	return <- conn.error
}

func (s *Server) BroadcastMessage(ctx context.Context, msg *chat_v1.Message) (*chat_v1.Close, error) {

	wait := sync.WaitGroup{}
	done := make(chan int)

	for _, conn := range s.Connection {
		wait.Add(1)

		go func(msg *chat_v1.Message, conn *Connection) {

			defer wait.Done()

			if conn.active {
				conn.stream.Send(msg)
				
			}
		} (msg, conn)
	}

	go func() {
		wait.Wait()
		close(done)
	}()

	<- done

	return &chat_v1.Close{}, nil
}


func main() {
	var connections []*Connection

	server := &Server{connections}

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer()


	chat_v1.RegisterBroadcastServer(s, server)

	s.Serve(lis)
}