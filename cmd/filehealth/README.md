# Introduction

The `filehealth` command line tool scans files for invalid names, timestamps
and other attributes. It was written for use on Windows file systems but it
may be of use on other platforms in the future.

# Status

This tool has been used on production data sets with millions of files. Even
so, it is still considered experimental and testing should always be performed
on non-production data sets before using it in production settings.

# Alternatives

As originally described by `Craig Landis` in his 2008
[DFSR Does Not Replicate Temporary Files](https://techcommunity.microsoft.com/t5/ask-the-directory-services-team/dfsr-does-not-replicate-temporary-files/ba-p/395763)
blog post, `PowerShell` can be used to remove the temporary flag from files
by running their attribute flags through a bitwise `AND` that masks out the
`FILE_ATTRIBUTE_TEMPORARY` bit:

```
Get-childitem D:Data -recurse | ForEach-Object -process {if (($_.attributes -band 0x100) -eq 0x100) {$_.attributes = ($_.attributes -band 0xFEFF)}}
```

The `filehealth` tool has some advantages:

1. It can fix timestamps and file names, not just attributes.
2. It can preview and dry run changes before taking action.
3. It will _always_ show you what it intends to do, and will _always_
ask for confirmation before it does it.
4. It's immune to file path length issues. You don't have to remember to supply
`\\?\` as a [prefix to your file paths](https://learn.microsoft.com/en-us/archive/blogs/bclteam/long-paths-in-net-part-1-of-3-kim-hamilton).
5. It's arguably easier to use for those who aren't as familiar with
`PowerShell` and [bitmasks](https://en.wikipedia.org/wiki/Mask_(computing)).

# Installation

If your Windows system has [a supported version of Go](https://go.dev/dl/)
installed, you can build and install from source like so:

```
go install github.com/gentlemanautomaton/filehealth/cmd/filehealth@latest
```

Once you've done that, `filehealth.exe` should be available via your user's
`PATH` and usable in `CMD` or `PowerShell`.

# Caution

Please ***use extreme caution*** when using this tool. It is up to you to use
it responsibly. This tool can modify many files very quickly, and invoking it
on the wrong file set can be disastrous and extremely difficult or impossible
to undo.

As with any administrative tool, always make sure you're running the right
command on the right machine on the right data.

> TODO: Add support for a `--record` option that records actions taken during
> a `fix` operation to some sort of `undo file`, and allows them to be undone
> by a future `filehealth undo <undo-file>` command.

# Impact and Assumptions

Right now this program assumes three things about the files you run it on:

1. You don't want files to have the [Temporary Flag](https://learn.microsoft.com/en-us/windows/win32/fileio/file-attribute-constants#FILE_ATTRIBUTE_TEMPORARY) set
2. You don't want files to have nonsensical timestamps from the future
3. You don't want files to have leading or trailing whitespace in their names

A likely future improvement will make these opt-in, but right now the tool
assumes that all of these things are true.

# Usage

To non-destructively scan a set of files for issues, run the program with
the `scan` command:

```
filehealth.exe scan "C:\Example"
```

The `scan` and `fix` commands take optional `--include` and `--exclude`
options, which will filter the set of files they operate on via
[regular expressions](https://pkg.go.dev/regexp/syntax#hdr-Syntax):

```
filehealth.exe scan "C:\Example" --include "(\.dwg|\.pdf)$" --exclude "^tmp"
```

The `fix` command is like `scan`, but when issues are detected it will try to
fix them. Before fixing them, it asks for confirmation that it should proceed:

```
filehealth.exe fix "C:\Example"
```

To perform a dry run of `fix` without actually making changes, run the `fix`
command with `--dry`:

```
filehealth.exe fix "C:\Example" --dry
```

The default behavior of `fix` is to scan all of the files, prompt for
confirmation, and then fix them all at once. To break the job up into smaller
chunks, run the `fix` command with `--batch`, which will limit the number of
issues it encounters before prompting for a fix:

```
filehealth.exe fix "C:\Example" --batch 20
```

Note that this program follows the [GNU convention](https://www.gnu.org/software/libc/manual/html_node/Argument-Syntax.html)
of using '`--`' for long option names, unlike PowerShell and other programs
which use a single '`-`' character for all options.

## Examples

### Scanning a large set of files without any issues

```
filehealth.exe scan P:\Projects
----P:\Projects----
----0 skipped, 2363749 scanned, 2363749 healthy, 0 unhealthy, 0 issues (50.3977116s)----
```

### Scanning a small set of files with issues

```
filehealth.exe scan "C:\Example"

----C:\Example----
[0.0] unwanted attributes T: "Book Sales.txt": (fix: A,T → A)
[4.0] unwanted attributes T: "Old Files/  blah": (fix: A,T → A)
[4.1] leading or trailing space: "Old Files/  blah": (fix: "  blah" → "blah")
[5.0] leading or trailing space: "Old Files/ funny email.msg": (fix: " funny email.msg" → "funny email.msg")
[11.0] mod time: "Photos/SDC11024.JPG": (fix: 2050-07-27 22:54:12 PDT → 2022-09-26 23:07:26 PDT)
[12.0] mod time: "Photos/SDC11029.JPG": (fix: 2050-07-27 22:57:58 PDT → 2022-09-26 23:07:26 PDT)
[15.0] unwanted attributes T: "The Theory of Everything.txt": (fix: A,T → A)
----0 skipped, 16 scanned, 10 healthy, 6 unhealthy, 7 issues (9.4787ms)----
```

### Fixing a small set of files with issues

```
filehealth.exe fix "C:\Example"

----C:\Example----
[0.0] unwanted attributes T: "Book Sales.txt": (fix: A,T → A)
[4.0] unwanted attributes T: "Old Files/  blah": (fix: A,T → A)
[4.1] leading or trailing space: "Old Files/  blah": (fix: "  blah" → "blah")
[5.0] leading or trailing space: "Old Files/ funny email.msg": (fix: " funny email.msg" → "funny email.msg")
[11.0] mod time: "Photos/SDC11024.JPG": (fix: 2050-07-27 22:54:12 PDT → 2022-09-26 23:27:47 PDT)
[12.0] mod time: "Photos/SDC11029.JPG": (fix: 2050-07-27 22:57:58 PDT → 2022-09-26 23:27:47 PDT)
[15.0] unwanted attributes T: "The Theory of Everything.txt": (fix: A,T → A)
----
Proceed with fixes affecting 6 files? [yes/no]
yes
----
FIXED: [0.0] unwanted attributes T: "Book Sales.txt": attribute change: A,T → A
FIXED: [4.0] unwanted attributes T: "Old Files/  blah": attribute change: A,T → A
FIXED: [4.1] leading or trailing space: "Old Files/  blah": name change: "C:\Example\Old Files\  blah" → "C:\Example\Old Files\blah"
FIXED: [5.0] leading or trailing space: "Old Files/ funny email.msg": name change: "C:\Example\Old Files\ funny email.msg" → "C:\Example\Old Files\funny email.msg"
FIXED: [11.0] mod time: "Photos/SDC11024.JPG": mod time: 2050-07-27 22:54:12 PDT → 2022-09-26 23:27:49 PDT
FIXED: [12.0] mod time: "Photos/SDC11029.JPG": mod time: 2050-07-27 22:57:58 PDT → 2022-09-26 23:27:49 PDT
FIXED: [15.0] unwanted attributes T: "The Theory of Everything.txt": attribute change: A,T → A
----0 skipped, 16 scanned, 10 healthy, 6 unhealthy, 7 issues (10.5793ms)----
```

## Reference

Supplying `--help` to the `filehealth` program will cause it to print help
text abut its usage. Its output is included below for easy reference.

### Commands

```
Usage: filehealth.exe <command>

Scans the file system for files with health issues and optionally fixes them.

Flags:
  -h, --help    Show context-sensitive help.

Commands:
  scan <paths> ...
    Scans a set of file paths recursively for issues.

  fix <paths> ...
    Scans and optionally fixes files with issues.

Run "filehealth.exe <command> --help" for more information on a command.
```

### The `scan` Command

```
Usage: filehealth.exe scan <paths> ...

Scans a set of file paths recursively for issues.

Arguments:
  <paths> ...    Paths to search recursively ($PATHS).

Flags:
  -h, --help                   Show context-sensitive help.

      --include=INCLUDE,...    Include files matching regular expression pattern
                               ($INCLUDE).
      --exclude=EXCLUDE,...    Exclude files matching regular expression pattern
                               ($EXCLUDE).
      --skipped                Report on skipped files ($SHOW_SKIPPED).
      --healthy                Report on healthy files ($SHOW_HEALTHY).
```

### The `fix` Command

```
Usage: filehealth.exe fix <paths> ...

Scans and optionally fixes files with issues.

Arguments:
  <paths> ...    Paths to search recursively ($PATHS).

Flags:
  -h, --help                   Show context-sensitive help.

      --include=INCLUDE,...    Include files matching regular expression
                               patterns ($INCLUDE).
      --exclude=EXCLUDE,...    Exclude files matching regular expression
                               patterns ($EXCLUDE).
      --skipped                Report on skipped files ($SHOW_SKIPPED).
      --healthy                Report on healthy files ($SHOW_HEALTHY).
      --batch=INT              Maximum number of files to fix at a time
                               ($BATCH).
      --dry                    Perform a dry run without modifying files
                               ($DRYRUN).
```