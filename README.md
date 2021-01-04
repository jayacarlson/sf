# sf
File / directory custom output generator that can be used to generate (bash) scripts

Usage of sf:  [args] [dir list]

Output can be set per directory and or per file with per file output as the default.

In addition, a starting output line and a tail output can be generated both at runtime and on a per 'Arg' [dir list] directory.

Various values are output using C syntax %metaChar tokens:

*  O   Origin path (PWD) from where **sf** was called, always homified
*  H   Users real HOME path: /home/<user>
*  r   [dir list] directory as given from [dir list]
*  R   Full dirpath of given [dir list] directory (homified by default)
*  p   Current dirpath, from [dir list] directory on down
*  P   Current full dirpath, from [dir list] dir on down (homified by default)
*  d   Latest directory name
*  D   Current dirpath, below [dir list] directory
*  s   The file/dir size
*  T   Current total file count (over all [dir list] dirs)
* --- files only output
*  f   Current filepath from [dir list] directory on down
*  F   Current full filepath (homified by default)
*  n   Current filename and extension as read: file.ext
*  N   Current filename without any extension: file
*  e   Current extension, without leading '.': ext
*  E   Current extension, including the '.':  .ext
*  c   Current file count inside the dir
*  C   Current file count inside [dir list] dir

In addition, certain values can be modified to upper-uase or lower-uase by prepending a 'u' or 'l' to the token: e.g. "%uf" to change the current filename to FILENAME.

File output can be filtered by file extension (include or exclusive) with files without an extension identified with - in the list: e.g. "txt - go"

===

Time for some examples...

Given a directory structure from HOME like this:

```
  d0
  |  file                    ! Note - No extension
  |  file.ext
  |- d1
     |  file.ex1
     |  File.Ex2
     |  file.Ext
     |- d2
        |  file              ! Note - No extension
        |  File.Ex1
        |  file.ex2
        |  file.ext
```

**>sf d0**   # will output

```
d0/file
d0/file.ext
```


**>sf -r d0**   # will output

```
d0/file
d0/file.ext
d0/d1/File.Ex2
d0/d1/file.Ext
d0/d1/file.ex1
d0/d1/d2/File.Ex1
d0/d1/d2/file
d0/d1/d2/file.ex2
d0/d1/d2/file.ext
```

**>sf -r -i "- ext" d0**   # will output

```
d0/file
d0/file.ext
d0/d1/d2/file
d0/d1/d2/file.ext
```

**>sf -r -I -i "- ext" d0**   # will output

```
d0/file
d0/file.ext
d0/d1/file.Ext
d0/d1/d2/file
d0/d1/d2/file.ext
```

**>sf -r -I -i "ex1 ex2" -b -d "# removing files from %p" -f "rm -f %f" d0**   # will output

```
#!/bin/bash
#
#       sf  -r -I -i "ex1 ex2" -b -d "# removing files from %p" -f "rm -f %f" d0 
#

# removing files from d0
# removing files from d0/d1
rm -f d0/d1/File.Ex2
rm -f d0/d1/file.ex1
# removing files from d0/d1/d2
rm -f d0/d1/d2/File.Ex1
rm -f d0/d1/d2/file.ex2
```

**>sf -r -I -i "ex1 ex2" -b -d "# removing files from %p" -f "rm -f %f" -o "do.sh" d0**

Would write the above to 'do.sh' as an executable file.

**>sf -r -d "%p" d0**   # will output

```
d0
d0/d1
d0/d1/d2
```

#### NOTE: Additional spaces added below to highlight the different dectory output for %r %R %p %P

**>sf -r -l "# Output for %R" -f "O:%O   r:%r   R:%R   d:%d   p:%p   P:%P   n:%n   e:%e   f:%f   F:%F" d0 d0/d1 d0/d1/d2**  

```
# Output for ~/d0
O:~   r:d0         R:~/d0         d:.    p:d0         P:~/d0/.         n:file       e:      f:d0/file             F:~/d0/file
O:~   r:d0         R:~/d0         d:.    p:d0         P:~/d0/.         n:file.ext   e:ext   f:d0/file.ext         F:~/d0/file.ext
O:~   r:d0         R:~/d0         d:d1   p:d0/d1      P:~/d0/d1        n:File.Ex2   e:Ex2   f:d0/d1/File.Ex2      F:~/d0/d1/File.Ex2
O:~   r:d0         R:~/d0         d:d1   p:d0/d1      P:~/d0/d1        n:file.Ext   e:Ext   f:d0/d1/file.Ext      F:~/d0/d1/file.Ext
O:~   r:d0         R:~/d0         d:d1   p:d0/d1      P:~/d0/d1        n:file.ex1   e:ex1   f:d0/d1/file.ex1      F:~/d0/d1/file.ex1
O:~   r:d0         R:~/d0         d:d2   p:d0/d1/d2   P:~/d0/d1/d2     n:File.Ex1   e:Ex1   f:d0/d1/d2/File.Ex1   F:~/d0/d1/d2/File.Ex1
O:~   r:d0         R:~/d0         d:d2   p:d0/d1/d2   P:~/d0/d1/d2     n:file       e:      f:d0/d1/d2/file       F:~/d0/d1/d2/file
O:~   r:d0         R:~/d0         d:d2   p:d0/d1/d2   P:~/d0/d1/d2     n:file.ex2   e:ex2   f:d0/d1/d2/file.ex2   F:~/d0/d1/d2/file.ex2
O:~   r:d0         R:~/d0         d:d2   p:d0/d1/d2   P:~/d0/d1/d2     n:file.ext   e:ext   f:d0/d1/d2/file.ext   F:~/d0/d1/d2/file.ext
# Output for ~/d0/d1
O:~   r:d0/d1      R:~/d0/d1      d:.    p:d0/d1      P:~/d0/d1/.      n:File.Ex2   e:Ex2   f:d0/d1/File.Ex2      F:~/d0/d1/File.Ex2
O:~   r:d0/d1      R:~/d0/d1      d:.    p:d0/d1      P:~/d0/d1/.      n:file.Ext   e:Ext   f:d0/d1/file.Ext      F:~/d0/d1/file.Ext
O:~   r:d0/d1      R:~/d0/d1      d:.    p:d0/d1      P:~/d0/d1/.      n:file.ex1   e:ex1   f:d0/d1/file.ex1      F:~/d0/d1/file.ex1
O:~   r:d0/d1      R:~/d0/d1      d:d2   p:d0/d1/d2   P:~/d0/d1/d2     n:File.Ex1   e:Ex1   f:d0/d1/d2/File.Ex1   F:~/d0/d1/d2/File.Ex1
O:~   r:d0/d1      R:~/d0/d1      d:d2   p:d0/d1/d2   P:~/d0/d1/d2     n:file       e:      f:d0/d1/d2/file       F:~/d0/d1/d2/file
O:~   r:d0/d1      R:~/d0/d1      d:d2   p:d0/d1/d2   P:~/d0/d1/d2     n:file.ex2   e:ex2   f:d0/d1/d2/file.ex2   F:~/d0/d1/d2/file.ex2
O:~   r:d0/d1      R:~/d0/d1      d:d2   p:d0/d1/d2   P:~/d0/d1/d2     n:file.ext   e:ext   f:d0/d1/d2/file.ext   F:~/d0/d1/d2/file.ext
# Output for ~/d0/d1/d2
O:~   r:d0/d1/d2   R:~/d0/d1/d2   d:.    p:d0/d1/d2   P:~/d0/d1/d2/.   n:File.Ex1   e:Ex1   f:d0/d1/d2/File.Ex1   F:~/d0/d1/d2/File.Ex1
O:~   r:d0/d1/d2   R:~/d0/d1/d2   d:.    p:d0/d1/d2   P:~/d0/d1/d2/.   n:file       e:      f:d0/d1/d2/file       F:~/d0/d1/d2/file
O:~   r:d0/d1/d2   R:~/d0/d1/d2   d:.    p:d0/d1/d2   P:~/d0/d1/d2/.   n:file.ex2   e:ex2   f:d0/d1/d2/file.ex2   F:~/d0/d1/d2/file.ex2
O:~   r:d0/d1/d2   R:~/d0/d1/d2   d:.    p:d0/d1/d2   P:~/d0/d1/d2/.   n:file.ext   e:ext   f:d0/d1/d2/file.ext   F:~/d0/d1/d2/file.ext
```