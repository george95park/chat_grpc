package main

import (
    "fmt"
    "log"
    "io"
    "os"
    "bufio"
    "golang.org/x/net/context"
    "google.golang.org/grpc"
    pb "chat_grpc/chat"
)
var username string
const addr = "localhost:9090"

func listenForMessages(stream pb.ChatService_CreateStreamClient, chat chan*pb.ChatMessage) {
    for {
        in, err := stream.Recv()
        if err == io.EOF {
            break
        }
        if err != nil {
            log.Fatalf("Failed to receive message: %v", err)
        }
        chat <- in
    }
}

func handleInput(scanner *bufio.Scanner, msg chan *pb.ChatMessage) {
    for scanner.Scan() {
        m := scanner.Text()
        req := &pb.ChatMessage {
            From: username,
            Message: m,
        }
        msg <- req
    }
}

func connectUser(ctx context.Context, scanner *bufio.Scanner, c pb.ChatServiceClient) (pb.ChatService_CreateStreamClient) {
    fmt.Println("Enter an username: ")
    scanner.Scan()
    userInput := scanner.Text()
    u := &pb.Connect{
        Name: userInput,
    }
    username = userInput
    stream, err := c.CreateStream(ctx, u)
    if err != nil {
        log.Fatalf("error creating stream: %v", err)
    }
    fmt.Printf("Connected\n")
    return stream
}

func main() {
    msg := make(chan *pb.ChatMessage)
    chat := make(chan *pb.ChatMessage)

    conn, err := grpc.Dial(addr, grpc.WithInsecure())
    fmt.Println("Connecting to gRPC server...")
    if err != nil {
        log.Fatalf("Did not connect: %v", err)
    }
    fmt.Println("Connected to gRPC!")
    defer conn.Close()

    client := pb.NewChatServiceClient(conn)
    ctx := context.Background()
    scanner := bufio.NewScanner(os.Stdin)
    stream := connectUser(ctx, scanner, client)
    go listenForMessages(stream, chat)
    go handleInput(scanner, msg)

    for {
        select {
            case m := <-msg:
                if _,err := client.BroadcastMessage(ctx, m); err != nil {
                    log.Fatalf("Failed to send message: %v", err)
                }
            case c := <-chat:
                fmt.Printf("> %v: %v\n", c.From, c.Message)
        }
    }
}
