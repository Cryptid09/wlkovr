# viasocket Automation Workflow Setup Guide (Sponsor Integration)

This guide documents the complete setup of the **viasocket** automation flow connecting citizen WhatsApp channels directly to the Public Development Intelligence Platform.

---

## 1. Architecture Overview

```
 [Citizen WhatsApp Message]
             │ (Meta WhatsApp Cloud API / Webhook)
             ▼
    ┌─────────────────┐
    │    viasocket    │
    │   Automation    │
    │    Workflow     │
    └────────┬────────┘
             │ (POST Normalized JSON)
             ▼
┌─────────────────────────┐
│ Golang Backend Engine   │
│ POST /api/v1/webhooks/  │
│      viasocket          │
└────────────┬────────────┘
             │ (WebSocket Broadcast <50ms)
             ▼
┌─────────────────────────┐
│ Next.js 15 Live         │
│ Policymaker Dashboard   │
└─────────────────────────┘
```

---

## 2. Setting Up the Workflow in viasocket Console

### Step 1: Create a New Flow in viasocket
1. Log in to [viasocket Console](https://viasocket.com).
2. Click **Create Flow** and name it: `Civic-Intelligence-WhatsApp-Ingestion`.
3. Select **Webhook Trigger** (or WhatsApp Cloud App Trigger).

### Step 2: Configure Webhook Trigger
1. Set the HTTP method to `POST`.
2. viasocket generates a unique webhook endpoint URL (e.g., `https://flow.sokt.io/func/xyz123`).
3. If using Meta WhatsApp Cloud API:
   - Configure this URL as the Callback URL in Meta App Dashboard.
   - Enter your Verify Token.

### Step 3: Add Transformation Step (viasocket Code Step)
Add a **JavaScript Code Block** in viasocket to normalize any incoming WhatsApp or SMS structure into our platform's canonical `ViasocketPayload`:

```javascript
// viasocket Transformation Code Step
function formatPayload(context) {
  const incoming = context.req.body;
  
  // Extract WhatsApp Message fields
  const messageObj = incoming.entry?.[0]?.changes?.[0]?.value?.messages?.[0] || incoming;
  const contactObj = incoming.entry?.[0]?.changes?.[0]?.value?.contacts?.[0] || {};

  const senderNumber = messageObj.from || incoming.sender || incoming.phone || "unknown";
  const senderName = contactObj.profile?.name || "Citizen";
  const textBody = messageObj.text?.body || incoming.body || incoming.text || "";

  return {
    event_id: `via-evt-${Date.now()}`,
    provider: "WhatsApp",
    sender: senderNumber,
    body: textBody,
    timestamp: new Date().toISOString(),
    metadata: {
      flow_id: "viasocket-civic-inbox",
      sender_name: senderName,
      raw_source: "whatsapp_cloud_api"
    }
  };
}
```

### Step 4: Add HTTP Request Action (Forward to Backend)
1. Select **HTTP Request / API Step**.
2. **Method**: `POST`
3. **URL**:
   - Local Dev with Tunnel: `https://<your-ngrok-or-localtunnel-subdomain>/api/v1/webhooks/viasocket`
   - Cloud Run: `https://civic-engine-<project-id>.a.run.app/api/v1/webhooks/viasocket`
4. **Headers**:
   ```json
   {
     "Content-Type": "application/json"
   }
   ```
5. **Body**: Pass the output JSON from Step 3:
   ```json
   {
     "event_id": "${context.steps.code.output.event_id}",
     "provider": "WhatsApp",
     "sender": "${context.steps.code.output.sender}",
     "body": "${context.steps.code.output.body}",
     "timestamp": "${context.steps.code.output.timestamp}",
     "metadata": {
       "sender_name": "${context.steps.code.output.metadata.sender_name}"
     }
   }
   ```
6. Click **Save & Publish Flow**.

---

## 3. Local Tunneling for Live Demo / Testing

To receive live webhooks on your local workstation during the demo presentation:

1. Launch Ngrok or Cloudflare Tunnel:
   ```bash
   ngrok http 8080
   ```
2. Copy the forwarding HTTPS address (e.g. `https://abc-123.ngrok-free.app`).
3. Update the HTTP URL in viasocket flow to:
   `https://abc-123.ngrok-free.app/api/v1/webhooks/viasocket`
4. Send a WhatsApp message to your connected WhatsApp bot.
5. Watch the signal appear on the Next.js dashboard within **$< 500\text{ ms}$**!

---

## 4. Webhook Payload Specification

### Request (`POST /api/v1/webhooks/viasocket`)
```json
{
  "event_id": "via-evt-wa-1001",
  "provider": "WhatsApp",
  "sender": "+91-9826012345",
  "body": "हमारे यहाँ नल से गंदा पानी आ रहा है बहुत बदबू है चंदन नगर गली 4",
  "timestamp": "2026-09-06T06:15:30Z",
  "metadata": {
    "sender_name": "Ramesh Sharma",
    "location_hint": "Chandan Nagar Street 4",
    "ward_id": "indore-ward-02"
  }
}
```

### Response (`200 OK`)
```json
{
  "success": true,
  "message": "Signal ingested and broadcasted successfully",
  "data": {
    "signal_id": "sig-a1b2c3d4",
    "timestamp": "2026-09-06T06:15:30.123456Z"
  }
}
```

---

## 5. Verification Checklist

- [x] viasocket receives incoming WhatsApp event.
- [x] JS code step normalizes body, sender, and metadata.
- [x] HTTP step delivers JSON to `/api/v1/webhooks/viasocket`.
- [x] Backend responds with HTTP 200 in $<20$ ms.
- [x] Backend triggers Go WebSocket broadcast (`SIGNAL_RECEIVED`).
- [x] Next.js dashboard UI updates live stream feed in $<2$ seconds.
