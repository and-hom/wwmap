```
migrate -url postgres://wwmap@localhost:5432/wwmap -path ./db/migrations
```

Revert latest
```
migrate -url postgres://wwmap:wwmap@localhost:5432/wwmap -path ./ migrate -1
```
