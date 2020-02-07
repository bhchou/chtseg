pychtseg
===

The python API for chtseg is done by compiling GO file into C shared library like what java API does. Using ctype library in python will make it easy while calling golang. One should be noted is I tried to make string arguments in go format and get return value in C bytes format. It's easy to understand from sample code.

## Installation and test

Please run

```
make all
```
and it will produce `pychtseg.so` and `pychtseg.h` for you to call from java. Please take a look on `.h` file for the declaration of public functions.


```
python pychtseg.py
```

to see the sample output. Note the sample code is tested under python3.8, it should worked on all version of python3. There should be some modifications if you are writing python2 code.

## How to use

Please see `pychtseg.py` for example.


