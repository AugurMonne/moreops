from flask import Flask, jsonify
import os

app = Flask(__name__)

@app.route("/healthz")
def healthz():
    return jsonify({"status": "healthy"}), 200

@app.route("/metrics")
def metrics():
    return "# HELP http_requests_total The total number of HTTP requests.\n# TYPE http_requests_total counter\nhttp_requests_total 42\n", 200, {'Content-Type': 'text/plain; version=0.0.4'}

@app.route("/")
def index():
    return jsonify({
        "service": "python-api",
        "version": os.getenv("APP_VERSION", "1.0.0"),
        "status": "running"
    })

if __name__ == "__main__":
    app.run(host="0.0.0.0", port=8000)
