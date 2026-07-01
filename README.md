You can put the video input .mp4 (or change it from main.go 'sourceFileKey' ) in a "raw-videos" bucket and the hls worker will download it, transcode it into HLS (multiple resolutions, .m3u8 index and .ts segments)
and put the output in the "processed-video" bucket.

And the streaming of the video can be done by sharing the master.m3u8 file (present inside the processed-video/test (or can be change from main.go 'targetPrefix') bucket.)

```
docker build -t hls-worker .

```

docker build -t hls-worker .

docker run --rm \
 -e R2_ENDPOINT="https://YOUR_CLOUDFLARE_ACCOUNT_ID.r2.cloudflarestorage.com" \
 -e R2_ACCESS_KEY="YOUR_R2_ACCESS_KEY_ID" \
 -e R2_SECRET_KEY="YOUR_R2_SECRET_ACCESS_KEY" \
 -e R2_PUBLIC_URL="https://YOUR_R2_PUBLIC_CUSTOM_DOMAIN_OR_DEV_URL" \
 -e RAW_BUCKET="raw-videos" \
 -e PROCESSED_BUCKET="processed-video" \
hls-worker

```

```

Next step is creating an autonomous pipeline with AWS services. 😉
