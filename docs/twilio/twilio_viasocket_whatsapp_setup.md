# Twilio WhatsApp + Viasocket Integration Setup

## 1. Project Overview

This document records the current setup for the WhatsApp AI chatbot demo
using:

-   **Twilio WhatsApp Sandbox**
-   **Viasocket Flow / Webhook**
-   **Gemini** (to be connected as the AI response step)
-   **WhatsApp** as the user interface

### Current working flow

``` text
WhatsApp User
      |
      v
Twilio WhatsApp Sandbox
      |
      | HTTP POST
      v
Viasocket Webhook
      |
      v
Gemini (next step)
      |
      v
Twilio WhatsApp reply
      |
      v
WhatsApp User
```

------------------------------------------------------------------------

# 2. Twilio Configuration

## Twilio Account

**Account SID**

``` text
ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

> The real Account SID lives in `.env` as `TWILIO_ACCOUNT_SID` and is not
> committed. It is not a credential on its own — the Auth Token is the
> secret — but it identifies the account, GitHub push protection blocks
> it, and this repository is public. Keep both out of Git.

## WhatsApp Sandbox Sender

The incoming webhook payload confirms that the active Sandbox sender is:

``` text
whatsapp:+14155238886
```

Plain number:

``` text
+14155238886
```

This is important because an earlier test used:

``` text
whatsapp:+17372508034
```

but the actual incoming request received by Viasocket contains:

``` json
"To": "whatsapp:+14155238886"
```

Therefore, for the current Sandbox integration, use:

``` text
From = whatsapp:+14155238886
```

Twilio documentation also specifies `whatsapp:+14155238886` as the
standard Sandbox sender format.

------------------------------------------------------------------------

# 3. WhatsApp Test User

The WhatsApp number currently sending messages to the Sandbox is:

``` text
whatsapp:+916232230297
```

Plain number:

``` text
+916232230297
```

The WhatsApp profile name observed in the webhook request is:

``` text
Nidhi Agrawal
```

------------------------------------------------------------------------

# 4. Viasocket Webhook

## Webhook URL

``` text
https://flow.sokt.io/func/scri1VhNpy2q
```

This URL is configured in Twilio under:

``` text
Sandbox Settings
    -> Sandbox Configuration
        -> When a message comes in
```

## HTTP Method

``` text
POST
```

## Status Callback

Currently not configured.

``` text
Status callback URL = empty
```

This is not required for receiving the incoming WhatsApp message.

------------------------------------------------------------------------

# 5. Twilio Webhook Configuration

Current configuration:

  Setting                   Value
  ------------------------- ------------------------------------------
  When a message comes in   `https://flow.sokt.io/func/scri1VhNpy2q`
  Method                    `POST`
  Status callback URL       Not configured
  Status callback method    `POST`

After changing the webhook configuration, click **Save**.

------------------------------------------------------------------------

# 6. Viasocket Flow

The current Viasocket flow contains:

``` text
Trigger
  |
  +-- Webhook
       https://flow.sokt.io/func/scri1VhNpy2q

Action
  |
  +-- Currently empty

Response
  |
  +-- Default response
```

The Viasocket webhook is successfully receiving Twilio requests.

The Viasocket run history shows successful executions.

------------------------------------------------------------------------

# 7. Verified Incoming Webhook Data

When the WhatsApp user sends:

``` text
Hi
```

Viasocket receives the following important values:

``` json
{
  "Body": "Hi",
  "From": "whatsapp:+916232230297",
  "To": "whatsapp:+14155238886",
  "SmsStatus": "received",
  "MessageType": "text",
  "ProfileName": "Nidhi Agrawal",
  "WaId": "916232230297",
  "NumMedia": "0",
  "AccountSid": "ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
  "ApiVersion": "2010-04-01"
}
```

The complete request also contains Twilio headers such as:

``` text
content-type: application/x-www-form-urlencoded
x-twilio-signature: <dynamic value>
```

The Twilio signature changes per request and should not be hard-coded.

------------------------------------------------------------------------

# 8. Important Finding

The webhook integration is **working**.

We have verified:

``` text
WhatsApp
   |
   v
Twilio
   |
   v
Viasocket Webhook
```

because Viasocket receives:

``` text
Body = Hi
From = whatsapp:+916232230297
To = whatsapp:+14155238886
SmsStatus = received
```

Therefore, the webhook URL itself does not need to be changed.

------------------------------------------------------------------------

# 9. Why Viasocket Shows an Empty Response

The Viasocket flow currently has no action after the webhook.

Current flow:

``` text
Webhook
   |
   v
No action
   |
   v
Default Response
   |
   v
""
```

Therefore, the empty response is expected.

It does **not** mean that the webhook failed.

The next step is to add an action that processes the incoming message.

------------------------------------------------------------------------

# 10. Target Flow — routed through the platform

The earlier plan (Webhook -> Gemini -> Twilio reply) would produce a generic
chatbot that answers WhatsApp messages. It would bypass the platform entirely:
no ward extraction, no clustering, no hotspot or blind-spot detection, and
nothing on the policymaker dashboard.

The integrated flow sends the citizen's message into the platform instead:

``` text
WhatsApp user
      |
      v
Twilio WhatsApp Sandbox
      |
      | POST (form-encoded)
      v
Viasocket flow            <- sponsor integration layer
      |
      | POST + X-Webhook-Secret header
      v
Go backend  /api/v1/webhooks/viasocket
      |
      +--> Gemini extraction (issue, ward, department, urgency, hazards)
      +--> gemini-embedding-001 vector
      +--> matched into an existing issue cluster and rescored
      +--> persisted to Firestore
      +--> broadcast live to the dashboard over WebSocket
      |
      v
Twilio reply to the citizen
"Understood as ... Location ... Routed to ...
 matches N other reports ... nothing approved automatically"
```

The backend accepts **both** payload shapes, so the viasocket action can either
transform the Twilio fields into JSON or forward them untouched:

* Twilio form-encoded: `Body`, `From`, `To`, `ProfileName`, `WaId`, `MessageSid`
* viasocket JSON: `{"provider": "...", "sender": "...", "body": "..."}`

`From: whatsapp:+91...` is detected as WhatsApp; a bare `+91...` is treated as
SMS. The `whatsapp:` prefix is stripped from the stored phone number, and
`ProfileName`, `WaId` and `MessageSid` are kept as signal metadata.

## What to configure in the viasocket flow

The flow currently has a webhook trigger and **no action**, which is why it
returns an empty response. Add one HTTP action:

``` text
Method  POST
URL     https://<public-backend-url>/api/v1/webhooks/viasocket
Header  X-Webhook-Secret: <value of VIASOCKET_WEBHOOK_SECRET>
Body    forward the incoming Twilio payload unchanged
```

The secret header is required — the endpoint returns HTTP 401 without it,
because a publicly reachable ingestion URL would otherwise let anyone inject
fabricated citizen reports.

To expose a local backend:

``` bash
cloudflared tunnel --url http://localhost:8080   # no account required
```

Note that a cloudflared quick tunnel gets a **new URL every restart**, so the
viasocket action must be updated whenever the tunnel is restarted.

## Replying to the citizen — the 24-hour session window

WhatsApp only allows free-form business replies inside a **24-hour customer
service window**, which opens when the citizen sends a message. Outside it,
Twilio rejects the reply with:

``` text
21654 ContentSid Required
```

That is expected when replaying a synthetic webhook, because no real inbound
WhatsApp message opened a window. During the live demo the citizen's own
message opens the window, so the acknowledgement sends normally.

The backend detects this specific failure and logs what to do rather than a raw
API error. Ingestion is unaffected either way — the signal is still extracted,
clustered and shown on the dashboard.

To reply outside the window an approved WhatsApp template is required
(`ContentSid`), which is not configured for this sandbox.

## Configuring replies

The backend sends the acknowledgement itself once extraction completes; the
viasocket flow does not need a Twilio step. Replies require:

``` env
TWILIO_ACCOUNT_SID=AC...
TWILIO_AUTH_TOKEN=<secret>
TWILIO_WHATSAPP_FROM=whatsapp:+14155238886
```

Without `TWILIO_AUTH_TOKEN` the platform ingests normally and simply does not
reply. The acknowledgement states what was understood and that nothing has been
approved automatically — it is a receipt, never a promise of work.

# 11. Important WhatsApp Sandbox Details

The Twilio WhatsApp Sandbox is intended for testing and development.

For the current Sandbox setup:

-   The user must join the Sandbox before receiving messages.
-   The Sandbox sender is `whatsapp:+14155238886`.
-   User-initiated messages open a customer-service window.
-   During the 24-hour customer-service window, free-form replies can be
    sent.
-   Outside that window, approved templates may be required for
    business-initiated messages.

------------------------------------------------------------------------

# 12. Credentials and Secrets

## Safe to document for team configuration

``` text
Twilio Account SID:
ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx

Twilio WhatsApp Sandbox:
whatsapp:+14155238886

Test WhatsApp user:
whatsapp:+916232230297

Viasocket webhook:
https://flow.sokt.io/func/scri1VhNpy2q
```

## DO NOT share publicly

Do NOT commit or paste the following into GitHub, public documents,
screenshots, or chat:

``` text
TWILIO_AUTH_TOKEN
GEMINI_API_KEY
Other API keys
Other access tokens
Private credentials
```

For local development, store secrets in environment variables or a
`.env` file that is excluded from Git.

Example:

``` env
TWILIO_ACCOUNT_SID=ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
TWILIO_AUTH_TOKEN=your_secret_token
GEMINI_API_KEY=your_secret_key
```

Add this to `.gitignore`:

``` gitignore
.env
```

------------------------------------------------------------------------

# 13. Current Status

  Component                           Status
  ----------------------------------- ------------------------
  Twilio account                      Working
  WhatsApp Sandbox                    Working
  WhatsApp user joined                Working
  Twilio → Viasocket webhook          **Verified**
  Viasocket receives `Body`           **Verified**
  Viasocket receives `From`           **Verified**
  Viasocket receives `To`             **Verified**
  Gemini integration                  **Done** (in the Go backend)
  AI response generation              **Done** (extraction + clustering)
  Twilio → WhatsApp automated reply   **Implemented**, needs TWILIO_AUTH_TOKEN
  Status callback                     Not required currently

------------------------------------------------------------------------

# 14. Next Steps

1. Start the backend and expose it (`cloudflared tunnel --url http://localhost:8080`).
2. Add the HTTP action to the viasocket flow, pointing at
   `https://<tunnel>/api/v1/webhooks/viasocket` with the `X-Webhook-Secret` header.
3. Send a WhatsApp message to the sandbox describing a civic issue in an Indore
   ward, for example: *"Chandan Nagar gali 4 me nal ka pani ganda aa raha hai"*.
4. Watch the dashboard: the signal appears immediately, then the cluster it
   joined rescores a few seconds later once Gemini has understood it.
5. Set `TWILIO_AUTH_TOKEN` if the citizen should receive the acknowledgement.

**Verified:** a Twilio-shaped form payload posted through the public tunnel was
parsed, extracted by Gemini, matched into `cluster-indore-001` (8 → 9 signals)
and persisted, with `ProfileName` and `WaId` retained as metadata.

# 15. Troubleshooting Notes

### If Viasocket has no run history

Check:

1.  Twilio Sandbox webhook URL.
2.  HTTP method is `POST`.
3.  Twilio configuration was saved.
4.  User has joined the correct Sandbox.
5.  The message was sent to the active Sandbox number.

### If Viasocket receives data but response is empty

Check the Viasocket flow actions.

An empty default response means no response body/action has been
configured.

### If Twilio reports "no approved WhatsApp sender"

Check the `From` value.

For the current Sandbox use:

``` text
whatsapp:+14155238886
```

Do not use:

``` text
whatsapp:+17372508034
```

unless Twilio explicitly shows that number as the active sender for the
account/environment being used.

------------------------------------------------------------------------

# 16. Useful References

-   Twilio WhatsApp Sandbox documentation:
    https://www.twilio.com/docs/whatsapp/sandbox

-   Twilio WhatsApp Quickstart:
    https://www.twilio.com/docs/whatsapp/quickstart

-   Twilio error 63007: https://www.twilio.com/docs/api/errors/63007

------------------------------------------------------------------------

## Final Verified Configuration

``` text
TWILIO_ACCOUNT_SID
    ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx

TWILIO_WHATSAPP_FROM
    whatsapp:+14155238886

TEST_WHATSAPP_USER
    whatsapp:+916232230297

VIASOCKET_WEBHOOK
    https://flow.sokt.io/func/scri1VhNpy2q

WEBHOOK_METHOD
    POST
```

**Current milestone:** Twilio → Viasocket incoming WhatsApp webhook is
successfully working.

**Next milestone:** Viasocket → Gemini → Twilio → WhatsApp automated AI
reply.
