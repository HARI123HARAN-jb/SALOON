package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"saloon/database"
)

// SendCustomerStatusNotification constructs and triggers the WhatsApp notification to the customer.
func SendCustomerStatusNotification(booking database.Booking, status string) {
	// Let status notifications run in a goroutine so we do not block HTTP handlers.
	toPhone := booking.Phone
	if toPhone == "" {
		log.Println("Customer has no phone number. Skipping WhatsApp status notification.")
		return
	}

	// Normalize phone number (ensure country code, default to +91 if length is 10 digits and starts with non-+ or non-0)
	toPhone = formatPhoneNumber(toPhone)

	var message string
	switch status {
	case "confirmed":
		message = fmt.Sprintf(
			"✨ *APPOINTMENT CONFIRMED* ✨\n"+
				"-----------------------------\n"+
				"Hello %s,\n\n"+
				"We are excited to let you know that your appointment at *Guna Men's Salon & Beauty Parlour* has been confirmed!\n\n"+
				"✂️ *Service:* %s\n"+
				"📅 *Date:* %s\n"+
				"⏰ *Time:* %s\n\n"+
				"📍 *Location:* NSJ Building, Mappedu Kootroad, Sunguvachathiram Road\n"+
				"📞 *Contact:* +91 9600407380\n\n"+
				"Looking forward to welcoming you to the sanctuary!",
			booking.Name, booking.Service, booking.Date, booking.Time,
		)
	case "cancelled":
		message = fmt.Sprintf(
			"⚠️ *APPOINTMENT CANCELLED* ⚠️\n"+
				"-----------------------------\n"+
				"Hello %s,\n\n"+
				"Your appointment for *%s* scheduled on %s at %s has been cancelled.\n\n"+
				"If you have any questions or would like to reschedule, please feel free to reach out to us at +91 9600407380 or book a new appointment on our website.\n\n"+
				"Best regards,\n"+
				"Guna Men's Salon & Beauty Parlour",
			booking.Name, booking.Service, booking.Date, booking.Time,
		)
	default:
		// We only notify on confirmation or cancellation
		return
	}

	log.Printf("Dispatching WhatsApp status notification to customer %s (%s)...", booking.Name, toPhone)
	err := sendWhatsApp(toPhone, message)
	if err != nil {
		log.Printf("Error sending WhatsApp status update to %s: %v", toPhone, err)
	} else {
		log.Printf("Successfully sent WhatsApp status notification to %s", toPhone)
	}
}

// formatPhoneNumber formats raw inputs into standard E.164 phone numbers (defaults to India +91 if length is 10 digits)
func formatPhoneNumber(phone string) string {
	cleaned := ""
	for _, ch := range phone {
		if ch >= '0' && ch <= '9' {
			cleaned += string(ch)
		}
	}
	if len(cleaned) == 10 {
		return "+91" + cleaned
	}
	if len(cleaned) > 10 && !strings.HasPrefix(phone, "+") {
		return "+" + cleaned
	}
	return phone
}

// sendWhatsApp calls Twilio or Custom WhatsApp API depending on config.
func sendWhatsApp(to, body string) error {
	provider := os.Getenv("WHATSAPP_PROVIDER")
	if provider == "" {
		// Fallback detection
		if os.Getenv("TWILIO_ACCOUNT_SID") != "" {
			provider = "twilio"
		} else if os.Getenv("WHATSAPP_CUSTOM_URL") != "" {
			provider = "custom"
		} else {
			log.Println("No WHATSAPP_PROVIDER configured. Skipping WhatsApp sending.")
			return nil
		}
	}

	switch strings.ToLower(provider) {
	case "twilio":
		accountSid := os.Getenv("TWILIO_ACCOUNT_SID")
		authToken := os.Getenv("TWILIO_AUTH_TOKEN")
		fromNumber := os.Getenv("TWILIO_FROM_NUMBER") // E.g., "whatsapp:+14155238886"

		if accountSid == "" || authToken == "" || fromNumber == "" {
			return fmt.Errorf("missing Twilio WhatsApp credentials in environment")
		}

		// Ensure Twilio recipient matches "whatsapp:+XXXX"
		toTwilio := to
		if !strings.HasPrefix(toTwilio, "whatsapp:") {
			toTwilio = "whatsapp:" + toTwilio
		}

		apiURL := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", accountSid)
		data := url.Values{}
		data.Set("From", fromNumber)
		data.Set("To", toTwilio)
		data.Set("Body", body)

		req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
		if err != nil {
			return err
		}

		req.SetBasicAuth(accountSid, authToken)
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			var errResp map[string]interface{}
			_ = json.NewDecoder(resp.Body).Decode(&errResp)
			return fmt.Errorf("twilio API status %d: %v", resp.StatusCode, errResp)
		}
		return nil

	case "custom":
		apiURL := os.Getenv("WHATSAPP_CUSTOM_URL")
		token := os.Getenv("WHATSAPP_CUSTOM_TOKEN")
		if apiURL == "" {
			return fmt.Errorf("missing WHATSAPP_CUSTOM_URL for custom WhatsApp API")
		}

		format := os.Getenv("WHATSAPP_CUSTOM_FORMAT")
		if format == "" {
			format = "json"
		}

		var req *http.Request
		var err error

		if strings.ToLower(format) == "form" {
			data := url.Values{}
			data.Set("token", token)
			data.Set("to", to)
			data.Set("body", body)
			req, err = http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
			if err == nil {
				req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			}
		} else {
			// JSON payload: supporting various formats
			payload := map[string]string{
				"to":      to,
				"phone":   to,
				"body":    body,
				"message": body,
				"text":    body,
				"token":   token,
			}
			jsonData, jsonErr := json.Marshal(payload)
			if jsonErr != nil {
				return jsonErr
			}
			req, err = http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
			if err == nil {
				req.Header.Set("Content-Type", "application/json")
			}
		}

		if err != nil {
			return err
		}

		authHeader := os.Getenv("WHATSAPP_CUSTOM_AUTH_HEADER")
		if authHeader != "" {
			req.Header.Set("Authorization", authHeader)
		}

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return fmt.Errorf("custom WhatsApp API status %d", resp.StatusCode)
		}
		return nil

	default:
		return fmt.Errorf("unsupported WhatsApp provider: %s", provider)
	}
}
