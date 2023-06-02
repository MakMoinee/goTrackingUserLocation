package common

import "goTrackingUserLocation/internal/models"

var (
	SERVER_PORT          string
	LOCATION_MAP         map[string]models.Location
	COUNTDOWN            = 1
	EMAIL_ADDRESS        string
	EMAIL_PORT           int
	EMAIL_APP_PASS       string
	EMAIL_HOST           string
	EMAIL_SUBJECT        string
	HISTORY_COUNT        = 1
	HISTORY_DELETE_COUNT = 0
	GOOGLE_MAP           string
)

const (
	DEVICES_REF = "devices"
	HISTORY_REF = "history"
	ALARM_MSG   = "Dear %s,\n" +
		"We hope this message finds you well. We are reaching out to inform you of a potential safety concern regarding Device ID %s, currently associated with the individual in question. It has come to our attention that the person using Device ID %s is currently online and could be in harm.\n" +
		"We kindly request your immediate attention and assistance in ensuring the well-being of the person connected to Device ID %s. As responsible members of our community, it is crucial that we come together to support those who may be in need.\n" +
		"If you have any information or observations that could aid us in assessing the situation or guaranteeing the safety of the individual, please do not hesitate to reach out to the appropriate authorities or notify us at %s. Your prompt action can make a significant difference in this situation.\n" +
		"Please remember that personal safety and the well-being of others are our top priorities. Let's work together to ensure the safety and security of everyone in our community.\n" +
		"Thank you for your immediate attention and cooperation.\n\n" +
		"Here is the link of the latest device location=%s \n\n" +
		"Best regards,"
	ALARM_SUBJ = "Urgent Alert - Potential Safety Concern for Device ID %s"
)
