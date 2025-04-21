// Package metrics provides standardized Prometheus metrics instrumentation
// for all CodeCourt services.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Notification service specific metrics
var (
	// NotificationsSentTotal counts the total number of notifications sent
	NotificationsSentTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "codecourt",
			Subsystem: "notification",
			Name:      "sent_total",
			Help:      "Total number of notifications sent",
		},
		[]string{"channel", "type", "status"},
	)

	// NotificationDeliveryTime observes the time taken to deliver notifications
	NotificationDeliveryTime = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "codecourt",
			Subsystem: "notification",
			Name:      "delivery_seconds",
			Help:      "Time taken to deliver notifications",
			Buckets:   []float64{0.1, 0.5, 1.0, 2.5, 5.0, 10.0, 30.0, 60.0},
		},
		[]string{"channel", "type"},
	)

	// EventProcessingTotal counts the total number of events processed
	EventProcessingTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "codecourt",
			Subsystem: "notification",
			Name:      "events_processed_total",
			Help:      "Total number of events processed",
		},
		[]string{"event_type", "status"},
	)

	// EventProcessingTime observes the time taken to process events
	EventProcessingTime = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "codecourt",
			Subsystem: "notification",
			Name:      "event_processing_seconds",
			Help:      "Time taken to process events",
			Buckets:   []float64{0.01, 0.05, 0.1, 0.5, 1.0, 2.0, 5.0},
		},
		[]string{"event_type"},
	)

	// NotificationQueueLength tracks the current length of the notification queue
	NotificationQueueLength = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "codecourt",
			Subsystem: "notification",
			Name:      "queue_length",
			Help:      "Current length of the notification queue",
		},
	)

	// TemplateRenderingTime observes the time taken to render notification templates
	TemplateRenderingTime = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "codecourt",
			Subsystem: "notification",
			Name:      "template_rendering_seconds",
			Help:      "Time taken to render notification templates",
			Buckets:   []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5},
		},
		[]string{"template_type"},
	)
)

// RecordNotificationSent records a notification being sent
func RecordNotificationSent(channel, notificationType, status string) {
	NotificationsSentTotal.WithLabelValues(channel, notificationType, status).Inc()
}

// ObserveNotificationDeliveryTime observes the time taken to deliver a notification
func ObserveNotificationDeliveryTime(channel, notificationType string, duration float64) {
	NotificationDeliveryTime.WithLabelValues(channel, notificationType).Observe(duration)
}

// RecordEventProcessing records an event being processed
func RecordEventProcessing(eventType, status string) {
	EventProcessingTotal.WithLabelValues(eventType, status).Inc()
}

// ObserveEventProcessingTime observes the time taken to process an event
func ObserveEventProcessingTime(eventType string, duration float64) {
	EventProcessingTime.WithLabelValues(eventType).Observe(duration)
}

// SetNotificationQueueLength sets the current length of the notification queue
func SetNotificationQueueLength(length int) {
	NotificationQueueLength.Set(float64(length))
}

// ObserveTemplateRenderingTime observes the time taken to render a notification template
func ObserveTemplateRenderingTime(templateType string, duration float64) {
	TemplateRenderingTime.WithLabelValues(templateType).Observe(duration)
}
