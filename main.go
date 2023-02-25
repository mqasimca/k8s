package main

import (
	"context"
	"sync"

	"github.com/mqasimca/k8s/client"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

var (
	counterIncremental = map[string]map[string]int{
		"pods": {
			"total":   0,
			"deleted": 0,
			"error":   0,
		},
	}
)

type Pod struct {
	PodName   string `json:"podName"`
	Namespace string `json:"namespace"`
	Reason    string `json:"reason"`
	State     string `json:"state"`
}

func main() {
	var wg sync.WaitGroup
	ctx := context.Background()
	ch := make(chan Pod)

	kClient, err := client.NewK8sClient()
	if err != nil {
		log.Panic().Err(err)
	}

	// List pods
	wg.Add(1)
	go func() {
		err = listPods(ctx, kClient, ch)
		if err != nil {
			log.Panic().Err(err).Msg("Failed to list pods")
		}
	}()

	// PODS
	go func() {
		for {
			p, ok := <-ch
			if !ok {
				log.Info().Msgf("Total pods found: %d", counterIncremental["pods"]["total"])
				log.Info().Msg("Finished deleting pods")
				wg.Done()
				return
			}

			log.Info().Fields(
				map[string]interface{}{
					"action":    "delete_pod",
					"podName":   p.PodName,
					"namespace": p.Namespace,
					"state":     p.State,
					"reason":    p.Reason,
				},
			).Msg("pod")
			if viper.GetBool("ENABLE_DELETE_PODS") {
				counterIncremental["pods"]["delete"]++
				log.Info().Msg("Deleing pod")
				// Delete pods
			}
		}
	}()

	wg.Wait()

}

func init() {
	viper.AutomaticEnv()

	// Deleting pods is disable
	viper.SetDefault("CREATED_AT", 5)
	viper.SetDefault("ENABLE_DELETE_PODS", false)
}
