# term

[What's the difference between various $TERM variables?](https://unix.stackexchange.com/a/43951)
[a explanation of the codes](https://cedocs.intersystems.com/ens20152/csp/docbook/DocBook.UI.Page.cls?KEY=GVTT_termdef)

vt100 -> vt220 -> xterm

xterm is supposed to be a superset of vt220, in other words it's like vt220 but has more features.
For example, xterm usually supports colors, but vt220 doesn't. You can test this by pressing z inside top.

In the same way, vt220 has more features than vt100. For example, vt100 doesn't seem to support F11 and F12.

Compare their features and escape sequences that your system thinks they have by running infocmp <term type 1> <term type 2>, e.g. infocmp vt100 vt220.

The full list varies from system to system. You should be able to get the list using toe, toe /usr/share/terminfo, or find ${TERMINFO:-/usr/share/terminfo}. If none of those work, you could also look at ncurses' terminfo.src, which is where most distributions get the data from these days.

But unless your terminal looks like this or this, there's only a few others you might want to use:

- xterm-color - if you're on an older system and colors don't work
- putty, konsole, Eterm, rxvt, gnome, etc. - if you're running an XTerm emulator and some of the function keys, Backspace, Delete, Home, and End don't work properly
- screen - if running inside GNU screen (or tmux)
- linux - when logging in via a Linux console (e.g. Ctrl+Alt+F1)
- dumb - when everything is broken

The first one having colors was vt241. All the vt220 you can find are white, green or orange, depending on the phosphors(萤光粉) used.
