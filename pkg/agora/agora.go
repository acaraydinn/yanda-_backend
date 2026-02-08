package agora

import (
	"fmt"

	rtctokenbuilder "github.com/AgoraIO/Tools/DynamicKey/AgoraDynamicKey/go/src/rtctokenbuilder2"
)

// GenerateRTCToken generates an Agora RTC token for joining a channel
func GenerateRTCToken(appID, appCertificate, channelName string, uid uint32, expireSeconds uint32) (string, error) {
	if appID == "" || appCertificate == "" {
		return "", fmt.Errorf("agora app ID and certificate are required")
	}

	token, err := rtctokenbuilder.BuildTokenWithUid(
		appID,
		appCertificate,
		channelName,
		uid,
		rtctokenbuilder.RolePublisher,
		expireSeconds,
		expireSeconds,
	)
	if err != nil {
		return "", fmt.Errorf("failed to build Agora token: %w", err)
	}

	return token, nil
}
