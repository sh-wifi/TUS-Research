package main

import (
        "context"
        "crypto/tls"
        "io"
        "log"
        "net/http"
        "os/exec"

        "github.com/quic-go/quic-go/http3"
        "github.com/quic-go/webtransport-go"
)

func main() {
        cert, err := tls.LoadX509KeyPair("server.crt", "server.key")
        if err != nil {
                log.Fatalf("Failed to load TLS certificate: %v", err)
        }

        server := webtransport.Server{
                H3: http3.Server{
                        Addr: ":4433",
                        TLSConfig: &tls.Config{
                                Certificates: []tls.Certificate{cert},
                                NextProtos:   []string{"h3"},
                        },
                },
                CheckOrigin: func(r *http.Request) bool {
                        log.Println("Origin:", r.Header.Get("Origin"))
                        return true
                },
        }

        http.HandleFunc("/webtransport", func(w http.ResponseWriter, r *http.Request){
                if r.Method != http.MethodConnect {
                        http.Error(w, "WebTransport requires CONNECT method", http.Stat>
                        return
                }

                session, err := server.Upgrade(w, r)
                if err != nil {
                        log.Printf("Upgrade failed: %v", err)
                        return
                }

                go func() {
                        for {
                                stream, err := session.AcceptStream(context.Background(>
                                if err != nil {
                                        log.Printf("Failed to accept stream: %v", err)
                                        return
                                }
                                go handleStream(stream, session)
                        }
                }()
        })

       log.Println("WebTransport server starting on https://127.0.0.1:4433")
        if err := server.ListenAndServe(); err != nil {
                log.Fatalf("Server failed: %v", err)
        }
}

func handleStream(stream webtransport.Stream, session *webtransport.Session) {
        data, err := io.ReadAll(stream)
        if err != nil {
                log.Printf("Failed to read from stream: %v", err)
                return
        }
        message := string(data)
        log.Println("Received message:", message)
        cmd := exec.Command("python3", "run_blindx.py", message)
        output, err := cmd.Output()
        if err != nil {
                log.Printf("Inference error: %v", err)
                return
        }
        result := string(output)
        log.Println("Sent result:", result)

        uni, err := session.OpenUniStream()
        if err != nil {
                log.Printf("Failed to open unidirectional stream: %v", err)
                return
        }
        uni.Write([]byte(result))
        uni.Close()
}
