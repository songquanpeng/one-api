set GIN_MODE=debug
set SQL_DSN=root:123456@tcp(localhost:3306)/oneapi
set REDIS_CONN_STRING=redis://:jifeng123Redis@www.jifeng.online:8867/3
set SYNC_FREQUENCY=1800
one-api.exe  --port 3000 --log-dir ./logs