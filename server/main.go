package main

import (
    "fmt"
    "log"
    "net"
    "google.golang.org/grpc"
    "golang.org/x/net/context"
    pb "chat_grpc/chat"
)

const port = ":9090"

type user struct {
    stream pb.ChatService_CreateStreamServer
    name string
    error chan error
}

type server struct {
    observers []*user
}

func (s* server) CreateStream(conn *pb.Connect, stream pb.ChatService_CreateStreamServer) error {
    fmt.Println(conn.Name)
    u := &user {
        stream: stream,
        name: conn.Name,
        error: make(chan error),
    }
    s.observers = append(s.observers, u)
    log.Printf("User: %v has connected", conn.Name)
    return <-u.error
}

func (s* server) BroadcastMessage(ctx context.Context, msg *pb.ChatMessage) (*pb.Empty, error) {
    for _,obs := range s.observers {
        err := obs.stream.Send(msg)
        if err != nil {
            return &pb.Empty{}, err
        }
        log.Printf("Broadcasting from: %v -- Message: %v", msg.From, msg.Message)
    }
    return &pb.Empty{}, nil
}

// func (s *server) Chat(stream pb.ChatService_ChatServer) error {
//     waitc := make(chan *pb.ChatMessage)
//     go func() {
//         for {
//             req, err := stream.Recv()
//             if err == io.EOF {
//                 close(waitc)
//                 break
//             }
//             if err != nil {
//                 log.Fatalf("error receiving: ", err)
//                 break
//             }
//             log.Printf("received request: %v", req)
//             if s.observers == nil {
//                 s.observers = make(map[pb.ChatService_ChatServer]bool)
//             }
//             s.observers[stream] = true
//             waitc <- req
//         }
//     }()
//     for m := range waitc {
//         for obs,_ := range s.observers {
//             err := obs.Send(&pb.ChatMessageFromServer{
//                 Message: &pb.ChatMessage{
//                     From: m.From,
//                     Message: m.Message,
//                 },
//             })
//             if err != nil {
//                 return err
//             }
//         }
//     }
//     return nil
// }

func main() {
    fmt.Printf("Starting server on port%v...\n", port)
    lis, err := net.Listen("tcp", port)
    if err != nil {
        log.Fatalf("Failed to listen: %v", err)
    }
    fmt.Println("Server started!")
    s := grpc.NewServer()
    pb.RegisterChatServiceServer(s, &server{})
    s.Serve(lis)
}
