# Astatin Assembler

Heyo !! This is the assembler I use to create [BunnyLand (Temporary name)](https://github.com/AstatinChan/BunnyLand-Gameboy) on [stream on Twitch](https://www.twitch.tv/astatinchan)

## Download

If you're on linux on x86\_64 you can download an already compiled binary [here](https://github.com/AstatinChan/gameboy-asm/releases/download/latest/gbasm_linux-x86_64).

## How to compile

The assembler is made in golang, check if you have it on your system, If you do not:

```bash
# Arch
sudo pacman -S go

# Debian/Ubuntu
sudo apt install golang

# Windows:
# I don't know lol
```

Clone the assembler:

```bash
git clone https://github.com/AstatinChan/gameboy-asm.git
cd gameboy-asm
```

Compile the assembler:

```bash
go build .
```
