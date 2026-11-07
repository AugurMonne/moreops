package alertmanager

// AlertmanagerIntegration provides helpers for Alertmanager receiver configuration.
type AlertmanagerIntegration struct {
	SlackWebhookURL string
	TelegramBotToken string
	TelegramChatID  string
}

func NewAlertmanagerIntegration() *AlertmanagerIntegration {
	return &AlertmanagerIntegration{}
}
