#!/bin/bash

## 1.Create the remote dump
# docker context use halendar
# rm -f ./remote/backup/db.dump
# docker exec postgres rm ./db.dump
# docker exec postgres pg_dump -U halendar -d halendar -F c -b -v -f ./db.dump
# docker cp postgres:/db.dump ./remote/backup/db.dump


## 2.Upload the new dump to local
docker context use default
docker exec postgres rm -f ./db.dump
docker cp ./remote/backup/db.dump postgres:/db.dump
docker exec -i postgres psql -U halendar -d postgres -c "SELECT pg_terminate_backend(pg_stat_activity.pid) FROM pg_stat_activity WHERE datname = 'halendar' AND pid <> pg_backend_pid();"
docker exec -i postgres psql -U halendar -d postgres -c "DROP DATABASE halendar;"
docker exec -i postgres psql -U halendar -d postgres -c "CREATE DATABASE halendar;"
docker exec -i postgres pg_restore -U halendar -d halendar ./db.dump