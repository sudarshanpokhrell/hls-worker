```
docker build -t hls-worker .

docker run --rm -v "$(pwd)":/data hls-worker
```

After output is generated

```
bun x serve .
```
