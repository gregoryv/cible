# cible - a multi user game

In cible, players connect to a server and embark on adventures as
different characters. It's a terminal based multi user dungeon(MUD)
game, where you control your character with the keyboard.

## Quick start

    $ go install github.com/gregoryv/cible/cmd/cible@latest
	$ cible -h
    Usage: cible [OPTIONS]
    
    Options
        -b, --bind : ":8089"
        -d, --debug
        -s, --server
        -h, --help


To play a local game start a server first

    $ cible -s

then in another terminal run the client

    $ USER=majorPain cible
	

## Download

Binaries are available at https://www.7de.se/dl
