// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package courier

import (
	"context"
	"encoding/json"

	"github.com/gofrs/uuid"
)

func (c *courier) QueueSMS(ctx context.Context, t SMSTemplate) (uuid.UUID, error) {
	recipient, err := t.PhoneNumber()
	if err != nil {
		return uuid.Nil, err
	}

	templateData, err := json.Marshal(t)
	if err != nil {
		return uuid.Nil, err
	}

	body, err := t.SMSBody(ctx)
	if err != nil {
		return uuid.Nil, err
	}

	message := &Message{
		Status:       MessageStatusQueued,
		Type:         MessageTypeSMS,
		Channel:      "sms",
		Recipient:    recipient,
		TemplateType: t.TemplateType(),
		TemplateData: templateData,
		Body:         body,
	}
	if err := c.deps.CourierPersister().AddMessage(ctx, message); err != nil {
		return uuid.Nil, err
	}

	return message.ID, nil
}
