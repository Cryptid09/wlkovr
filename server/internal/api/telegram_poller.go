package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type telegramUpdatesResponse struct {
	OK          bool             `json:"ok"`
	Result      []telegramUpdate `json:"result"`
	Description string           `json:"description"`
}

// StartTelegramPolling is the local-development path; production can use the
// secured webhook route instead.
func StartTelegramPolling(ctx context.Context, token string, handler *Handler) {
	if token == "" {
		return
	}
	client := &http.Client{Timeout: 35 * time.Second}
	var offset int64
	log.Printf("[TELEGRAM] Long polling enabled")
	for ctx.Err() == nil {
		query := url.Values{"timeout": {"25"}, "offset": {strconv.FormatInt(offset, 10)}, "allowed_updates": {`["message","edited_message","channel_post"]`}}
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?%s", token, query.Encode()), nil)
		if err != nil {
			return
		}
		response, err := client.Do(request)
		if err != nil {
			time.Sleep(3 * time.Second)
			continue
		}
		var result telegramUpdatesResponse
		decodeErr := json.NewDecoder(response.Body).Decode(&result)
		response.Body.Close()
		if decodeErr != nil || response.StatusCode != http.StatusOK || !result.OK {
			time.Sleep(3 * time.Second)
			continue
		}
		for _, update := range result.Result {
			if update.UpdateID >= offset {
				offset = update.UpdateID + 1
			}
			handler.ingestTelegramUpdate(update)
		}
	}
}
