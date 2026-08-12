# SiftQ provider

One API exposes SiftQ as an independent asynchronous video provider through a provider-neutral, OpenAI-style Videos API. Clients never call SiftQ's native routes directly.

## Channel configuration

- Type: `SiftQ`
- API key: a SiftQ bearer key stored in the channel secret field
- Default Base URL: `https://siftq.com/api/minimax/`
- Model: `MiniMax-H3`

The Base URL can be overridden per channel. Video IDs are bound to the One API user and SiftQ channel that created them, so follow-up requests return to the same upstream account and cannot access another user's tasks.

## Public API

All routes use a normal One API token as `Authorization: Bearer <one-api-token>`.

| Operation | One API route |
| --- | --- |
| Create video | `POST /v1/videos` |
| Retrieve video | `GET /v1/videos/{video_id}` |
| List videos | `GET /v1/videos` |
| Download completed content | `GET /v1/videos/{video_id}/content` |
| Delete or cancel | `DELETE /v1/videos/{video_id}` |

The native `/v2/*` SiftQ paths are not public One API routes. The gateway selects the configured `MiniMax-H3` SiftQ channel and translates the operations internally:

| One API operation | SiftQ upstream operation |
| --- | --- |
| Create | `POST v2/video_generation` |
| Retrieve / download metadata | `GET v2/query/video_generation/{video_id}` |
| Delete or cancel | `DELETE v2/video_generation/{video_id}` |

## Create a video

JSON example:

```bash
curl http://127.0.0.1:3000/v1/videos \
  -H "Authorization: Bearer $ONE_API_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "MiniMax-H3",
    "prompt": "A cinematic ocean sunrise.",
    "seconds": "5",
    "size": "1280x720"
  }'
```

`multipart/form-data` is also supported and follows the OpenAI Videos API style:

```bash
curl http://127.0.0.1:3000/v1/videos \
  -H "Authorization: Bearer $ONE_API_TOKEN" \
  -F "model=MiniMax-H3" \
  -F "prompt=The camera slowly pushes in." \
  -F "seconds=5" \
  -F "size=1280x720" \
  -F "input_reference=@first-frame.png"
```

JSON callers can use `"input_reference":{"image_url":"https://example.com/first-frame.png"}`. SiftQ does not accept OpenAI file IDs, so `input_reference.file_id` is rejected explicitly.

The supported sizes map to SiftQ as follows:

| Public `size` | SiftQ `resolution` | SiftQ `ratio` |
| --- | --- | --- |
| `720x1280` | `768P` | `9:16` |
| `1280x720` | `768P` | `16:9` |
| `1024x1792` | `2K` | `9:16` |
| `1792x1024` | `2K` | `16:9` |

An input image is sent as SiftQ's `first_frame` and uses `adaptive` ratio. `seconds` accepts each integer from 4 through 15, matching the provider's range. Defaults are `MiniMax-H3`, 4 seconds, and `720x1280`.

Create and retrieve return an OpenAI-style video object. SiftQ states are normalized as `queued`, `in_progress`, `completed`, or `failed`. Use the returned `id` for polling and download:

```bash
curl http://127.0.0.1:3000/v1/videos/video_id_here \
  -H "Authorization: Bearer $ONE_API_TOKEN"

curl http://127.0.0.1:3000/v1/videos/video_id_here/content \
  -H "Authorization: Bearer $ONE_API_TOKEN" \
  --output result.mp4
```

List accepts OpenAI-style `limit`, `order`, and `after` query parameters. Delete maps to SiftQ's state-aware cancel/delete operation and returns a `video.deleted` object.

The normalized `MiniMax-H3` model ratio is charged once per requested second and remains configurable through One API's model-ratio setting. Failed submissions are refunded; query, list, download, and delete do not consume quota. Provider failures are converted to the standard One API/OpenAI error envelope without logging credentials or raw malformed bodies.
