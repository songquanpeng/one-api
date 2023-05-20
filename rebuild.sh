cd web
npm run build
cd ..
go build -ldflags "-s -w" -o one-api
./one-api --port 3000 --log-dir ./logs