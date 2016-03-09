package collector

import (
	"log"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	allTopics []string
)

type topicStats []struct {
	val func(*topic) float64
	vec *prometheus.GaugeVec
}

// TopicStats creates a new stats collector which is able to
// expose the topic metrics of a nsqd node to Prometheus.
func TopicStats(namespace string) StatsCollector {
	labels := []string{"type", "topic", "paused"}

	return topicStats{
		{
			val: func(t *topic) float64 { return float64(len(t.Channels)) },
			vec: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "channel_count",
				Help:      "Number of channels",
			}, labels),
		},
		{
			val: func(t *topic) float64 { return float64(t.Depth) },
			vec: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "depth",
				Help:      "Queue depth",
			}, labels),
		},
		{
			val: func(t *topic) float64 { return float64(t.BackendDepth) },
			vec: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "backend_depth",
				Help:      "Queue backend depth",
			}, labels),
		},
		{
			val: func(t *topic) float64 { return float64(t.MessageCount) },
			vec: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "message_count",
				Help:      "Queue message count",
			}, labels),
		},
	}
}

func (ts topicStats) collect(s *stats, out chan<- prometheus.Metric) {
	// Exit if any "dead" topics are detected - Docker will restart the container
	for _, topicName := range allTopics {
		found := false
		for _, topic := range s.Topics {
			if topicName == topic.Name {
				found = true
				break
			}
		}
		if !found {
			log.Fatal("At least one old topic no longer included in nsqd stats - exiting")
		}
	}

	allTopics = nil // Rebuild list of all topics
	for _, topic := range s.Topics {
		allTopics = append(allTopics, topic.Name)

		labels := prometheus.Labels{
			"type":   "topic",
			"topic":  topic.Name,
			"paused": strconv.FormatBool(topic.Paused),
		}

		for _, c := range ts {
			c.vec.With(labels).Set(c.val(topic))
			c.vec.Collect(out)
		}
	}
}
