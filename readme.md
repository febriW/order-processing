<!-- Build and run -->
docker compose up --build -d

<!-- Example Scale Service -->
docker compose up -d --scale auth_service=3