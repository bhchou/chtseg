chtseg
===

chtseg is a Chinese Segmation Processor built by Go language. It's based on my 1992
master degree proposal about Chinese segmetation by using constraint satification and 
statistics optimization. Though it depends on the collection of Chinese words and their
frequecies in the set of collections, it may be suitable for specific field like EC
to abstract key words. 

Currently it can be used in command line to segment a simple sentence input, or batch processing
by using --with-batchfile option and the out put will be in a file. The PHP/Java API or webAPI
will be developed. 

In the time this README being created, To process a sentence will take about 
0.03-0.04 sec on a Raspberry Pi4/4GB platform. The correction rate noted on the proposal
is 98.02% and recall rate is 93.19% respectively. 

## Preparing

Make sure you have a working Go environment.  Go version 1.2+ is supported.  [See
the install instructions for Go](http://golang.org/doc/install.html).

Then the following go application should be installed first by using `go get` command:

```
github.com/spf13/viper
github.com/urfave/cli
github.com/go-sql-driver/mysql
github.com/mattn/go-sqlite3
```
Note the `github.com/mattn/go-sqlite3` needs some other process to be installed correctly, please
see its document in github

A copy of mysql/mariaDB or sqlite3 db file is required for words references. 

Before using chtseg, a collection of Chinese words with its apperance count must be 
supported as a mysql table. Please refer to the files in `db` folder about the two necessary 
tables schema and there is a big sample collection of words abstracted in serval Taiwan EC websites.

If your mysql installation or sqlite3 db file is permenant, you can create `config.yml` file to meet your 
installation. Please see `config.yml.sample` for reference.

## Usage

Go to the directly you download chtseg and simply run:
```
$ go run chtseg 
```
If you do not have a correct config.yml, you may get an error to indicate how to 
set the mysql connection string temporarily. The `-m/--with-mysql` setting will 
overwrite the setting in config.yml. Surely you can omit `-m/--with-mysql` option
if there is a correct `config.yml` file.

## Example

```
$ go run chtseg.go 
MYSQL connected
Please drop a line to be segmented, or just enter to quit:
【NEG扭扭扭蛋】現貨 T-arts 暴牙動物 兔寶寶牙動物 兔寶寶牙齒 扭蛋 轉蛋 收藏 娛樂 全5種
NEG|扭|扭|扭蛋|現貨|T-arts|暴|牙|動物|兔|寶寶|牙|動物|兔|寶寶|牙齒|扭蛋|轉蛋|收藏|娛樂|全|5|種
---Found Keywords---
現貨[6.541704]
寶寶[5.656759]
牙齒[3.602060]
娛樂[2.397940]
扭蛋[7.348701]
牙[2.812913]
動物[5.510545]
兔[3.431364]
轉蛋[6.256778]
收藏[3.723456]
全[2.944483]
種[2.230449]
Please drop a line to be segmented, or just enter to quit:

bye
```
The keywords are found in DB with respectively score for futhur classification reference 


For batch processing:
```
$ go run chtseg.go --with-batchfile baichain.smname
```
the output will be baichain.sname.out

Note: it will be an error if the output file exists

Other useful option: -v will dump debug information

To see help:
```
$ go run chtseg.go --help
```
