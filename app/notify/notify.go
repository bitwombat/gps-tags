package notify

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

const ntfyURL = "http://ntfy.sh/ahShaikee"

func Notify(ctx context.Context, message string) error {
	// Set up client and crap
	client := &http.Client{}
	buf := strings.NewReader(message)

	// Make the request object
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ntfyURL, buf)
	if err != nil {
		return fmt.Errorf("error while making http POST request to ntfy.sh: %v", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Now actually send the request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error while sending request to ntfy.sh: %v", err)
	}
	defer resp.Body.Close()

	// Check the response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error returned in response from ntfy.sh: %d", resp.StatusCode)
	}

	return nil
}
