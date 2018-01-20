package commands

import (
	"context"
	"github.com/ipfs-search/ipfs-search/crawler/factory"
	"github.com/ipfs-search/ipfs-search/worker"
	"golang.org/x/sync/errgroup"
	"log"
)

// block blocks until context is cancelled
func block(ctx context.Context) error {
	for {
		<-ctx.Done()
		return ctx.Err()
	}
}

// log errors from errc
func errorLoop(errc chan error) {
	for {
		err := <-errc
		log.Printf("%T: %v", err, err)
	}
}

// Crawl configures and initializes crawling
func Crawl(ctx context.Context) error {
	config, err := getConfig()
	if err != nil {
		return err
	}

	errc := make(chan error, 1)

	factory, err := factory.New(config, errc)
	if err != nil {
		return err
	}

	hashGroup := worker.Group{
		Count:   config.HashWorkers,
		Factory: factory.NewHashWorker,
	}
	fileGroup := worker.Group{
		Count:   config.FileWorkers,
		Factory: factory.NewFileWorker,
	}

	// Create error group and context
	errg, ctx := errgroup.WithContext(ctx)

	// Start work loop
	errg.Go(func() error { return hashGroup.Work(ctx) })
	errg.Go(func() error { return fileGroup.Work(ctx) })

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")

	// Log messages, wait for context break
	go errorLoop(errc)
	err = block(ctx)

	log.Printf("Shutting down: %s", err)
	log.Print("Waiting for processes to finish")

	err = errg.Wait()
	log.Printf("Error group finished: %s", err)
	return err
}
