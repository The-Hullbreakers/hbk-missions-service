package sns_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	missionsns "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/sns"
)

// fakeSNS records Publish calls; SNS is a form-encoded HTTP API, so an
// httptest server is a faithful local stand-in for the boundary contract.
type fakeSNS struct {
	server *httptest.Server
	bodies []url.Values
}

func newFakeSNS(t *testing.T, status int) *fakeSNS {
	t.Helper()
	f := &fakeSNS{}
	f.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		values, err := url.ParseQuery(string(raw))
		require.NoError(t, err)
		f.bodies = append(f.bodies, values)
		w.Header().Set("Content-Type", "text/xml")
		w.WriteHeader(status)
		_, _ = io.WriteString(w, `<PublishResponse xmlns="http://sns.amazonaws.com/doc/2010-03-31/"><PublishResult><MessageId>msg-1</MessageId></PublishResult><ResponseMetadata><RequestId>req-1</RequestId></ResponseMetadata></PublishResponse>`)
	}))
	t.Cleanup(f.server.Close)
	return f
}

// messageAttributes decodes the Query-protocol MessageAttributes.entry.N
// pairs into a name -> StringValue map.
func messageAttributes(t *testing.T, body url.Values) map[string]string {
	t.Helper()
	attrs := map[string]string{}
	for i := 1; ; i++ {
		name := body.Get("MessageAttributes.entry." + strconv.Itoa(i) + ".Name")
		if name == "" {
			break
		}
		attrs[name] = body.Get("MessageAttributes.entry." + strconv.Itoa(i) + ".Value.StringValue")
	}
	return attrs
}

func TestDomainEventPublisher_PublishesEnvelopeWithFilterAttributes(t *testing.T) {
	sns := newFakeSNS(t, http.StatusOK)
	publisher := missionsns.NewPublisher(sns.server.URL, "arn:aws:sns:us-east-1:000000000000:domain-events")
	intent := application.PendingIntent{
		EventID:   "evt-1",
		EventType: "missions.season.placement-awarded",
		Envelope:  `{"event_id":"evt-1","event_type":"missions.season.placement-awarded","event_version":"1.0"}`,
	}

	err := publisher.Publish(context.Background(), intent)

	require.NoError(t, err)
	require.Len(t, sns.bodies, 1)
	body := sns.bodies[0]
	require.Equal(t, "Publish", body.Get("Action"))
	require.Equal(t, "arn:aws:sns:us-east-1:000000000000:domain-events", body.Get("TopicArn"))
	require.Equal(t, intent.Envelope, body.Get("Message"))
	require.Equal(t, map[string]string{
		"event_type": "missions.season.placement-awarded",
		"source":     "hbk-missions-service",
	}, messageAttributes(t, body))
}

func TestDomainEventPublisher_SNSFailureReturnsError(t *testing.T) {
	sns := newFakeSNS(t, http.StatusBadRequest)
	publisher := missionsns.NewPublisher(sns.server.URL, "arn:aws:sns:us-east-1:000000000000:domain-events")
	intent := application.PendingIntent{
		EventID:   "evt-1",
		EventType: "missions.season.placement-awarded",
		Envelope:  `{"event_id":"evt-1"}`,
	}

	err := publisher.Publish(context.Background(), intent)

	require.Error(t, err)
}
