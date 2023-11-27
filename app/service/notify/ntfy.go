package notify

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

const ntfyURLBase = "https://ntfy.sh/"

type Ntfy struct {
	subscriptionId string
}

func NewNtfyNotifier(subscriptionId string) Notifier {
	return Ntfy{
		subscriptionId: subscriptionId,
	}
}

func (n Ntfy) Notify(ctx context.Context, title, message string) error {
	// Set up client and crap
	client := &http.Client{}
	buf := strings.NewReader(message)

	// Make the request object
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ntfyURLBase+n.subscriptionId, buf)
	if err != nil {
		return fmt.Errorf("error while making http POST request to ntfy.sh: %w", err)
	}
	req.Header.Set("Title", title)
	req.Header.Set("Actions", `[{ "action": "view", "label": "Show me", "url": "https://tags.bitwombat.com.au/current" }]`)

	// Now actually send the request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error while sending request to ntfy.sh: %w", err)
	}
	defer resp.Body.Close()

	// Check the response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error returned in response from ntfy.sh: %d", resp.StatusCode)
	}

	return nil
}