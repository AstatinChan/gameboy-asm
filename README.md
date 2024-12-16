# Astatin Assembler

Heyo !! This is the assembler I use to create [BunnyLand (Temporary name)](https://github.com/AstatinChan/BunnyLand-Gameboy) on [stream on Twitch](https://www.twitch.tv/astatinchan)

## Get the assembler

If you're on linux on x86\_64 you can download an already compiled binary [here](https://github.com/AstatinChan/gameboy-asm/releases/download/latest/gbasm_linux-x86_64).

If you can't use this binary you will need to compile the binary by yourself:

### Compiling

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

## Usage

To assemble gbasm files, call the assembler with the input file as the first parameter and the output file as the second

### Example

When having the file [`zoom.gbasm`](https://github.com/AstatinChan/gameboy-asm/blob/main/examples/zoom.gbasm) in your current directory, this will compile to a usable rom file:

```bash
gbasm zoom.gbasm zoom.rom
```

## Gameboy assembly

To even be able to start, gameboy roms need to contain some data to be validated by the boot rom. The minimal rom which starts, clear the screen and starts an infinite loop to hang is available in [examples/minimal.gbasm](https://github.com/AstatinChan/gameboy-asm/blob/main/examples/minimal.gbasm)

### Labels

Labels are defined with alphanumerical strings followed by a colon at the start of a line and can be referenced by appending the alphanumerical string to a `=`.

Inside of .MACRODEF, labels must start with `$` and cannot be referenced outside of the macro.

### Parameters

The list of the parameters that could be passed to the opcodes:

| Name | abbreviation | possible values | .DEFINE constants | Labels |
| ---- | ------------ | --------------- | ----------------- | ------ |
| 8 bits registers | r8 | `A`, `B`, `C`, `D`, `E`, `H`, `L`, `(HL)` | No | No |
| 16 bits registers | r16 | `BC`, `DE`, `HL`, `SP` (for PUSH and PULL `SP` is replaced by `AF`) | No | No |
| Raw 8 bits | 8b | `0x00` to `0xff`, can also be written `$00` to `$ff`. | Yes | No |
| Raw 16 bits | 16b | `0x0000` to `0xffff`, can also be written `$0000` to `$ffff` | Yes | Yes |
| register indirect | ri | `(BC)`, `(DE)`, `(HL+)`, `(HL-)` | No | No |
| Raw 8 indirect | 8i | `(0x00)` to `(0xff)`, can also be written `($00)` to `($ff)`. | Yes | No |
| Raw 16 indirect | 16i | `(0x0000)` to `(0xffff)`, can also be written `($0000)` to `($ffff)`. | Yes | No |
| Condition | cc | `NZ`, `Z`, `NC`, `C` | No | No |
| Bit ordinal | o | from `0` to `7` | No | No |

### Opcodes

The list of different possible opcodes:

| Opcode | First parameter | Second parameter |
| ------ | --------------- | ---------------- |
| **LD** | r8 | r8 |
| | r8 | 8b |
| | `A` | 8i |
| | 8i | `A` |
| | `A` | ri |
| | ri | `A` |
| | `A` | 16i |
| | 16i | `A` |
| | `A` | `(C)` |
| | `(C)` | `A` |
| | r16 | 16b |
| | 16i | `SP` |
| | `SP` | `HL` |
| **PUSH** | r16 | |
| **POP** | r16 | |
| **ADD** | r8 | |
| | 8b | |
| | `SP` | 8b |
| **ADC** | r8 | |
| | 8b | |
| **SUB** | r8 | |
| | 8b | |
| **SBC** | r8 | |
| | 8b | |
| **CP** | r8 | |
| | 8b | |
| **AND** | r8 | |
| | 8b | |
| **OR** | r8 | |
| | 8b | |
| **XOR** | r8 | |
| | 8b | |
| **INC** | r8 | |
| | r16 | |
| **DEC** | r8 | |
| | r16 | |
| **CCF** | | |
| **SCF** | | |
| **DAA** | | |
| **CPL** | | |
| **JP** | 16b | |
| | HL | |
| | cc | 16b |
| **JR** | 8b | |
| | cc | 8b |
| | 16b[^1] | |
| | cc | 16b[^1] |
| **CALL** | 16b | |
| | cc | 16b |
| **RET** | | |
| | cc | |
| **RETI** | | |
| **RST** | o | |
| **DI** | | |
| **EI** | | |
| **NOP** | | |
| **HALT** | | |
| **STOP** | | |
| **RLCA** | | |
| **RLA** | | |
| **RRCA** | | |
| **RRA** | | |
| **BIT** | o | r8 |
| **SET** | o | r8 |
| **RES** | o | r8 |
| **RLC** | r8 | |
| **RL** | r8 | |
| **RRC** | r8 | |
| **RR** | r8 | |
| **SLA** | r8 | |
| **SRA** | r8 | |
| **SWAP** | r8 | |
| **SRA** | r8 | |
| **SRL** | r8 | |
| **DBG**[^2] | | |

### MACROS

| Name | Parameters | Explanation | Usable in .MACRODEF |
| ---- | ---------- | ----------- | ------------------- |
| **.DB** | Any number of 8b | Will insert the 8b in the ROM as is | Yes |
| **.PADTO** | 16b | Will insert 0x00 in the ROM so that the next instruction is situated as the address provider | No |
| **.INCLUDE** | A file path in double quotes (example: `"./file-to-be-included.gbasm"`) | Will include all of the code inside the file provided in parameters | No |
| **.DEFINE** | A alphanumerical string as first parameter and a 8b, 16b, 8i or 16i to use as value | The alphanumerical string in parameter will be able to be used instead of the value | No |
| **.MACRODEF** | An alphanumeric string | Creates a new macro that will insert all of the code between this macro and the .END macro when called. The macro will be able to be called by calling the string provided in parameter prefixed by a `.` | No |
| **.END** | | Ends a .MACRODEF block | N/A |
| *User defined with .MACRODEF* | | | Yes |

[^1]: This is only syntaxic sugar that will be converted to 8b relative to the instruction to allow the use of labels. If the address is too far away from the address of the instruction in rom to be converted to 8b, the assembly will fail with an error suggesting to use JP instead of JR.
[^2]: This instruction is not standard and may cause error or crashes on both emulators and real hardware. In [my gameboy emulator](https://github.com/AstatinChan/gameboy-emulator) it is used to tell the emulator to dump the content of the registers.
