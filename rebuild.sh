cd web
npm run build
cd ..
go build -ldflags "-s -w" -o one-api
./one-api --port 3001 --log-dir ./logs