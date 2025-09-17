package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

func main() {
	defer fmt.Println("done")
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ltime | log.LUTC)

	ctx := context.Background()
	apiConn := open()

	var wg sync.WaitGroup
	wg.Add(20)
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			err := apiConn.readFile(ctx)
			if err != nil {
				log.Printf("could not read file: %v\n", err)
			}
			log.Println("read file")
		}()
	}
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			err := apiConn.httpCall(ctx)
			if err != nil {
				log.Printf("could not make http call: %v\n", err)
			}
			log.Println("http call")
		}()
	}

	wg.Wait()
}

type RateLimiter interface {
	Wait(ctx context.Context) error
	Limit() rate.Limit
}

func newMultiLimiter(limiters ...RateLimiter) *multiLimiter {
	byLimit := func(i, j int) bool {
		return limiters[i].Limit() < limiters[j].Limit()
	}
	sort.Slice(limiters, byLimit)
	return &multiLimiter{
		limiters: limiters,
	}
}

type multiLimiter struct {
	limiters []RateLimiter
}

func (l multiLimiter) Wait(ctx context.Context) error {
	for _, limiter := range l.limiters {
		if err := limiter.Wait(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (l multiLimiter) Limit() rate.Limit {
	return l.limiters[0].Limit()
}

func open() *apiConnection {
	per := func(eventCount int, duration time.Duration) rate.Limit {
		return rate.Every(duration / time.Duration(eventCount))
	}
	secondLimit := rate.NewLimiter(per(3, time.Second), 1)
	minuteLimit := rate.NewLimiter(per(10, time.Minute), 10)
	return &apiConnection{
		rateLimiter: newMultiLimiter(secondLimit, minuteLimit),
	}
}

type apiConnection struct {
	rateLimiter RateLimiter
}

func (a *apiConnection) readFile(ctx context.Context) error {
	if err := a.rateLimiter.Wait(ctx); err != nil {
		return err
	}
	// do some work here
	return nil
}

func (a *apiConnection) httpCall(ctx context.Context) error {
	if err := a.rateLimiter.Wait(ctx); err != nil {
		return err
	}
	// do some work here
	return nil
}
