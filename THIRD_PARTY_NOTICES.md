# Third Party Notices

TypeType Downloader Go depends on Go modules listed in `go.mod` and `go.sum`.

The Docker runtime image also links against FFmpeg libraries from Wolfi:

| Package | License metadata |
|---|---|
| `ffmpeg-8.1-libavformat62` | `GPL-2.0-or-later AND LGPL-2.1-or-later` |
| `ffmpeg-8.1-libavcodec62` | `GPL-2.0-or-later AND LGPL-2.1-or-later` |
| `ffmpeg-8.1-libavutil60` | `GPL-2.0-or-later AND LGPL-2.1-or-later` |

The Wolfi `libavcodec` package depends on GPL-related codec libraries including
`libx264` and `libx265`. Because of that, the distributed Docker image should be
treated as GPL-compatible, not Apache-only.

If the project later switches to an LGPL-only FFmpeg build, the project license
can be reconsidered.
