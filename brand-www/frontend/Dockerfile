# Use a minimal Python image
FROM python:3.12-slim

# Set the working directory inside the container
WORKDIR /app

# Copy all the project files into the container
COPY . .

# Expose port 8080
EXPOSE 8080

# Command to start the HTTP server
CMD ["python3", "-m", "http.server", "8080"]
