#!/usr/bin/env python3

import requests
import subprocess
import os
import time
import threading


# Print all environment variables (for debugging purposes)
print("Current environment variables:")
for key, value in os.environ.items():
    print(f"{key}: {value}")


# Function to start the Go application
def start_go_app():
    # Get the current environment variables
    env = os.environ.copy()

    # Set the GIN_MODE environment variable
    env['GIN_MODE'] = 'release'  # or 'debug', depending on your needs

    # Start the Go application as a subprocess
    go_process = subprocess.Popen(['./gin-blog'], stdout=subprocess.PIPE, stderr=subprocess.PIPE, \
                                  text=True, \
                                  env=env  # Pass the modified environment
                                  )

    # Function to read output from the Go application
    def read_output(process):
        for line in process.stdout:
            print(f"[Go App] {line.strip()}")
        for line in process.stderr:
            print(f"[Go App Error] {line.strip()}")

    # Start a thread to read the output
    threading.Thread(target=read_output, args=(go_process,), daemon=True).start()

    # Wait for a moment to ensure the Go server starts
    time.sleep(2)
    return go_process

# Start the Go application
go_process = start_go_app()

def application(environ, start_response):
    # Extract the request method and path
    method = environ['REQUEST_METHOD']
    path = environ['PATH_INFO']
    headers = {key: environ[key] for key in environ if key.startswith('HTTP_')}

    # Prepare the request to the GoLang application
    go_url = f'http://localhost:3000{path}'

    try:
        # Forward the request to the GoLang application
        if method == 'GET':
            response = requests.get(go_url)
        elif method == 'POST':
            body = environ['wsgi.input'].read(int(environ.get('CONTENT_LENGTH', 0)))
            response = requests.post(go_url, data=body, headers=environ)
        else:
            start_response('405 Method Not Allowed', [('Content-Type', 'text/plain')])

            return [b'Method Not Allowed']

        # Prepare the response from the GoLang application
        status = f"{response.status_code} {response.reason}"
        headers = [(key, value) for key, value in response.headers.items()]
        start_response(status, headers)

        return [response.content]

    except Exception as e:
        start_response('500 Internal Server Error', [('Content-Type', 'text/plain')])

        return [f'Proxy error: {str(e)}'.encode()]


if __name__ == '__main__':
    from wsgiref.simple_server import make_server

    PORT = 5000
    httpd = make_server('', PORT, application)
    print(f"WSGI Proxy Server is running on http://localhost:{PORT}")

    try:
        httpd.serve_forever()
    except KeyboardInterrupt:
        print("Shutting down the server.")
        go_process.terminate()  # Terminate the Go process on exit
