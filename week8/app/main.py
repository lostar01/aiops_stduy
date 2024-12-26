import time
from datetime import datetime, timezone
from flask import Flask, render_template_string
import threading
from kubernetes import client, config
import os


error_messages = [
    "[ERROR] Database connection failed: Unable to connect to database at 'db.payment.local'.",
    "[ERROR] Service 500 Error: Downstream service 'order-processing' returned status code 500.",
    "[ERROR] Memory OOM: Container 'payment-service' exceeded memory limit.",
    "[WARING] Alipay API call failed: Response timeout after 10 seconds.",
    "[WARING] Payment gateway timeout: No response from gateway after 30 seconds.",
    "[WARING] Invalid payment method: Credit card number is invalid or expired.",
    "[WARING] Insufficient funds: The account balance is lower than the requested payment amount.",
    "[ERROR] Fraud detection failed: Payment flagged as potentially fraudulent.",
    "[WARING] Configuration error: Missing API key for payment provider.",
    "[WARING] Unexpected token in JSON: Malformed response from payment processor.",
]

app = Flask(__name__)


@app.route("/")
def home():
    try:
        config.load_incluster_config()
        with open("/var/run/secrets/kubernetes.io/serviceaccount/namespace") as f:
            current_namespace = f.read()
    except Exception as e:
        current_namespace = "Unknown"
    hostname = os.getenv("HOSTNAME", "Unknown")
    html_content = (
        """
    <html>
        <head>
            <style>
                body {
                    background-color: blue;
                    display: flex;
                    justify-content: center;
                    align-items: center;
                    height: 100vh;
                    margin: 0;
                    color: white;
                    font-size: 2em;
                }
            </style>
        </head>
        <body>
            <div>ChatOps example app
            <br/><br/>namespace: """
        + current_namespace
        + """<br/><br/>host: """
        + hostname
        + """</div>
        </body>
    </html>
    """
    )
    return render_template_string(html_content)


def simulate_payment_service_errors():
    while True:
        for error_message in error_messages:
            current_time = datetime.now(timezone.utc)
            # formatted_time = (
            #     current_time.strftime("%Y-%m-%dT%H:%M:%S.")
            #     + f"{current_time.microsecond:06d}000Z"
            # )
            formatted_time = (
                current_time.strftime("%Y-%m-%d %H:%M:%S.")
                + f"{current_time.microsecond // 1000:03d}"
            )
            print(f"{formatted_time} {error_message}")
            time.sleep(2)


if __name__ == "__main__":
    error_thread = threading.Thread(target=simulate_payment_service_errors)
    error_thread.daemon = True
    error_thread.start()

    app.run(host="0.0.0.0", port=8080)
