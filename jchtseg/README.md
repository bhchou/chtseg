jchtseg
===

The java API for chtseg is done by compiling GO file into C shared library which is called by java [JNA](https://github.com/java-native-access/jna). There are web pages which may help besides JNA document:

* [Calling Go Functions from Other Languages](https://medium.com/learning-the-go-programming-language/calling-go-functions-from-other-languages-4c7d8bcc69bf)
* [freewind-demos](https://github.com/freewind-demos/call-go-function-from-java-demo)


## Installation and test

Please run

```
make all
```
and it will produce `jchtseg.so` and `jchtseg.h` for you to call from java. Please take a look on `.h` file for the declaration of public functions.

After JNA is installed and making it to your CLASSPATH, you can run

```
java jtest.java
```

to see the output.

## How to use

From `jtest.java`, there are two points showing how to link to shared library file.

Decalring the go function interface by extending Library class

```
public interface ChtsegLib extends Library {
    TwoStringsResult Getchtseg(String db, String conn, String teststr);
}
```

and set/create an instance refering to it from main class

```
private static String LIB_PATH = new File("jchtseg.so").getAbsolutePath();
static ChtsegLib INSTANCE = (ChtsegLib) Native.loadLibrary(LIB_PATH, ChtsegLib.class);

```

then call the function directly

```
TwoStringsResult segRet = INSTANCE.Getchtseg(db, conn, test);

```
Given three parameters for input:

* **db** : must be "sqlite3" or "mysql"
* **conn** : should be db file path or connection string *user:password@conntype(ip:port)/dbname* regarding to **db**
* **test** : string to be processed

The output format should meet the form denoted in `.h` file like the structure decalred in java file:

```
public class TwoStringsResult extends Structure implements Structure.ByValue {
    public String r0;
    public String r1;

    protected List<String> getFieldOrder() {
        return Arrays.asList("r0", "r1");
    }
}
```
where `r0` and `r1` are json output and error message respectively. You can compare the json output with the test string while running `jtest.java`.

## To do

Provide more sophisticated settings or functions.
