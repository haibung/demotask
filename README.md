# Demo Task Backend (Docker Quickstart)

This project uses **Golang** + **PostgreSQL** + **Redis**.

## 1. Initial Setup

- Ensure the backend source code (including `go.mod`, `go.sum`, `main.go`) is in this directory.
- Copy the configuration file:

```bash
cp config.yml.example config.yml
```

- Open `config.yml` and provide your sandbox credentials from the [PayPal Developer Dashboard](https://developer.paypal.com/):
  - `clientId`: Your PayPal Sandbox Client ID.
  - `clientSecret`: Your PayPal Sandbox Client Secret.
  - `webhookId`: Your Webhook ID (See Section 4 on how to generate this).

## 2. Run with Docker Compose

Start the database, redis, and the backend service:

```bash
docker compose up -d --build
```

Access the services at:
- API: `http://localhost:8090`
- Demo static UI: `http://localhost:8090/static/index.html`

## 3. Database Migrations

Run database migrations inside the running backend container:

```bash
docker compose exec backend /app/app migrate up
```

*(Note: If you need to revert or create new migrations, you can use `migrate down` or `migrate create table_name`).*

## 4. PayPal Webhook & Ngrok Setup

To receive PayPal webhooks locally, you must run the Ngrok tunnel.

1. Export your Ngrok authentication token (get this from your [Ngrok Dashboard](https://dashboard.ngrok.com/)):
```bash
export NGROK_AUTHTOKEN="your_ngrok_auth_token_here"
```

2. Start the tunnel via the compose profile:
```bash
docker compose --profile tunnel up -d ngrok
```

3. Open the Ngrok Inspector UI at `http://localhost:4040` and copy your public HTTPS URL (e.g., `https://<id>.ngrok-free.app`).
4. Go to your **PayPal Developer Dashboard** -> **App & Credentials** -> **Webhooks** -> **Add Webhook**.
5. Paste your Ngrok URL followed by your webhook endpoint (e.g., `https://<id>.ngrok-free.app/v1/paypal/webhook`). Select all events or specific payment/subscription events, and click Save.
6. PayPal will provide a **Webhook ID**. Copy this ID into your `config.yml` under `paypal.webhookId`.
7. Restart the backend to apply the new configuration:
```bash
docker compose restart backend
```

## 5. Typical API Workflow

Use the provided Postman collection or `curl` to test the endpoints in this order:

- Create Product on PayPal: `POST http://localhost:8090/v1/products`
- Create a One-Time Price for Product: `POST http://localhost:8090/v1/products/one-time-price`
- Create a Billing Plan (Subscription): `POST http://localhost:8090/v1/billing-plans`

> Note: Docker compose exposes the application on **8090** globally. The configuration is primarily managed through `config.yml`.

---

## Manual Installation (Without Docker)

This project uses **Golang** and **PostgreSQL**. If you prefer to run the application locally without Docker, follow these steps:

### 1. Installation & Configuration

1. Copy the configuration file:
   ```bash
   cp config.yml.example config.yml
   ```
2. Open `config.yml` and set up your PostgreSQL database configuration (`host` should be `localhost` or your local DB IP).
3. Ensure Redis is running locally and update `config.yml` if necessary.
4. Configure an NGROK tunnel or port forwarding to expose your local port (default `8090`) to receive PayPal webhooks. Update the `webhookId` in `config.yml` accordingly.

### 2. Database Migrations

Run migrations using the built-in command (Goose is included):

```bash
go run main.go migrate up
```

Other available migration commands:
```bash
go run main.go migrate down
go run main.go migrate create table_name
```

### 3. Start the Service

Start the backend service (default port is `8090` based on `config.yml`):

```bash
go run main.go start
```

### 4. Setup Data & Workflow

1. Import the provided **DEMO TASK Postman collection**.
2. Insert manual product data into the database via Postman (if required by your flow).
3. Execute the API workflow in order:
   - Create Product on PayPal via Postman: `POST http://localhost:8090/v1/products`
   - Create One-Time Price Product `{Token}` via Postman: `POST http://localhost:8090/v1/products/one-time-price`
   - Create Billing Plan `{User Token}` via Postman: `POST http://localhost:8090/v1/billing-plans`

### 5. Access the Demo Application

Once running, access the live demo interface at:
`http://localhost:8090/static/index.html`
