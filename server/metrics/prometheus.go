package metrics

import (
	"os"
	"strings"
	"time"

	"github.com/jitsucom/jitsu/server/logging"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/spf13/viper"
)

var Exported bool

var Registry *prometheus.Registry

func Enabled() bool {
	return Registry != nil
}

func NewCounterVec(opts prometheus.CounterOpts, labels []string) *prometheus.CounterVec {
	vec := prometheus.NewCounterVec(opts, labels)
	Registry.MustRegister(vec)
	return vec
}

func NewGaugeVec(opts prometheus.GaugeOpts, labels []string) *prometheus.GaugeVec {
	vec := prometheus.NewGaugeVec(opts, labels)
	Registry.MustRegister(vec)
	return vec
}

const Unknown = "unknown"

func Init(exported bool) {
	Exported = exported
	if Exported {
		logging.Info("✅ Initializing Prometheus metrics..")
	}

	Registry = prometheus.NewRegistry()
	Registry.MustRegister(
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		collectors.NewGoCollector(),
	)

	initEvents()
	initSourcesPool()
	initSourceObjects()
	initMetaRedis()
	initCoordinationRedis()
	initEventsRedis()
	initUsersRecognitionQueue()
	initUsersRecognitionRedis()
	initStreamEventsQueue()
}

func InitRelay(clusterID string, viper *viper.Viper) *Relay {
	relay := &Relay{
		URL:          DefaultRelayURL,
		HostID:       Unknown,
		DeploymentID: clusterID,
		Timeout:      time.Second / 10,
	}

	hostID, err := os.Hostname()
	if err != nil {
		logging.Debugf("Failed to get hostname for metrics relay, using '%s': %s", Unknown, err)
	} else {
		relay.HostID = hostID
	}

	if viper != nil {
		if viper.GetBool("disabled") {
			logging.Debugf("Metrics relay is disabled")
			return nil
		}

		url := viper.GetString("url")
		if url != "" {
			relay.URL = url
		}

		deploymentID := viper.GetString("deployment_id")
		if deploymentID != "" {
			relay.DeploymentID = deploymentID
		}

		if viper.IsSet("timeout") {
			relay.Timeout = viper.GetDuration("timeout")
		}
	}

	logging.Debugf("✅ Initialized metrics relay to %s as [host: %s, deployment: %s]",
		relay.URL, relay.HostID, relay.DeploymentID)
	return relay
}

func extractLabels(destinationName string) (projectID, destinationID string) {
	splitted := strings.Split(destinationName, ".")
	if len(splitted) > 1 {
		return splitted[0], splitted[1]
	}

	return "-", destinationName
}
