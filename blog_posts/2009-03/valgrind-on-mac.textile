Id: 16003
Title: Valgrind on mac
Tags: mac,programming
Date: 2009-03-27T21:04:36-07:00
Format: Markdown
--------------
As of Mar 27 2009 Valgrind doesn’t officially support mac but a port
that is good enough is in progress. You can learn more and monitor the
progress by reading [Nicholas Nethercote’s
blog](http://blog.mozilla.com/nnethercote/) (main person working on
Valgrind mac port).

This post is a summary of how I built Valgrind on Mac (OS X 10.5.6).

Since this involves compiling, you probably need developer tools. I have
both XCode and a healthy does of [MacPorts](http://www.macports.org/)
dev packages installed. I’m not sure how much you need, but I’m guessing
at least gcc, make, autotools.

<code>\
svn co svn://svn.valgrind.org/valgrind/trunk valgrind\
cd valgrind\
./autogen.sh\
./configure\
make\
sudo make install\
</code>

So far I’ve only tried Valgrind on mupdf and it worked, so this is very
promising. Valgrind is invaluable tool for C/C++ memory leak detection,
detecting memory mismanagement issues and profiling the code.

An unfortunate deficiency is lack of a GUI for inspecting callgrind
results (i.e. `kcachegrind` equivalent on mac). Maybe the cheapest way
to get that would be to convert callgrind results to shark format so
that shark tools can be used. Should be possible - I believe WebKit’s
JavaScript implementation saves profiling results in shark format, so
there is code that could be used as a template for that.
