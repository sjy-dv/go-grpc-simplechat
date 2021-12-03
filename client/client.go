package main

import (
	"bufio"
	"context"
	"fmt"
	"gigchatr/chat_v1"
	"os"
	"sync"
	"time"

	"github.com/teris-io/shortid"
	"google.golang.org/grpc"
)


var client chat_v1.BroadcastClient
var wait *sync.WaitGroup

func init() {
	wait = &sync.WaitGroup{}
}

func connect(user *chat_v1.User) error {
	var streamerror error

	stream, err := client.CreateStream(context.Background(), &chat_v1.Connect{
		User: user,
		Active: true,
	})

	if err != nil {
		fmt.Println("connection error")
	}
	wait.Add(1)

	go func(stream chat_v1.Broadcast_CreateStreamClient) {
		defer wait.Done()

		for {
			msg, err := stream.Recv()
			if err != nil {
				streamerror = fmt.Errorf("msg error : %v", err)
				break
			}

			fmt.Printf("%v : %s \n", msg.Id, msg.Message)
		}
	}(stream)

	return streamerror
}

func main() {
	timestamp := time.Now()
	done := make(chan int)

	name := shortid.MustGenerate()

	id := shortid.MustGenerate()

	conn, _ := grpc.Dial(":50051", grpc.WithInsecure())

	client = chat_v1.NewBroadcastClient(conn)
	user := &chat_v1.User{
		Id: id,
		DisplayName: name,
	}

	connect(user)

	wait.Add(1)

	go func() {
		defer wait.Done()

		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			msg := &chat_v1.Message{
				Id: user.Id,
				Message: scanner.Text(),
				Timestamp: timestamp.String(),
			}

			_, err := client.BroadcastMessage(context.Background(), msg)
			if err != nil {
				fmt.Println("send msg error : ", err)
				break
			}
		}
	}()

	go func() {
		wait.Wait()
		close(done)
	}()

	<- done
}