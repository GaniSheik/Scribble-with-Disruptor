<h1 align="center">Scribble.rs</h1>

<p align="center">
  <a href="https://discord.gg/cE5BKP2UnE"><img src="https://dcbadge.limes.pink/api/server/https://discord.gg/cE5BKP2UnE"></a>
  <a href="https://ko-fi.com/N4N07DNY"><img src="https://ko-fi.com/img/githubbutton_sm.svg"></a>
</p>

![demo](.github/demo.png)

Scribble.rs is a free and privacy respecting pictionary game. There's no
advertisements and you don't need an account to play.

It is an alternative to the web-based drawing game skribbl.io.

## Play now

There are some community hosted versions of the game (feel free to host your own instance and add it here!):
 - [scribblers.bios-marcel.link](https://scribblers.bios-marcel.link) (Official instance, Note
that the instance may not respond instantly, as it automatically shuts down
if no traffic is received.)
 - [scribble.bixilon.de](https://scribble.bixilon.de) (community instance maintained by @Bixilon)
 - [scribble.drifty.win](https://scribble.drifty.win) (community instance maintained by @driftywinds for better latency in Asia)

## Join The Discord

Feel free to join the community Discord server to find people to play, talk, get
help or whatever else: https://discord.gg/cE5BKP2UnE

Note, the server is NOT very active.

## Donations

I haven't really accepted donations for a long time. But I think the project is
polished enough now to dare taking some money. Right now the hosting is very
minimal and there's no domain. The server is located in amsterdam and many
people playing seem to be from outside of europe. So it'd be nice to have a
decentralised deployment to provide a nice experience for everyone.

So donations would go towards infrastructure and a domain!

In the future I might add some fun little benefits for donators, I have no clear
vision of it yet though.

You can donate via Ko-Fi: https://ko-fi.com/biosmarcel

## Configuration

Configuration is read from environment variables or a `.env` file located in
the working directory.

Available settings:

| Key                                       | Description                                                      | Default | Required |
| ----------------------------------------- | ---------------------------------------------------------------- | ------- | -------- |
| PORT                                      | HTTP port that the server listens to.                            | 8080    | True     |
| NETWORK_ADDRESS                           | TCP address that the server listens to.                          |         | False    |
| ROOT_PATH                                 | Changes the path (after your domain) that the server listens to. |         | False    |
| CORS_ALLOWED_ORIGINS                      |                                                                  | *       | False    |
| CORS_ALLOW_CREDENTIALS                    |                                                                  |         | False    |
| LOBBY_CLEANUP_INTERVAL                    |                                                                  | 90s     | False    |
| LOBBY_CLEANUP_PLAYER_INACTIVITY_THRESHOLD |                                                                  | 75s     | False    |

For more up-to-date configuration, read the
[config.go](/internal/config/config.go) file.

## Docker

It is recommended that you run the server via Docker, as this will rule out
almost all compatibility issues.

Starting from v0.8.5, docker images are only built on tagged pushes. Each git
tag becomes a docker tag, however `latest` will always point to the latest
version released via GitHub.

### Linux Docker

Download the image:

```shell
docker pull biosmarcel/scribble.rs:latest
```

### Windows Docker

Only use this one if you want to run a native Windows container. Otherwise use
the Linux variant, as that's the default mode on Windows:

```shell
docker pull biosmarcel/scribble.rs:windows-latest
```

### Running the Docker container

Run the following, replacing `<port>` with the port you want the container to be
reachable from outside:

```shell
docker run --pull always --env PORT=8080 -p <port>:8080 biosmarcel/scribble.rs:latest
```

For example:

```shell
docker run --pull always --env PORT=8080 -p 80:8080 biosmarcel/scribble.rs:latest
```

Note that you can change `8080` too, but it is the internal port of the
container and you shouldn't have to change it under normal circumstances.

## Building / Running

Dependencies:
  * [go](https://go.dev/doc/install) version 1.25.0 or later
  * [git](https://git-scm.com/) (You can also download a .zip from Github)

In order to download and build, open a terminal and execute:

```shell
git clone https://github.com/scribble-rs/scribble.rs.git
cd scribble.rs
go build ./cmd/scribblers
```

This will produce a portable binary called `scribblers` or `scribblers.exe` if
you are on Windows.

## Pre-compiled binaries

In the [Releases](https://github.com/scribble-rs/scribble.rs/releases) section
you can find the latest stable release.

Alternatively each commit uploads artifacts which will be available for a
certain time.

**Note that these binaries might not necessarily be compatible with your
system. In this case, please use Docker or compile them yourself.**

## nginx 

Since Scribble.rs uses WebSockets, when running it behind an nginx reverse
proxy, you have to configure nginx to support that. You will find an example
configuration on the [related Wiki page](https://github.com/scribble-rs/scribble.rs/wiki/reverse-proxy-(nginx)).

Other reverse proxies may require similar configuration. If you are using a
well known reverse proxy, you are free to contribute a configuration to the
wiki.

## Server-Side Metrics

While there's a Prometheus metrics endpoint at `/v1/metrics`, it currently
doesn't expose a lot of information. If there are any requests for certain data,
I'd be willing to extend it, as long as it doesn't expose any personal data.

While I do have a dashboard for it, my hoster (fly.io) sadly doesn't support
public dashboards right now, so the data will remain closed for now.

## Contributing

There are many ways you can contribute:

* Update / Add documentation in the wiki of the GitHub repository
* Create feature requests and bug reports
* Solve issues by creating Pull Requests
  * If  you are changing / adding behaviour, make an issue beforehand
* Tell your friends about the project

For contributions guidelines, see: https://github.com/scribble-rs/scribble.rs/wiki/Contributing

### Translations

For translations, please always use the `en_us.go` as your base. Other
translations might not be up-to-date. This will cause you to have missing keys
or translate obsolete keys.

## Credits

These resources are by people unrelated to the project, whilst not every of
these resources requires attribution as per license, we'll do it either way ;)

If you happen to find a mistake here, please make a PR. If you are one of the
authors and feel like we've wronged you, please reach out.

Some of these were slightly altered if the license allowed it.
Treat each of the files in this repository with the same license terms as the
original file.

* Logo - All rights reserved, excluded from BSD-3 licensing
* Background - All rights reserved, excluded from BSD-3 licensing
* Favicon - All rights reserved, excluded from BSD-3 licensing
* Rubber Icon - Made by [Pixel Buddha](https://www.flaticon.com/authors/pixel-buddha) from [flaticon.com](https://flaticon.com)
* Fill Bucket Icon - Made by [inipagistudio](https://www.flaticon.com/authors/inipagistudio) from [flaticon.com](https://flaticon.com)
* Kicking Icon - [Kicking Icon #309402](https://icon-library.net/icon/kicking-icon-4.html)
* Sound / No sound Icon - Made by Viktor Erikson (If this is you or you know who this is, send me a link to that persons Homepage)
* Profile Icon - Made by [kumakamu](https://www.iconfinder.com/kumakamu)
* [Help Icon](https://www.iconfinder.com/icons/211675/help_icon) - Made by Ionicons
* [Fullscreen Icon](https://www.iconfinder.com/icons/298714/screen_full_icon) - Made by Github
* [Pencil Icon](https://github.com/twitter/twemoji/blob/8e58ae4/svg/270f.svg)
* [Drawing Tablet Pen Icon](https://www.iconfinder.com/icons/8665767/pen_icon)
* [Checkmark Icon](https://commons.wikimedia.org/wiki/File:Green_check_icon_with_gradient.svg)
* [Fill Icon](https://commons.wikimedia.org/wiki/File:Circle-icons-paintcan.svg)
* [Trash Icon](https://www.iconfinder.com/icons/315225/trash_can_icon) - Made by [Yannick Lung](https://yannicklung.com)
* [Undo Icon](https://www.iconfinder.com/icons/308948/arrow_undo_icon) - Made by [Ivan Boyko](https://www.iconfinder.com/visualpharm)
* [Alarmclock Icon](https://www.iconfinder.com/icons/4280508/alarm_outlined_alert_clock_icon) - Made by [Kit of Parts](https://www.iconfinder.com/kitofparts)
* https://www.iconfinder.com/icons/808399/load_turn_turnaround_icon TODO


## Disruptor mode (custom fork)

This fork adds a hidden-role twist to the existing multiplayer game:

- Each turn, one connected, participating non-drawer is randomly picked as the **Disruptor**. At least **3 players** are needed, so that someone is still left to guess.
- The drawer still selects the secret word and draws normally.
- The disruptor **cannot draw**. Instead they get a panel of buttons, each of which temporarily messes with the **drawer's** pen. The server checks the role and the cooldown of every press, so a modified client can't skip them. The disruptor also **cannot guess** the word (their guesses never score and never leak to the chat) and doesn't keep the turn alive once all regular guessers have guessed.
- Every button has its own cooldown, which starts when it is pressed and grows with how severe the effect is:

  | Effect | Severity | Lasts | Cooldown |
  | --- | --- | --- | --- |
  | **Shaky hand** (random jitter on strokes) | mild | 6s | 15s |
  | **Random size** (brush size keeps changing) | mild | 8s | 15s |
  | **Random color** (brush color keeps changing) | mild | 8s | 15s |
  | **Laggy pen** (the line trails behind the cursor) | medium | 6s | 25s |
  | **Forced eraser** (the drawer's pen turns into the eraser) | medium | 4s | 30s |
  | **Flipped canvas** (strokes are mirrored or rotated by 90°) | severe | 5s | 40s |
  | **Undo** (removes the drawer's last stroke) | severe | instant | 45s |

  The numbers live in `internal/game/disruptor.go` and are easy to tune. The drawer gets their pen back when an effect runs out or the turn ends.
- The disruptor earns **5 points per effect** they successfully trigger, at most one reward per second and **at most 50 points per turn**.
- The disruptor's points don't influence the drawer's score.
- The disruptor is announced in chat and shown in the player list. Roles are re-rolled every turn and cleared when the game ends.

### Run locally

Install the Go version required by `go.mod` (Go 1.25.0), then run:

```sh
go test ./...
go run ./cmd/scribblers
```

Then open http://localhost:8080. Configuration is documented in the original configuration section above.
