Display code and stuff
======================

The tile layout is like this:

    0,0 1,0 1,2 1,3 ...
      0,1 1,1 2,1 ...
    0,2 1,2 2,2 3,2 ...
      0,3 1,3 2,3 ...
    ... ... ... ... ...

So if (x, y) is a tile then

NW -> x-1+y%2, y-1
NE -> x+y%2, y-1
SW -> x-1+y%2, y+1
SE -> x+y%2, y+1
