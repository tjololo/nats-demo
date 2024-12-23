package publisher

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
)

type PublisherConfig struct {
	NatsServerURL string
	DefaultSubject string
	Port int16
	Every time.Duration
	ReplySubject string
}

func Publisher(config PublisherConfig) {
	router := gin.Default()
	nc, err := nats.Connect(config.NatsServerURL)
	if err != nil {
		log.Fatalf("Failed to connect to nats server %s\n", err)
	}
	defer nc.Close()

	router.POST("/publish", func(c *gin.Context) {
		now := time.Now().Format("2006-12-13 15:04:05")
		sub := c.Query("subject")
		if sub == "" {
			sub = config.DefaultSubject
		}
		message := c.Query("message")
		if message == "" {
			message = "New message published at " + now
		}
		err = nc.Publish(sub, []byte(message))
		if err != nil {
			log.Printf("Failed to publish message %s\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish message"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Message published"})
	})

	router.POST("/request", func(c *gin.Context) {
		now := time.Now().Format("2006-12-13 15:04:05")
		sub := c.Query("subject")
		if sub == "" {
			sub = config.DefaultSubject
		}
		message := c.Query("message")
		if message == "" {
			message = "New message published at " + now
		}
		msg, err := nc.Request(sub, []byte(message), 2*time.Second)
		if err != nil {
			log.Printf("Failed to request message %s\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to request message"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": string(msg.Data)})
	})

	srv := &http.Server{
		Addr:    ":" + strconv.Itoa(int(config.Port)),
		Handler: router.Handler(),
	}

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	running := true
	if config.Every != 0 {
		log.Printf("Publishing message every %s\n", config.Every)
		go func() {
			for running {
				time.Sleep(config.Every)
				err = nc.Publish(config.DefaultSubject, []byte("Automatic message published at " + time.Now().Format("2006-12-13 15:04:05")))
				if err != nil {
					log.Printf("Failed to publish message %s\n", err)
				}
			}
			log.Println("Stopped publishing messages")
		}()
	}

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall. SIGKILL but can"t be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown Server ...")
	timeout := config.Every + 5*time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	running = false
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	// catching ctxTimeout.Done(). timeout of 5 seconds.
	select {
	case <-ctx.Done():
		log.Printf("Stopped after wating for %s timeout\n", timeout)
	}
	log.Println("Server exiting")
}
