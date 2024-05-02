package common

import log "github.com/sirupsen/logrus"

// Prints an error log message where the state is not found for some node.
func LogStateNotFoundError(handler string, function string, found bool) {
	log.WithFields(
		log.Fields{
			"Found":   found,
			"Message": "State not found",
		},
	).Error(handler + ": " + function)

}

// Prints an error log message where the state retrieval has errors.
func LogStateRetrieveError(handler string, function string, err error) {
	log.WithFields(
		log.Fields{
			"Error":   err,
			"Message": "Error retrieving the state of the node.",
		},
	).Error(handler + ": " + function)
}

// Prints an error log message when marshaling a message into bytes returns an
// error.
func LogMarshalError(handler string, function string, err error) {
	log.WithFields(log.Fields{
		"Error":   err,
		"Message": "Error while converting the message into bytes",
	}).Error(handler + ": " + function)
}

// Prints an error log message when the creation of a message fails.
func LogErrorNewMessage(handler string, function string, messageType string, err error) {
	log.WithFields(
		log.Fields{
			"MessageType": messageType,
			"Error":       err,
			"Message":     "Error creating the message of the given type",
		},
	).Error(handler + ": " + function)
}

// Prints a error log message when updating the state of a node fails.
func LogStateUpdateError(handler string, function string, stateType string, err error) {
	log.WithFields(
		log.Fields{
			"Message": "error updating the state",
			"State":   stateType,
			"Error":   err,
		},
	).Error(handler + ": " + function)
}
