# catalauncher

This is a tool for helping you play [Cataclysm: Dark Days
Ahead](https://github.com/CleverRaven/Cataclysm-DDA/) (CDDA). CDDA is an open
source post-apocalyptic roguelike. It's a lot of fun, and this tool will make
it easier to play on Linux (and possible macOS and other Unix systems).

Please note that if you're on Windows there is a much more polished GUI
launcher, the [CDDA Game
Launcher](https://github.com/remyroy/CDDA-Game-Launcher).

## How to Use It

You can download a binary version of the latest build on this project's
[GitHub releases
page](https://github.com/houseabsolute/catalauncher/releases). Put this in a
directory in your path.

You'll need [Docker CE installed](https://docs.docker.com/install/) in order
to play the game since it's launched inside a Docker image. See the linked
docs for information on installing Docker CE.

## Setup

Before launching the game you need to set the launcher up. This is mostly to
tell it where to store files:

```
$> catalauncher setup
```

This will ask where you want to store game files (and in the future may ask
more questions). By default files are stored under
`$HOME/.catalauncher`. Accepting this default will make your life a little
simpler. Otherwise you'll need to tell it where your config file lives every
time you run it.

## Launching

To launch the game simply run the `launch` subcommand:

```
$> catalauncher launch
```

## Options

* `--config` - The location of your config file. This is accepted by all
  subcommands. Note that if your config file is not in the default location,
  `$HOME/.catalauncher`, then you'll need to pass this every time you run the
  `launch `subcommand.
* `--build` - This is an option for the launch subcommand. Pass this to
  specify which build you'd like to launch. By default you always get the most
  recent build.

## How It Works and What It Does

When you run `launch` there a number of things that happen.

### Fetching New Builds

First, the launcher checks for a new binary build in
http://dev.narc.ro/cataclysm/jenkins-latest/Linux_x64/Tiles/. These builds are
created via Jenkins.

If there is a new build it will downloaded and untarred (unless you asked for
an older build with the `--build` flag).

If the launcher is fetching a new build it will open the [Jenkins CDDA changes
list](http://gorgon.narc.ro:8080/job/Cataclysm-Matrix/changes) in your browser
so you can see what's new.

### Character Creation Templates

If your most recent local build has any character creation templates, those
are copied into the new build automatically (unfortunately these must live
under the game's directory, rather than in a shared location).

### Extras (Mods & Soundpacks)

The launcher also provides some mods and soundpacks. These are maintained in
the [houseabsolute/cataclysm-extras-collection
repo](https://github.com/houseabsolute/cataclysm-extras-collection). If you
want to add a mod, soundpack, or tileset, please [file a PR or issue
there](https://github.com/houseabsolute/cataclysm-extras-collection/issues). If
a mod is causing the game to crash, please [file a PR or issue to remove
it](https://github.com/houseabsolute/cataclysm-extras-collection/issues).

Whenever you run the launcher a local copy of that git repo is
pulled/updated. The contents are then copied into the per-build game directory
(unfortunately CDDA does not work when these directories are symlinked).

### Docker

The game itself is run in a Docker container using my
[houseabsolute/catalauncher-player](https://cloud.docker.com/u/houseabsolute/repository/docker/houseabsolute/catalauncher-player)
image. This image is built using the
[Dockerfile](https://github.com/houseabsolute/catalauncher/blob/master/docker/Dockerfile)
in this repo. This avoids the need to install any libraries on the host
system.

This container is run with quite a bit of access to the host system in order
to make video and sound work. I'm using Docker primarily for convenience
rather than isolation. Notably, docker is run with access to the following
files/directories/env vars on the host system:

* `/etc/machine-id`
* `/run/user/$USER_ID/pulse`
* `$HOME/.pulse`
* `/dev/dri`
* `/tmp/.X11-unix`
* `/var/lib/dbus`
* The `$DISPLAY` env var

However, The game will be executed using your user and group ids, not `root`.

Given this it's not clear to me whether this will work with Docker on macOS or
Windows (or even a Linux system that is very different from my own desktop
running Ubuntu 18.04). Patches to handle a greater variety of host systems are
welcome!

### Game Files

The launcher stores your game config, graveyard, and save files outside of the
per-build game directories, so they will persist as you run new builds.

## Todo and Known Issues

When your character dies (which will happen a lot) you'll see an error about
renaming some memorial files. I'm not sure why these are showing up. The
graveyard _does_ contain records of character deaths so it doesn't seem to be
a major error.

As mentioned above, the way Docker is run in order to provide it access to
video and sound on the host system greatly reduces container isolation and may
be very specific to my desktop.
