# strimgo

![preview](https://i.imgur.com/b4XASwv.png)

Strimgo by default looks for a file called `.strimgo`, which is a list of streams separated by newlines, present in your home directory, unless you specify a file. Streams will only show up in the viewer if they are actually online and the width of the three columns (channel name, game name, stream title) is automatically adjusted to match the longest entry of all online streams. `livestreamer` needs to be visible in `$PATH`.

### Usage

strimgo [file]


    R               refresh stream view, check for online streams
    Q/Escape        quit
    Up/Down k/j     scroll up/down
    Left/Right
    h,l             scroll left/right
    Home/End        scroll to start/end
    Enter           run livestreamer with channel name and `medium,source` as arguments
    S/H/L/M/W/A     use source/high/low/mobile/worst/audio quality instead
    B               open stream page in default web browser
    C               open chat popout page in default web browser
    V               open stream popout page in default web browser
    Mouse:
    Left click      select stream
    MWheel Up/Down  scroll up/down


`--help` as an argument will display a help text

Keep in mind that there is no intention of rate-limiting your keyboard - strimgo doesn't discriminate your input and unassumingly runs 100 streams in source quality if you hold down shift+s. That is perfectly reasonable.

### Stream list

As mentioned above, the file is a list of channels, with one channel per line. Clear it of any spaces or tabs and make sure you write the channel names correctly.

###### Useful shell commands:

    sort -d file

(sort lines alphabetically)

    tail -1 file | sed -e "s/ \{1,\}$//"

(remove trailing spaces)

### Notes

This program's binaries have only been thoroughly tested on Linux, not on Mac or Windows. **Use this program at your own risk.**

The license may be found in the `LICENSE` file above.

Initially supplied binaries have been compiled using 1.7rc6 go tools due to improved code generation and then packed using upx to reduce the humongous binary size that go programs have due to static linking - upx has been known to cause issues with certain software.
