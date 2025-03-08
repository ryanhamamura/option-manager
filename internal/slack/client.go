package slack

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"option-manager/internal/models"
	"option-manager/internal/storage"

	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"
	"golang.org/x/sync/errgroup"
)

// TradeAlertPattern is a regular expression for matching trade alerts
// This is a simplified example - you'll need to adjust based on your actual trade alert format
var TradeAlertPattern = regexp.MustCompile(`(?i)([A-Z]+)\s+(CALL|PUT)\s+\$?(\d+(\.\d+)?)\s+([\d/]+)`)

// SlackMonitor watches for trade alerts from a Slack channel
type SlackMonitor struct {
	client     *slack.Client
	channelID  string
	tradeStore storage.TradeStore
	useRTM     bool
	stopChan   chan struct{}
	wg         sync.WaitGroup
}

// NewSlackMonitor creates a new Slack monitor
func NewSlackMonitor(token, channelID string, tradeStore storage.TradeStore, useRTM bool) *SlackMonitor {
	client := slack.New(token)
	return &SlackMonitor{
		client:     client,
		channelID:  channelID,
		tradeStore: tradeStore,
		useRTM:     useRTM,
		stopChan:   make(chan struct{}),
	}
}

// Start begins monitoring the Slack channel for trade alerts
func (s *SlackMonitor) Start(ctx context.Context) error {
	if s.useRTM {
		return s.startRTM(ctx)
	}
	return s.startPolling(ctx)
}

// Stop gracefully shuts down the Slack monitor
func (s *SlackMonitor) Stop() {
	close(s.stopChan)
	s.wg.Wait()
}

// startRTM uses the Slack Real Time Messaging API to listen for messages
func (s *SlackMonitor) startRTM(ctx context.Context) error {
	rtm := s.client.NewRTM()

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		go rtm.ManageConnection()

		for {
			select {
			case msg := <-rtm.IncomingEvents:
				switch ev := msg.Data.(type) {
				case *slack.MessageEvent:
					if ev.Channel == s.channelID {
						s.processMessage(ev.Text, ev.Timestamp)
					}
				case *slack.RTMError:
					log.Error().Err(ev).Msg("Slack RTM error")
				case *slack.DisconnectedEvent:
					log.Info().Msg("Disconnected from Slack RTM")
				}
			case <-s.stopChan:
				rtm.Disconnect()
				return
			case <-ctx.Done():
				rtm.Disconnect()
				return
			}
		}
	}()

	return nil
}

// startPolling polls the Slack API periodically for new messages
func (s *SlackMonitor) startPolling(ctx context.Context) error {
	ticker := time.NewTicker(10 * time.Second)

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		defer ticker.Stop()

		var lastTimestamp string

		for {
			select {
			case <-ticker.C:
				// Get history
				params := slack.NewHistoryParameters()
				params.Count = 50
				if lastTimestamp != "" {
					params.Oldest = lastTimestamp
				}

				history, err := s.client.GetConversationHistory(&slack.GetConversationHistoryParameters{
					ChannelID: s.channelID,
					Limit:     50,
					Oldest:    lastTimestamp,
				})
				if err != nil {
					log.Error().Err(err).Msg("Failed to get channel history")
					continue
				}

				// Process messages in parallel for efficiency
				g, gctx := errgroup.WithContext(ctx)

				for _, msg := range history.Messages {
					if msg.Timestamp > lastTimestamp {
						lastTimestamp = msg.Timestamp
					}

					msg := msg // Create local copy for goroutine
					g.Go(func() error {
						select {
						case <-gctx.Done():
							return gctx.Err()
						default:
							s.processMessage(msg.Text, msg.Timestamp)
							return nil
						}
					})
				}

				if err := g.Wait(); err != nil {
					log.Error().Err(err).Msg("Error processing messages")
				}

			case <-s.stopChan:
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

// processMessage analyzes a Slack message to determine if it's a trade alert
func (s *SlackMonitor) processMessage(text, timestamp string) {
	// Skip if not a trade alert
	if !s.isTradeAlert(text) {
		return
	}

	trade, err := s.parseTradeAlert(text, timestamp)
	if err != nil {
		log.Error().Err(err).Str("message", text).Msg("Failed to parse trade alert")
		return
	}

	if err := s.tradeStore.Create(trade); err != nil {
		log.Error().Err(err).Interface("trade", trade).Msg("Failed to store trade")
	} else {
		log.Info().Str("id", trade.ID).Str("symbol", trade.Symbol).Msg("New trade alert detected")
	}
}

// isTradeAlert determines if a message contains a trade alert
func (s *SlackMonitor) isTradeAlert(text string) bool {
	return TradeAlertPattern.MatchString(text)
}

// parseTradeAlert extracts trade information from a message
func (s *SlackMonitor) parseTradeAlert(text, timestamp string) (*models.Trade, error) {
	matches := TradeAlertPattern.FindStringSubmatch(text)
	if len(matches) < 6 {
		return nil, fmt.Errorf("invalid trade alert format")
	}

	// Extract trade details
	symbol := matches[1]
	tradeType := models.TradeType(strings.ToUpper(matches[2]))

	strike, err := strconv.ParseFloat(matches[3], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid strike price: %w", err)
	}

	// Parse expiration date (simplified - adjust based on your actual format)
	expDate := matches[5]
	expiration, err := time.Parse("01/02/06", expDate)
	if err != nil {
		// Try alternative format
		expiration, err = time.Parse("01/02/2006", expDate)
		if err != nil {
			return nil, fmt.Errorf("invalid expiration date: %w", err)
		}
	}

	// Create the trade
	trade := &models.Trade{
		ID:          timestamp, // Use Slack timestamp as unique ID
		Symbol:      symbol,
		TradeType:   tradeType,
		Strike:      strike,
		Expiration:  expiration,
		Price:       0, // Price might be in another part of the message
		Status:      models.Pending,
		Description: fmt.Sprintf("%s %s $%.2f %s", symbol, tradeType, strike, expiration.Format("Jan 2")),
		RawText:     text, // Store the original text
	}

	return trade, nil
}
